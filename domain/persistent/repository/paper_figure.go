package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperFigureRepo interface {
	ListActiveByManuscript(ctx context.Context, manuscriptID uint64) ([]entity.PaperFigure, error)
	Create(ctx context.Context, row *entity.PaperFigure) error
	ListActiveByUserPaged(ctx context.Context, userID uint64, filter FigureListFilter, offset, limit int) ([]FigureListRow, error)
	CountActiveByUser(ctx context.Context, userID uint64, filter FigureListFilter) (int64, error)
	ListActiveCategoryBindingsByFigureIDs(
		ctx context.Context,
		userID uint64,
		figureIDs []uint64,
	) ([]FigureCategoryBinding, error)
	ListActiveFigureIDsByCategory(ctx context.Context, userID, categoryID uint64) ([]uint64, error)
	SyncCategoryFigureMembership(
		ctx context.Context,
		userID, categoryID uint64,
		figureIDs []uint64,
	) error
}
