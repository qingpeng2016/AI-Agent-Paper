package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
)

type ProductsCategoryRepo interface {
	ListActive(ctx context.Context) ([]entity.ProductsCategory, error)
}
