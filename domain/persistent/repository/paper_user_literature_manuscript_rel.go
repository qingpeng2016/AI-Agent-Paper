package repository

import "context"

type PaperUserLiteratureManuscriptRelRepo interface {
	ListActiveManuscriptBindingsByLiteratureIDs(
		ctx context.Context,
		userID uint,
		literatureIDs []uint64,
	) ([]LiteratureManuscriptBinding, error)
	ListActiveLiteratureBindingsByManuscriptIDs(
		ctx context.Context,
		userID uint,
		manuscriptIDs []uint64,
	) ([]ManuscriptLiteratureBinding, error)
	ListActiveLiteratureIDsByManuscript(ctx context.Context, userID uint, manuscriptID uint64) ([]uint64, error)
	SyncManuscriptLiteratureMembership(
		ctx context.Context,
		userID uint,
		manuscriptID uint64,
		literatureIDs []uint64,
	) error
}
