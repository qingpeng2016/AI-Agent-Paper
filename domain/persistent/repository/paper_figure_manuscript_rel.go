package repository

import "context"

type PaperFigureManuscriptRelRepo interface {
	ListActiveManuscriptBindingsByFigureIDs(
		ctx context.Context,
		userID uint,
		figureIDs []uint64,
	) ([]FigureManuscriptBinding, error)
	ListActiveFigureBindingsByManuscriptIDs(
		ctx context.Context,
		userID uint,
		manuscriptIDs []uint64,
	) ([]ManuscriptFigureBinding, error)
	ListActiveFigureIDsByManuscript(ctx context.Context, userID uint, manuscriptID uint64) ([]uint64, error)
	SyncManuscriptFigureMembership(
		ctx context.Context,
		userID uint,
		manuscriptID uint64,
		figureIDs []uint64,
	) error
}
