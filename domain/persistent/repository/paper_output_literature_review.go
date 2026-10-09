package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperOutputLiteratureReviewRepo interface {
	CreateAsCurrent(ctx context.Context, row *entity.PaperOutputLiteratureReview) error
	ListByManuscriptForUser(ctx context.Context, userID uint, manuscriptID uint64) ([]entity.PaperOutputLiteratureReview, error)
	// ListByManuscript 按论文查全部版本（调用方须已校验 manuscript 归属）。
	ListByManuscript(ctx context.Context, manuscriptID uint64) ([]entity.PaperOutputLiteratureReview, error)
}
