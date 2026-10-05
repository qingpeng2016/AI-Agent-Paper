package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type TutorialCategoryImpl struct {
	db *gorm.DB
}

func NewTutorialCategoryImpl(db *gorm.DB) repository.TutorialCategoryRepo {
	return &TutorialCategoryImpl{db: db}
}

func (r *TutorialCategoryImpl) ListActive(ctx context.Context) ([]entity.TutorialCategory, error) {
	var rows []entity.TutorialCategory
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *TutorialCategoryImpl) GetByID(ctx context.Context, id uint) (*entity.TutorialCategory, error) {
	var row entity.TutorialCategory
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
