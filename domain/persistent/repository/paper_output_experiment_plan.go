package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperOutputExperimentPlanRepo interface {
	Create(ctx context.Context, row *entity.PaperOutputExperimentPlan) error
}
