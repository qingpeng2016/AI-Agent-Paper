package repository

import (
	"context"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperManuscriptSubmissionRepo interface {
	Create(ctx context.Context, row *entity.PaperManuscriptSubmission) error
	GetByIDForUser(ctx context.Context, id, manuscriptID uint64, userID uint) (*entity.PaperManuscriptSubmission, error)
	ListByManuscriptPaged(ctx context.Context, manuscriptID uint64, offset, limit int) ([]entity.PaperManuscriptSubmission, error)
	CountByManuscript(ctx context.Context, manuscriptID uint64) (int64, error)
	UpdateRevision(ctx context.Context, id, manuscriptID uint64, userID uint, status string, decision, comments *string, revisionDeadline *time.Time) (bool, error)
	MarkResubmitted(ctx context.Context, id, manuscriptID uint64, userID uint, newRound int, status string) (bool, error)
	UpdateMainPDF(ctx context.Context, id, manuscriptID uint64, userID uint, uri, filename string) (bool, error)
	GetLatestForManuscript(ctx context.Context, manuscriptID uint64) (*entity.PaperManuscriptSubmission, error)
	InsertEvent(ctx context.Context, ev *entity.PaperManuscriptSubmissionEvent) error
	ListEvents(ctx context.Context, submissionID uint64) ([]entity.PaperManuscriptSubmissionEvent, error)
}
