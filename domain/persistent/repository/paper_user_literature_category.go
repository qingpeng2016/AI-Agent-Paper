package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperUserLiteratureCategoryRepo interface {
	ListActivePaged(ctx context.Context, offset, limit int) ([]entity.PaperUserLiteratureCategory, error)
	ExistsActiveByName(ctx context.Context, name string) (bool, error)
	Create(ctx context.Context, row *entity.PaperUserLiteratureCategory) error
	CountActive(ctx context.Context) (int64, error)
	FindActiveByID(ctx context.Context, id uint64) (*entity.PaperUserLiteratureCategory, error)
	ExistsActiveByNameExceptID(ctx context.Context, name string, exceptID uint64) (bool, error)
	UpdateActiveName(ctx context.Context, id uint64, name string) error
}
