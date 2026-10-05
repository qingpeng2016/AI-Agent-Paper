package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
)

type EnterpriseProductsRepo interface {
	ListActive(ctx context.Context) ([]entity.EnterpriseProducts, error)
	FindActiveByCode(ctx context.Context, code string) (*entity.EnterpriseProducts, error)
}
