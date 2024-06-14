package wx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/open4go/log"
	"io"
	"net/http"
	"os"
)

// GetWxaCodeUnLimit 获取不受限制的小程序码
// https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/qrcode-link/qr-code/getUnlimitedQRCode.html
func GetWxaCodeUnLimit(ctx context.Context, accessToken string, scene string, page string, envVer string) ([]byte, error) {
	urlStr := fmt.Sprintf(getwxacodeunlimitUrl, accessToken)

	reqParam := &GetWxaCodeUnLimitParam{
		Page:       page, // 默认是主页
		Scene:      scene,
		CheckPath:  true,
		EnvVersion: envVer,
	}
	reqByte, _ := json.Marshal(reqParam)
	wxResp, err := http.Post(urlStr, "application/json", bytes.NewBuffer(reqByte))
	if err != nil {
		return nil, err
	}
	defer wxResp.Body.Close()

	//wxRespByte, err := ioutil.ReadAll(wxResp.Body)
	//
	//if wxResp.StatusCode != http.StatusOK {
	//	return nil, fmt.Errorf("http status code: %d", wxResp.StatusCode)
	//}

	//// 创建一个新文件，用于存储下载的图片
	outFile, err := os.Create("output.jpg")
	if err != nil {
		log.Log(ctx).Fatal("Error creating file: ", err)
	}
	defer outFile.Close()

	// 将二进制流写入新文件
	_, err = io.Copy(outFile, wxResp.Body)
	if err != nil {
		log.Log(ctx).Fatal("Error writing file: ", err)
	}

	return []byte(""), nil
}
