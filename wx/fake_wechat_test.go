package wx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// 生产代码把微信地址写死为 https://api.weixin.qq.com。
// 测试期间替换 http.DefaultTransport：发往 api.weixin.qq.com 的请求全部转给本地 httptest 假服务，
// 其他任何外部主机直接报错，保证测试绝不会访问真实的微信接口。
type fakeWeChatTransport struct {
	target *url.URL
	next   http.RoundTripper
}

func (f *fakeWeChatTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.URL.Host != "api.weixin.qq.com" {
		return nil, errors.New("测试中禁止访问外部主机: " + r.URL.Host)
	}
	r2 := r.Clone(r.Context())
	r2.URL.Scheme = f.target.Scheme
	r2.URL.Host = f.target.Host
	r2.Host = f.target.Host
	return f.next.RoundTrip(r2)
}

// fakeWeChat 启动本地假微信服务，并在测试结束时恢复 http.DefaultTransport。
func fakeWeChat(t *testing.T, h http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(h)
	u, _ := url.Parse(srv.URL)
	prev := http.DefaultTransport
	http.DefaultTransport = &fakeWeChatTransport{target: u, next: prev}
	t.Cleanup(func() {
		http.DefaultTransport = prev
		srv.Close()
	})
	return srv
}

// 假的凭据，只会发给本地假服务。
const (
	fakeAppID  = "wx_fake_appid"
	fakeSecret = "fake-secret-for-test"
)

func TestFakeTransportBlocksOtherHosts(t *testing.T) {
	fakeWeChat(t, func(w http.ResponseWriter, _ *http.Request) {})
	cases := []struct {
		name, url string
		wantErr   bool
	}{
		{"微信域名转到本地假服务", "https://api.weixin.qq.com/ping", false},
		{"其他外部主机被拒绝", "https://example.com/", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(tc.url)
			if resp != nil {
				_ = resp.Body.Close()
			}
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v", err)
			}
		})
	}
}
