package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperUserLiteratureRepo interface {
	ListActiveByUser(ctx context.Context, userID uint64) ([]entity.PaperUserLiterature, error)
	FindActiveByUserStorageURL(ctx context.Context, userID uint64, storageURL string) (*entity.PaperUserLiterature, error)
	Create(ctx context.Context, row *entity.PaperUserLiterature) error
}
