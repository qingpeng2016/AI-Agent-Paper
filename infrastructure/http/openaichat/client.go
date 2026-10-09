package openaichat

import (
	"context"

	httpentity "github.com/qingpeng2016/ai-agent-paper/domain/http/entity"
	httprepo "github.com/qingpeng2016/ai-agent-paper/domain/http/repository"
	httpx "github.com/qingpeng2016/ai-agent-paper/infrastructure/http"
)

// Client 实现 domain/http/repository.OpenAIChatRepo
type Client struct {
	http *httpx.Client
}

func NewClient(http *httpx.Client) httprepo.OpenAIChatRepo {
	return &Client{http: http}
}

func (c *Client) ChatCompletions(ctx context.Context, req httpentity.OpenAIChatCompletionRequest) (httpentity.OpenAIChatCompletionResult, error) {
	return DoChatCompletions(ctx, c.http, req)
}
