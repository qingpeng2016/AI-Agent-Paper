package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperManuscriptRepo interface {
	GetByIDForUser(ctx context.Context, manuscriptID, userID uint) (*entity.PaperManuscript, error)
	Create(ctx context.Context, userID uint, title string) (*entity.PaperManuscript, error)
}
