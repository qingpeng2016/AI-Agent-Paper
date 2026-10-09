package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperOutputLiteratureReviewRepo interface {
	CreateAsCurrent(ctx context.Context, row *entity.PaperOutputLiteratureReview) error
}
