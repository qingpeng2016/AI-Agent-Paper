package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperRefLiteratureSourceRepo interface {
	FindActiveByCode(ctx context.Context, code string) (*entity.PaperRefLiteratureSource, error)
	ListActiveOrdered(ctx context.Context) ([]entity.PaperRefLiteratureSource, error)
}
