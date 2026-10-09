package google

import (
	"context"

	httpentity "github.com/qingpeng2016/ai-agent-paper/domain/http/entity"
	httprepo "github.com/qingpeng2016/ai-agent-paper/domain/http/repository"
	httpx "github.com/qingpeng2016/ai-agent-paper/infrastructure/http"
	"github.com/qingpeng2016/ai-agent-paper/infrastructure/http/openaichat"
)

// Client 实现 domain/http/repository.GoogleGeminiRepo（Gemini OpenAI 兼容端点）。
type Client struct {
	http *httpx.Client
}

func NewClient(http *httpx.Client) httprepo.GoogleGeminiRepo {
	return &Client{http: http}
}

func (c *Client) ChatCompletions(ctx context.Context, req httpentity.OpenAIChatCompletionRequest) (httpentity.OpenAIChatCompletionResult, error) {
	return openaichat.DoChatCompletions(ctx, c.http, req)
}
