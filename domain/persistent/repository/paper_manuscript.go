package repository

import (
	"context"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperManuscriptRepo interface {
	GetByIDForUser(ctx context.Context, manuscriptID, userID uint) (*entity.PaperManuscript, error)
	ListByUser(ctx context.Context, userID uint) ([]entity.PaperManuscript, error)
	ListByUserPaged(ctx context.Context, userID uint, offset, limit int) ([]entity.PaperManuscript, error)
	CountByUserActive(ctx context.Context, userID uint) (int64, error)
	Create(ctx context.Context, userID uint, title string, disciplineID *uint64, language string) (*entity.PaperManuscript, error)
	SetCurrentForUser(ctx context.Context, userID, manuscriptID uint) error
	UpdateSubmissionSummary(
		ctx context.Context,
		manuscriptID, userID uint,
		submissionStatus string,
		submissionLabel *string,
		submissionDeadline *time.Time,
	) error
}
