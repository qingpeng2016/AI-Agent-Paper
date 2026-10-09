package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperOutputTopicStepRepo interface {
	BeginRun(ctx context.Context, userID uint64, inputParams []byte) (runVersion int, err error)
	ListByUserRun(ctx context.Context, userID uint64, runVersion int) ([]entity.PaperOutputTopicStep, error)
	GetLatestRunByUser(ctx context.Context, userID uint64) (runVersion int, steps []entity.PaperOutputTopicStep, err error)
	GetStep(ctx context.Context, userID uint64, runVersion int, stageCode string) (*entity.PaperOutputTopicStep, error)
	SaveStep(ctx context.Context, step *entity.PaperOutputTopicStep) error
	CancelRun(ctx context.Context, userID uint64, runVersion int) error
	BindManuscript(ctx context.Context, userID uint64, runVersion int, manuscriptID uint64) error
}
