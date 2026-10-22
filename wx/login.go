package wx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// 开发文档: https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/login.html
const (
	grantType = "authorization_code"
)

var wxLoginEndpoint = "https://api.weixin.qq.com/sns/jscode2session"

// LoginResp 登陆
type LoginResp struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// GetSession 获取微信登陆返回
func GetSession(ctx context.Context, code string, secretKey string, appId string) (*LoginResp, error) {
	// 拼接请求地址（参数转义；地址中带有 secret，任何错误/日志都不能包含它）
	q := url.Values{}
	q.Set("appid", appId)
	q.Set("secret", secretKey)
	q.Set("js_code", code)
	q.Set("grant_type", grantType)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wxLoginEndpoint+"?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("jscode2session: build request failed")
	}
	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jscode2session request failed: %w", stripURL(err))
	}
	defer func() { _ = httpResp.Body.Close() }()
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jscode2session: status %d", httpResp.StatusCode)
	}
	b, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	resp := &LoginResp{}
	if err := json.Unmarshal(b, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
