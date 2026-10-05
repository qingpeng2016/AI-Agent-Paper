package invite_rebate

import (
	"encoding/base64"
	"strings"

	"github.com/qingpeng2016/ai-token-mall/common/errorx"
)

const payoutQRMaxBytes = 2 << 20 // 2MB

var payoutQRAllowedMimes = map[string]struct{}{
	"image/png":  {},
	"image/jpeg": {},
	"image/webp": {},
}

func validatePayoutQRImage(image []byte, mime string) error {
	if len(image) == 0 {
		return errorx.ErrParamsError
	}
	if len(image) > payoutQRMaxBytes {
		return errorx.ErrParamsError
	}
	mime = strings.ToLower(strings.TrimSpace(mime))
	if _, ok := payoutQRAllowedMimes[mime]; !ok {
		return errorx.ErrParamsError
	}
	return nil
}

func qrImageDataURL(mime *string, image []byte) string {
	if mime == nil || len(image) == 0 {
		return ""
	}
	m := strings.TrimSpace(*mime)
	if m == "" {
		return ""
	}
	return "data:" + m + ";base64," + base64.StdEncoding.EncodeToString(image)
}
