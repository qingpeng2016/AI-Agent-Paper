package http

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/qingpeng2016/ai-agent-paper/conf"
)

// LiteratureOutboundMinInterval 文献检索 API 与 PDF 下载等对外 HTTP 的最小间隔。
const LiteratureOutboundMinInterval = 2 * time.Second

// literatureHTTPRequestTimeout 文献第三方 API 常比 LLM 接口更慢，单独放宽（仍不低于 http_timeout_sec）。
const literatureHTTPRequestTimeout = 90 * time.Second

const literatureHTTPMaxAttempts = 4

var (
	literatureReqMu     sync.Mutex
	literatureReqLastAt time.Time
)

type literatureRetryHTTP struct {
	status int
}

func (e literatureRetryHTTP) Error() string {
	return http.StatusText(e.status)
}

func sleepContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// WithLiteratureOutbound 全局串行：距上次文献相关 HTTP 结束至少间隔 LiteratureOutboundMinInterval 后再执行 fn。
func WithLiteratureOutbound(ctx context.Context, fn func() error) error {
	literatureReqMu.Lock()
	defer literatureReqMu.Unlock()

	if !literatureReqLastAt.IsZero() {
		wait := LiteratureOutboundMinInterval - time.Since(literatureReqLastAt)
		if wait > 0 {
			if err := sleepContext(ctx, wait); err != nil {
				return err
			}
		}
	}

	err := fn()
	literatureReqLastAt = time.Now()
	return err
}

func literatureRetryableStatus(code int) bool {
	return code == http.StatusTooManyRequests || code == http.StatusServiceUnavailable
}

func literatureRetryableTransport(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return false
}

func (c *Client) getForLiterature(ctx context.Context, url string, query map[string]string, headers map[string]string) (*resty.Response, error) {
	rc := c.restyForContext(ctx)
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		rc = resty.New()
		timeout := conf.GetHTTPTimeout()
		if timeout < literatureHTTPRequestTimeout {
			timeout = literatureHTTPRequestTimeout
		}
		rc.SetTimeout(timeout)
	}
	req := rc.R().SetContext(ctx)
	if query != nil {
		req.SetQueryParams(query)
	}
	for k, v := range headers {
		req.SetHeader(k, v)
	}
	req.SetHeader("User-Agent", "AI-Agent-Paper/1.0")
	return req.Get(url)
}

func literatureShouldRetry(lastErr error, attempt int) bool {
	if attempt >= literatureHTTPMaxAttempts-1 {
		return false
	}
	var retry literatureRetryHTTP
	if errors.As(lastErr, &retry) {
		return true
	}
	return literatureRetryableTransport(lastErr)
}

// GetLiterature 文献第三方 GET：串行 + 最小间隔；429/503 与超时类错误退避重试。
func (c *Client) GetLiterature(ctx context.Context, url string, query map[string]string, headers map[string]string) (*resty.Response, error) {
	var resp *resty.Response
	var lastErr error
	for attempt := 0; attempt < literatureHTTPMaxAttempts; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * 5 * time.Second
			if err := sleepContext(ctx, backoff); err != nil {
				return resp, err
			}
		}
		lastErr = WithLiteratureOutbound(ctx, func() error {
			var e error
			resp, e = c.getForLiterature(ctx, url, query, headers)
			if e != nil {
				return e
			}
			if resp != nil && literatureRetryableStatus(resp.StatusCode()) {
				return literatureRetryHTTP{status: resp.StatusCode()}
			}
			return nil
		})
		if lastErr == nil {
			return resp, nil
		}
		if !literatureShouldRetry(lastErr, attempt) {
			break
		}
	}
	if lastErr != nil {
		var retry literatureRetryHTTP
		if errors.As(lastErr, &retry) && resp != nil {
			return resp, nil
		}
	}
	return resp, lastErr
}
