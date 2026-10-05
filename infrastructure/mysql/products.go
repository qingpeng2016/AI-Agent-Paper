package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type ProductsImpl struct {
	db *gorm.DB
}

func NewProductsImpl(db *gorm.DB) repository.ProductsRepo {
	return &ProductsImpl{db: db}
}

func (r *ProductsImpl) ListOnSale(ctx context.Context, categoryID uint) ([]entity.Products, error) {
	q := r.db.WithContext(ctx).Where("status = ?", "on_sale")
	if categoryID > 0 {
		q = q.Where("products_category_id = ?", categoryID)
	}
	var rows []entity.Products
	err := q.Order("sort_order ASC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *ProductsImpl) FindOnSaleByID(ctx context.Context, id uint) (*entity.Products, error) {
	var row entity.Products
	err := r.db.WithContext(ctx).
		Where("id = ? AND status = ?", id, "on_sale").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ProductsImpl) FindOnSaleBySKUCode(ctx context.Context, skuCode string) (*entity.Products, error) {
	var row entity.Products
	err := r.db.WithContext(ctx).
		Where("sku_code = ? AND status = ?", skuCode, "on_sale").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ProductsImpl) FindByID(ctx context.Context, tx *gorm.DB, id uint) (*entity.Products, error) {
	var row entity.Products
	err := repository.GormDB(ctx, r.db, tx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
