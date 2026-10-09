package http

import (
	"context"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/trace"
	"github.com/qingpeng2016/ai-agent-paper/conf"
)

type Client struct {
	rc *resty.Client
}

func NewHTTPClient() *Client {
	c := resty.New()
	c.SetTimeout(conf.GetHTTPTimeout())
	return &Client{rc: c}
}

// restyForContext：若 ctx 带 deadline（如 LLM 的 timeout_ms），HTTP 超时与之一致；否则用全局 http_timeout_sec。
func (c *Client) restyForContext(ctx context.Context) *resty.Client {
	if deadline, ok := ctx.Deadline(); ok {
		d := time.Until(deadline)
		if d <= 0 {
			d = time.Millisecond
		}
		rc := resty.New()
		rc.SetTimeout(d)
		return rc
	}
	return c.rc
}

func (c *Client) PostForm(ctx context.Context, url string, form map[string]string) (*resty.Response, error) {
	req := c.restyForContext(ctx).R().SetContext(ctx)
	if form != nil {
		req.SetFormData(form)
	}
	req.SetHeader(trace.HeaderTraceID, trace.GetTraceIdByCtx(ctx))
	return req.Post(url)
}

func (c *Client) PostJSON(ctx context.Context, url string, body any, headers map[string]string) (*resty.Response, error) {
	req := c.restyForContext(ctx).R().SetContext(ctx).SetBody(body)
	for k, v := range headers {
		req.SetHeader(k, v)
	}
	req.SetHeader(trace.HeaderTraceID, trace.GetTraceIdByCtx(ctx))
	req.SetHeader("Content-Type", "application/json")
	return req.Post(url)
}

func (c *Client) Get(ctx context.Context, url string, query map[string]string, headers map[string]string) (*resty.Response, error) {
	req := c.restyForContext(ctx).R().SetContext(ctx)
	if query != nil {
		req.SetQueryParams(query)
	}
	for k, v := range headers {
		req.SetHeader(k, v)
	}
	req.SetHeader(trace.HeaderTraceID, trace.GetTraceIdByCtx(ctx))
	return req.Get(url)
}
