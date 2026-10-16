package http

import (
	"context"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

const literatureMinRequestInterval = time.Second

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

// GetLiterature 文献第三方 GET；全局串行，任意两次请求结束到下一次开始至少间隔 literatureMinRequestInterval。
func (c *Client) GetLiterature(ctx context.Context, url string, query map[string]string, headers map[string]string) (*resty.Response, error) {
	literatureReqMu.Lock()
	defer literatureReqMu.Unlock()

	if !literatureReqLastAt.IsZero() {
		wait := literatureMinRequestInterval - time.Since(literatureReqLastAt)
		if wait > 0 {
			if err := sleepContext(ctx, wait); err != nil {
				return nil, err
			}
		}
	}

	resp, err := c.Get(ctx, url, query, headers)
	literatureReqLastAt = time.Now()
	return resp, err
}
