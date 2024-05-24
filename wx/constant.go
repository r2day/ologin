package wx

const (
	getwxacodeunlimitUrl = "https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=%s"
	// 获取微信小程序的AccessToken
	wxaAccessTokenKey    = "wxa_access_token_%s"
	getWxaAccessTokenUrl = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
)
