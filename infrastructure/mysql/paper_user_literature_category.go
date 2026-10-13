package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

type PaperUserLiteratureCategoryImpl struct {
	db *gorm.DB
}

func NewPaperUserLiteratureCategoryImpl(db *gorm.DB) repository.PaperUserLiteratureCategoryRepo {
	return &PaperUserLiteratureCategoryImpl{db: db}
}

func (r *PaperUserLiteratureCategoryImpl) ListActivePaged(ctx context.Context, offset, limit int) ([]entity.PaperUserLiteratureCategory, error) {
	var rows []entity.PaperUserLiteratureCategory
	q := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort ASC, id ASC")
	if limit > 0 {
		q = q.Offset(offset).Limit(limit)
	}
	err := q.Find(&rows).Error
	return rows, err
}

func (r *PaperUserLiteratureCategoryImpl) CountActive(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&entity.PaperUserLiteratureCategory{}).
		Where("status = ?", "active").
		Count(&n).Error
	return n, err
}
