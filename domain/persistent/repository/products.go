package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

type ProductsRepo interface {
	ListOnSale(ctx context.Context, categoryID uint) ([]entity.Products, error)
	FindOnSaleBySKUCode(ctx context.Context, skuCode string) (*entity.Products, error)
	FindOnSaleByID(ctx context.Context, id uint) (*entity.Products, error)
	FindByID(ctx context.Context, tx *gorm.DB, id uint) (*entity.Products, error)
}
