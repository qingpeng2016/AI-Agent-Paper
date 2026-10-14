package errorx

import (
	"errors"
	"strings"
)

// ShouldAlertHTTPError API 返回错误是否应推送 Lark（SendErrorLog）。
// 纯用户参数/步骤状态不报警；带 detail 的参数错误与未知/服务端错误报警。
func ShouldAlertHTTPError(err error) bool {
	if err == nil {
		return false
	}
	var biz *RespErr
	if !errors.As(err, &biz) {
		return true
	}
	switch biz.Code {
	case ErrParamsError.Code:
		return respErrHasDetail(biz)
	case ErrTopicStepInvalid.Code,
		ErrTopicRunNotFound.Code,
		ErrManuscriptNotFound.Code,
		ErrUserNotFound.Code,
		ErrWrongPassword.Code,
		ErrUserDisabled.Code,
		ErrPasswordMismatch.Code,
		ErrInvalidCredential.Code:
		return respErrHasDetail(biz)
	default:
		return true
	}
}

func respErrHasDetail(biz *RespErr) bool {
	if biz == nil {
		return false
	}
	if strings.Contains(biz.Msg, "：") {
		return true
	}
	if strings.Contains(biz.EnMsg, ": ") {
		return true
	}
	return false
}
