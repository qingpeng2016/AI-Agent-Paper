package response

import (
	"errors"
	"net/http"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/common/notification"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ApiResp 统一 HTTP JSON 响应包装。
type ApiResp struct {
	Code    uint32      `json:"code"`    // 业务码，200 表示成功
	Data    interface{} `json:"data"`    // 业务数据，失败时可为 null
	Message string      `json:"message"` // 提示信息
}

const (
	OkBizCode uint32 = 200
)

func localeFromGin(c *gin.Context) string {
	raw := strings.TrimSpace(c.GetHeader("X-Locale"))
	if raw == "" {
		raw = strings.TrimSpace(c.Query("locale"))
	}
	return errorx.NormalizeLocale(raw)
}

// ResponseErr 统一错误响应处理（按 X-Locale / ?locale= 返回对应语言 message）
func ResponseErr(c *gin.Context, err error) {
	if errorx.ShouldAlertHTTPError(err) {
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		notification.SendErrorLog(c.Request.Context(), "http-api-error",
			zap.String("path", path),
			zap.String("method", c.Request.Method),
			zap.Error(err),
		)
	}
	locale := localeFromGin(c)
	var bizErr *errorx.RespErr
	if errors.As(err, &bizErr) {
		Response(c, http.StatusOK, uint32(bizErr.Code), bizErr.Data, bizErr.LocalizedMessage(locale))
		return
	}

	Response(c, http.StatusOK, uint32(errorx.ErrUnknown.Code), nil, errorx.ErrUnknown.LocalizedMessage(locale))
}

// ResponseSuccess 成功响应
func ResponseSuccess(c *gin.Context, data interface{}) {
	Response(c, http.StatusOK, OkBizCode, data, "success")
}

// ResponseSuccessWithMsg 成功响应（自定义消息）
func ResponseSuccessWithMsg(c *gin.Context, data interface{}, msg string) {
	Response(c, http.StatusOK, OkBizCode, data, msg)
}

// Response 统一响应方法
func Response(c *gin.Context, httpCode, errCode uint32, data interface{}, message string) {
	c.JSON(int(httpCode), gin.H{
		"code":    errCode,
		"message": message,
		"data":    data,
	})
	c.Abort()
}
