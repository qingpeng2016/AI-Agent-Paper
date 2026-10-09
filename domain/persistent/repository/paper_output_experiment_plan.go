package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperOutputExperimentPlanRepo interface {
	Create(ctx context.Context, row *entity.PaperOutputExperimentPlan) error
	ListByManuscript(ctx context.Context, manuscriptID uint64) ([]entity.PaperOutputExperimentPlan, error)
	GetByIDForUser(
		ctx context.Context,
		id, manuscriptID uint64,
		userID uint,
	) (*entity.PaperOutputExperimentPlan, error)
	// GetByIDForManuscript 调用方须已校验 manuscript 归属。
	GetByIDForManuscript(ctx context.Context, id, manuscriptID uint64) (*entity.PaperOutputExperimentPlan, error)
}
