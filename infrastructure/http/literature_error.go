package http

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/qingpeng2016/ai-agent-paper/common/notification"
	"go.uber.org/zap"
)

const literatureResponsePreviewMax = 2048

// LiteratureHTTPError 文献第三方 HTTP 失败详情（可 errors.As 取出写入 step.extra）。
type LiteratureHTTPError struct {
	Platform        string
	Operation       string
	RequestURL      string
	HTTPStatus      int
	ResponsePreview string
	Cause           error
}

func (e *LiteratureHTTPError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	if e.HTTPStatus > 0 {
		return fmt.Sprintf("%s %s http %d", e.Platform, e.Operation, e.HTTPStatus)
	}
	return fmt.Sprintf("%s %s failed", e.Platform, e.Operation)
}

func (e *LiteratureHTTPError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func TruncateLiteratureBody(body []byte, max int) string {
	if max <= 0 {
		max = literatureResponsePreviewMax
	}
	s := strings.TrimSpace(string(body))
	if len(s) <= max {
		return s
	}
	return s[:max] + "…(truncated)"
}

// LogLiteratureTransportError 网络层失败（无 HTTP 响应体）。
func LogLiteratureTransportError(ctx context.Context, platform, operation, requestURL string, err error) error {
	if err == nil {
		return nil
	}
	notification.SendErrorLog(ctx, "literature-http-failure",
		zap.String("platform", platform),
		zap.String("operation", operation),
		zap.String("request_url", requestURL),
		zap.Error(err))
	return err
}

// NewLiteratureHTTPError 记录响应摘要并返回可 As 的结构化错误。
func NewLiteratureHTTPError(ctx context.Context, platform, operation, requestURL string, resp *resty.Response, cause error) error {
	if cause == nil {
		cause = fmt.Errorf("%s %s failed", platform, operation)
	}
	preview := ""
	status := 0
	if resp != nil {
		status = resp.StatusCode()
		preview = TruncateLiteratureBody(resp.Body(), literatureResponsePreviewMax)
	}
	notification.SendErrorLog(ctx, "literature-http-failure",
		zap.String("platform", platform),
		zap.String("operation", operation),
		zap.String("request_url", requestURL),
		zap.Int("http_status", status),
		zap.String("response_preview", preview),
		zap.Error(cause))
	return &LiteratureHTTPError{
		Platform:        platform,
		Operation:       operation,
		RequestURL:      requestURL,
		HTTPStatus:      status,
		ResponsePreview: preview,
		Cause:           cause,
	}
}
