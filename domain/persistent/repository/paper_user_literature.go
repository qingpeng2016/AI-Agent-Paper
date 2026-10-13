package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperUserLiteratureRepo interface {
	ListActiveByUser(ctx context.Context, userID uint64) ([]entity.PaperUserLiterature, error)
	ListActiveByUserPaged(ctx context.Context, userID uint64, offset, limit int) ([]entity.PaperUserLiterature, error)
	CountActiveByUser(ctx context.Context, userID uint64) (int64, error)
	FindActiveByUserStorageURI(ctx context.Context, userID uint64, storageURI string) (*entity.PaperUserLiterature, error)
	Create(ctx context.Context, row *entity.PaperUserLiterature) error
}
