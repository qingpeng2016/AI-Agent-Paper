package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
)

type TutorialCategoryRepo interface {
	ListActive(ctx context.Context) ([]entity.TutorialCategory, error)
	GetByID(ctx context.Context, id uint) (*entity.TutorialCategory, error)
}
