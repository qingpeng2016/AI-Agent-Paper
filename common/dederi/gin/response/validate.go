package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/qingpeng2016/ai-token-mall/common/errorx"
)

// ResponseBindErr JSON 绑定 / validator 错误（避免落进「未知错误」）
func ResponseBindErr(c *gin.Context, err error) {
	locale := localeFromGin(c)
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		Response(c, http.StatusOK, uint32(errorx.ErrParamsError.Code), nil, formatValidationErrors(ve, locale))
		return
	}
	Response(c, http.StatusOK, uint32(errorx.ErrParamsError.Code), nil, errorx.ErrParamsError.LocalizedMessage(locale))
}

func formatValidationErrors(ve validator.ValidationErrors, locale string) string {
	for _, fe := range ve {
		switch fe.Field() {
		case "Password":
			switch fe.Tag() {
			case "min":
				return "密码至少 6 位"
			case "required":
				return "请填写密码"
			}
		case "ConfirmPassword":
			switch fe.Tag() {
			case "min":
				return "确认密码至少 6 位"
			case "required":
				return "请填写确认密码"
			}
		}
	}
	return errorx.ErrParamsError.LocalizedMessage(locale)
}
