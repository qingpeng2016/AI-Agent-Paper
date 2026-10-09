package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/http/entity"
)

// OpenAIChatRepo OpenAI 及 openai_compatible / gateway 等 Chat Completions，由 infrastructure/http/openaichat 实现。
type OpenAIChatRepo interface {
	ChatCompletions(ctx context.Context, req entity.OpenAIChatCompletionRequest) (entity.OpenAIChatCompletionResult, error)
}
