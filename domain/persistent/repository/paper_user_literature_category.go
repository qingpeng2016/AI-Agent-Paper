package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperUserLiteratureCategoryRepo interface {
	ListActivePaged(ctx context.Context, offset, limit int) ([]entity.PaperUserLiteratureCategory, error)
	CountActive(ctx context.Context) (int64, error)
}
