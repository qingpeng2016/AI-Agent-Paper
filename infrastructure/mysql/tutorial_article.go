package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type TutorialArticleImpl struct {
	db *gorm.DB
}

func NewTutorialArticleImpl(db *gorm.DB) repository.TutorialArticleRepo {
	return &TutorialArticleImpl{db: db}
}

func (r *TutorialArticleImpl) ListPublished(ctx context.Context, categoryCode string) ([]entity.TutorialArticle, error) {
	q := r.db.WithContext(ctx).
		Table("tutorial_article AS a").
		Select("a.*").
		Where("a.status = ?", "published")
	if categoryCode != "" {
		q = q.Joins("INNER JOIN tutorial_category c ON c.id = a.category_id AND c.status = ?", "active").
			Where("c.code = ?", categoryCode)
	}
	var rows []entity.TutorialArticle
	err := q.Order("a.published_at DESC, a.sort ASC, a.id DESC").Find(&rows).Error
	return rows, err
}

func (r *TutorialArticleImpl) FindPublishedBySlug(ctx context.Context, slug string) (*entity.TutorialArticle, error) {
	var row entity.TutorialArticle
	err := r.db.WithContext(ctx).
		Where("slug = ? AND status = ?", slug, "published").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
