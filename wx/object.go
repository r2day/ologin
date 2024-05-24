package wx

type GetWxaCodeUnLimitParam struct {
	Page       string `json:"page"`
	Scene      string `json:"scene"`
	CheckPath  bool   `json:"check_path"`
	EnvVersion string `json:"env_version"`
}

type GetWxaCodeUnLimitResp struct {
	Buffer  []byte `json:"buffer"` // 图片二进制流
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type GetWxaAccessTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`

	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}
