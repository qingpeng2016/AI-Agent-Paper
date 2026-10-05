package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
)

type TutorialArticleRepo interface {
	ListPublished(ctx context.Context, categoryCode string) ([]entity.TutorialArticle, error)
	FindPublishedBySlug(ctx context.Context, slug string) (*entity.TutorialArticle, error)
}
