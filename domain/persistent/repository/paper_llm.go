package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperLLMRepo interface {
	GetActiveModelByID(ctx context.Context, id uint) (*entity.PaperLLMModelConfig, error)
	// ResolveActiveModel：configID>0 时用指定配置；否则取最新 active 端点（优先 anthropic）。
	ResolveActiveModel(ctx context.Context, configID uint) (*entity.PaperLLMModelConfig, error)
	InsertCallLog(ctx context.Context, row *entity.PaperLLMCallLog) error
	// AddTokensUsed 分别累加 use_in_tokens / use_out_tokens。
	AddTokensUsed(ctx context.Context, modelConfigID uint, promptDelta, completionDelta int) error
}
