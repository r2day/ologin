package wx

const (
	getwxacodeunlimitUrl = "https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=%s"
	// 获取微信小程序的AccessToken
	wxaAccessTokenKey = "wxa_access_token_%s"
)

// 微信接口地址（变量便于测试替换）
var (
	wxaAccessTokenEndpoint = "https://api.weixin.qq.com/cgi-bin/token"
)
