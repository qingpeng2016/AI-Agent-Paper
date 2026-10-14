package mysql

import (
	"context"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

type PaperFigureCategoryImpl struct {
	db *gorm.DB
}

func NewPaperFigureCategoryImpl(db *gorm.DB) repository.PaperFigureCategoryRepo {
	return &PaperFigureCategoryImpl{db: db}
}

func (r *PaperFigureCategoryImpl) ListActivePaged(ctx context.Context, offset, limit int) ([]entity.PaperFigureCategory, error) {
	var rows []entity.PaperFigureCategory
	q := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("id ASC")
	if limit > 0 {
		q = q.Offset(offset).Limit(limit)
	}
	err := q.Find(&rows).Error
	return rows, err
}

func (r *PaperFigureCategoryImpl) CountActive(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&entity.PaperFigureCategory{}).
		Where("status = ?", "active").
		Count(&n).Error
	return n, err
}

func (r *PaperFigureCategoryImpl) ExistsActiveByName(ctx context.Context, name string) (bool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return false, nil
	}
	var n int64
	err := r.db.WithContext(ctx).
		Model(&entity.PaperFigureCategory{}).
		Where("status = ? AND name = ?", "active", name).
		Count(&n).Error
	return n > 0, err
}

func (r *PaperFigureCategoryImpl) Create(ctx context.Context, row *entity.PaperFigureCategory) error {
	if row == nil {
		return nil
	}
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *PaperFigureCategoryImpl) FindActiveByID(ctx context.Context, id uint64) (*entity.PaperFigureCategory, error) {
	if id == 0 {
		return nil, nil
	}
	var row entity.PaperFigureCategory
	err := r.db.WithContext(ctx).
		Where("id = ? AND status = ?", id, "active").
		First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperFigureCategoryImpl) ExistsActiveByNameExceptID(
	ctx context.Context,
	name string,
	exceptID uint64,
) (bool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return false, nil
	}
	var n int64
	err := r.db.WithContext(ctx).
		Model(&entity.PaperFigureCategory{}).
		Where("status = ? AND name = ? AND id <> ?", "active", name, exceptID).
		Count(&n).Error
	return n > 0, err
}

func (r *PaperFigureCategoryImpl) UpdateActiveName(ctx context.Context, id uint64, name string) error {
	if id == 0 {
		return nil
	}
	name = strings.TrimSpace(name)
	return r.db.WithContext(ctx).
		Model(&entity.PaperFigureCategory{}).
		Where("id = ? AND status = ?", id, "active").
		Update("name", name).Error
}
