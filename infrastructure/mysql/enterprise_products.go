package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type EnterpriseProductsImpl struct {
	db *gorm.DB
}

func NewEnterpriseProductsImpl(db *gorm.DB) repository.EnterpriseProductsRepo {
	return &EnterpriseProductsImpl{db: db}
}

func (r *EnterpriseProductsImpl) ListActive(ctx context.Context) ([]entity.EnterpriseProducts, error) {
	var rows []entity.EnterpriseProducts
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *EnterpriseProductsImpl) FindActiveByCode(ctx context.Context, code string) (*entity.EnterpriseProducts, error) {
	var row entity.EnterpriseProducts
	err := r.db.WithContext(ctx).
		Where("code = ? AND status = ?", code, "active").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
