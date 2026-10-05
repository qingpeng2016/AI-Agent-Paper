package http

import (
	"context"

	"github.com/go-resty/resty/v2"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/trace"
	"github.com/qingpeng2016/ai-token-mall/conf"
)

type Client struct {
	rc *resty.Client
}

func NewHTTPClient() *Client {
	c := resty.New()
	c.SetTimeout(conf.GetHTTPTimeout())
	return &Client{rc: c}
}

func (c *Client) PostForm(ctx context.Context, url string, form map[string]string) (*resty.Response, error) {
	req := c.rc.R().SetContext(ctx)
	if form != nil {
		req.SetFormData(form)
	}
	req.SetHeader(trace.HeaderTraceID, trace.GetTraceIdByCtx(ctx))
	return req.Post(url)
}
