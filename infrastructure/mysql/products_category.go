package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type ProductsCategoryImpl struct {
	db *gorm.DB
}

func NewProductsCategoryImpl(db *gorm.DB) repository.ProductsCategoryRepo {
	return &ProductsCategoryImpl{db: db}
}

func (r *ProductsCategoryImpl) ListActive(ctx context.Context) ([]entity.ProductsCategory, error) {
	var rows []entity.ProductsCategory
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort ASC, id ASC").
		Find(&rows).Error
	return rows, err
}
