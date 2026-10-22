package wx

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestGetWxaAccessToken(t *testing.T) {
	cases := []struct {
		name      string
		status    int
		body      string
		wantErr   string
		wantToken string
		wantHits  int32
	}{
		{"成功", 200, `{"access_token":"tok-1","expires_in":7200}`, "", "tok-1", 1},
		{"微信返回 errcode，重试 3 次后失败", 200, `{"errcode":40013,"errmsg":"invalid appid"}`, "获取微信小程序AccessToken失败", "", 3},
		{"非 JSON，重试 3 次后失败", 200, `<html>`, "获取微信小程序AccessToken失败", "", 3},
		// 现状：不检查 HTTP 状态码，只要响应体是合法 JSON 就当作成功
		{"HTTP 500 但响应体合法（现状：视为成功）", 500, `{"access_token":"tok-500","expires_in":7200}`, "", "tok-500", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var hits int32
			fakeWeChat(t, func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&hits, 1)
				q := r.URL.Query()
				if r.Method != http.MethodGet || r.URL.Path != "/cgi-bin/token" ||
					q.Get("grant_type") != "client_credential" || q.Get("appid") != fakeAppID || q.Get("secret") != fakeSecret {
					t.Errorf("请求不符合预期: %s %s", r.Method, r.URL)
				}
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			})
			resp, err := GetWxaAccessToken(context.Background(), fakeAppID, fakeSecret)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) || resp != nil {
					t.Fatalf("resp=%v err=%v", resp, err)
				}
			} else if err != nil || resp.AccessToken != tc.wantToken || resp.ExpiresIn != 7200 {
				t.Fatalf("resp=%+v err=%v", resp, err)
			}
			if got := atomic.LoadInt32(&hits); got != tc.wantHits {
				t.Fatalf("请求次数=%d want %d", got, tc.wantHits)
			}
		})
	}
}

func TestGetWxaAccessTokenUnreachable(t *testing.T) {
	srv := fakeWeChat(t, func(w http.ResponseWriter, _ *http.Request) {})
	srv.Close() // 假服务关闭后连接失败
	if _, err := getWxaAccessToken(context.Background(), fakeAppID, fakeSecret); err == nil {
		t.Fatal("连接失败应返回错误")
	}
	if _, err := GetWxaAccessToken(context.Background(), fakeAppID, fakeSecret); err == nil {
		t.Fatal("重试后仍应失败")
	}
}

func TestGetSession(t *testing.T) {
	cases := []struct {
		name    string
		code    string
		status  int
		body    string
		wantErr bool
		check   func(t *testing.T, r *LoginResp, q map[string][]string)
	}{
		{"成功", "code-1", 200, `{"openid":"o1","session_key":"sk","unionid":"u1"}`, false,
			func(t *testing.T, r *LoginResp, q map[string][]string) {
				if r.OpenID != "o1" || r.SessionKey != "sk" || r.UnionID != "u1" {
					t.Fatalf("resp=%+v", r)
				}
				if q["js_code"][0] != "code-1" || q["grant_type"][0] != "authorization_code" || q["appid"][0] != fakeAppID || q["secret"][0] != fakeSecret {
					t.Fatalf("query=%v", q)
				}
			}},
		// 现状：微信返回 errcode 时不报错，调用方必须自己检查 ErrCode
		{"errcode 原样返回（现状：err 为 nil）", "bad", 200, `{"errcode":40029,"errmsg":"invalid code"}`, false,
			func(t *testing.T, r *LoginResp, _ map[string][]string) {
				if r.ErrCode != 40029 || r.OpenID != "" {
					t.Fatalf("resp=%+v", r)
				}
			}},
		{"HTTP 非 200 报错", "c", 502, `{}`, true, nil},
		{"非 JSON 报错", "c", 200, `oops`, true, nil},
		// 现状：code 未做 URL 转义，可以注入额外的查询参数
		{"code 未转义可注入参数（现状）", "abc&appid=evil&x=1", 200, `{"openid":"o2"}`, false,
			func(t *testing.T, _ *LoginResp, q map[string][]string) {
				if len(q["appid"]) != 2 || q["appid"][1] != "evil" || q["js_code"][0] != "abc" || q["x"][0] != "1" {
					t.Fatalf("query=%v", q)
				}
			}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var query map[string][]string
			fakeWeChat(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/sns/jscode2session" {
					t.Errorf("path=%s", r.URL.Path)
				}
				query = r.URL.Query()
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			})
			resp, err := GetSession(context.Background(), tc.code, fakeSecret, fakeAppID)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v", err)
			}
			if tc.check != nil {
				tc.check(t, resp, query)
			}
		})
	}
}

// GetWxaCodeUnLimit 会把响应写到当前目录下固定的 output.jpg，这里切换到临时目录执行。
func TestGetWxaCodeUnLimit(t *testing.T) {
	cases := []struct {
		name, body, envVer string
	}{
		{"图片写入 output.jpg，返回值为空", "\x89PNG-fake-image", "release"},
		// 现状：微信返回错误 JSON 时也写入 output.jpg，且不返回错误
		{"错误响应也被当作图片写入（现状）", `{"errcode":40001,"errmsg":"invalid credential"}`, "trial"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)
			fakeWeChat(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/wxa/getwxacodeunlimit" || r.URL.Query().Get("access_token") != "fake-token" {
					t.Errorf("请求不符合预期: %s %s", r.Method, r.URL)
				}
				var p GetWxaCodeUnLimitParam
				if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
					t.Errorf("body: %v", err)
				}
				if p.Scene != "table=1" || p.Page != "pages/index" || !p.CheckPath || p.EnvVersion != tc.envVer {
					t.Errorf("param=%+v", p)
				}
				_, _ = io.WriteString(w, tc.body)
			})
			got, err := GetWxaCodeUnLimit(context.Background(), "fake-token", "table=1", "pages/index", tc.envVer)
			if err != nil || len(got) != 0 {
				t.Fatalf("got=%q err=%v", got, err)
			}
			b, err := os.ReadFile(filepath.Join(dir, "output.jpg"))
			if err != nil || string(b) != tc.body {
				t.Fatalf("output.jpg=%q err=%v", b, err)
			}
		})
	}
}

func TestGetWxaCodeUnLimitUnreachable(t *testing.T) {
	t.Chdir(t.TempDir())
	srv := fakeWeChat(t, func(w http.ResponseWriter, _ *http.Request) {})
	srv.Close()
	if _, err := GetWxaCodeUnLimit(context.Background(), "fake-token", "s", "p", "develop"); err == nil {
		t.Fatal("连接失败应返回错误")
	}
	if _, err := os.Stat("output.jpg"); !os.IsNotExist(err) {
		t.Fatalf("连接失败时不应创建文件: %v", err)
	}
}
