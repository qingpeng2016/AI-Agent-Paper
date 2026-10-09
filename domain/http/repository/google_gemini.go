package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/http/entity"
)

// GoogleGeminiRepo Gemini OpenAI 兼容 Chat Completions，由 infrastructure/http/google 实现。
type GoogleGeminiRepo interface {
	ChatCompletions(ctx context.Context, req entity.OpenAIChatCompletionRequest) (entity.OpenAIChatCompletionResult, error)
}
