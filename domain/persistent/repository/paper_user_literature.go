package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperUserLiteratureRepo interface {
	ListActiveByManuscript(ctx context.Context, manuscriptID uint64) ([]entity.PaperUserLiterature, error)
	Create(ctx context.Context, row *entity.PaperUserLiterature) error
}
