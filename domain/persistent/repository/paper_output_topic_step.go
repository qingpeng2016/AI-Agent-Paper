package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperOutputTopicStepRepo interface {
	BeginRun(ctx context.Context, manuscriptID, userID uint64, inputParams []byte) (runVersion int, err error)
	ListByRun(ctx context.Context, manuscriptID uint64, runVersion int) ([]entity.PaperOutputTopicStep, error)
	GetCurrentRun(ctx context.Context, manuscriptID uint64) (runVersion int, steps []entity.PaperOutputTopicStep, err error)
	GetStep(ctx context.Context, manuscriptID uint64, runVersion int, stageCode string) (*entity.PaperOutputTopicStep, error)
	SaveStep(ctx context.Context, step *entity.PaperOutputTopicStep) error
	CancelCurrentRun(ctx context.Context, manuscriptID uint64) error
}
