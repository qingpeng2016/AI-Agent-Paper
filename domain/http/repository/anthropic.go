package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/http/entity"
)

// AnthropicRepo Claude Messages API，由 infrastructure/http/anthropic 实现。
type AnthropicRepo interface {
	CreateMessage(ctx context.Context, req entity.AnthropicMessageRequest) (entity.AnthropicMessageResult, error)
}
