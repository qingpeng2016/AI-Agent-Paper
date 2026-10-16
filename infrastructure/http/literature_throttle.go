package http

import (
	"context"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

// LiteratureOutboundMinInterval 文献检索 API 与 PDF 下载等对外 HTTP 的最小间隔。
const LiteratureOutboundMinInterval = 2 * time.Second

var (
	literatureReqMu     sync.Mutex
	literatureReqLastAt time.Time
)

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

// GetLiterature 文献第三方 GET（走 WithLiteratureOutbound）。
func (c *Client) GetLiterature(ctx context.Context, url string, query map[string]string, headers map[string]string) (*resty.Response, error) {
	var resp *resty.Response
	err := WithLiteratureOutbound(ctx, func() error {
		var e error
		resp, e = c.Get(ctx, url, query, headers)
		return e
	})
	return resp, err
}
