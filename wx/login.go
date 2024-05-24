package wx

import (
	"github.com/open4go/util9s"
)

// 开发文档: https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/login.html
const (
	wxLoginEndpoint = "https://api.weixin.qq.com/sns/jscode2session?appid="
	grantType       = "authorization_code"
)

// LoginResp 登陆
type LoginResp struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// GetSession 获取微信登陆返回
func GetSession(code string, secretKey string, appId string) (*LoginResp, error) {
	// 拼接请求地址
	url := wxLoginEndpoint + appId + "&secret=" +
		secretKey + "&js_code=" +
		code + "&grant_type=" + grantType
	resp := &LoginResp{}
	err := util9s.Get(url, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
