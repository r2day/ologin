package wx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testSecret = "TEST-APP-SECRET-should-not-leak"

func TestGetSessionErrorDoesNotContainSecret(t *testing.T) {
	prev := wxLoginEndpoint
	t.Cleanup(func() { wxLoginEndpoint = prev })
	wxLoginEndpoint = "http://127.0.0.1:1/sns/jscode2session" // 连接失败
	_, err := GetSession(context.Background(), "code", testSecret, "appid")
	if err == nil || strings.Contains(err.Error(), testSecret) {
		t.Fatalf("err must not contain secret: %v", err)
	}
}

func TestGetSessionEscapesParams(t *testing.T) {
	var gotQuery map[string][]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"openid":"o1","session_key":"k"}`))
	}))
	defer srv.Close()
	prev := wxLoginEndpoint
	t.Cleanup(func() { wxLoginEndpoint = prev })
	wxLoginEndpoint = srv.URL + "/sns/jscode2session"

	rsp, err := GetSession(context.Background(), "abc&appid=evil", testSecret, "appid")
	if err != nil || rsp.OpenID != "o1" {
		t.Fatalf("rsp=%+v err=%v", rsp, err)
	}
	if len(gotQuery["appid"]) != 1 || gotQuery["appid"][0] != "appid" || gotQuery["js_code"][0] != "abc&appid=evil" {
		t.Fatalf("query=%v", gotQuery)
	}
}

func TestGetWxaAccessTokenErrorDoesNotContainSecret(t *testing.T) {
	prev := wxaAccessTokenEndpoint
	t.Cleanup(func() { wxaAccessTokenEndpoint = prev })
	wxaAccessTokenEndpoint = "http://127.0.0.1:1/cgi-bin/token"
	_, err := getWxaAccessToken(context.Background(), "appid", testSecret)
	if err == nil || strings.Contains(err.Error(), testSecret) {
		t.Fatalf("err must not contain secret: %v", err)
	}
}
