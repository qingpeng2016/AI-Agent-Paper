package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperLLMRepo interface {
	GetActiveModelByID(ctx context.Context, id uint) (*entity.PaperLLMModelConfig, error)
	GetActiveBindingByStage(ctx context.Context, stageCode string) (*entity.PaperLLMWorkflowBinding, error)
	GetActivePromptByStage(ctx context.Context, stageCode string) (*entity.PaperLLMPromptTemplate, error)
	InsertCallLog(ctx context.Context, row *entity.PaperLLMCallLog) error
}
