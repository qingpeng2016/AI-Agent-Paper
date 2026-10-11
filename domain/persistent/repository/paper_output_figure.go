package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperOutputFigureRepo interface {
	ListActiveByManuscript(ctx context.Context, manuscriptID uint64) ([]entity.PaperOutputFigure, error)
	Create(ctx context.Context, row *entity.PaperOutputFigure) error
}
