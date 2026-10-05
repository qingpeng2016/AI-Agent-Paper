package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
)

type VipDomainConfigRepo interface {
	FindByID(ctx context.Context, id uint) (*entity.VipDomainConfig, error)
	FindByDomain(ctx context.Context, domain string) (*entity.VipDomainConfig, error)
	ListAll(ctx context.Context) ([]entity.VipDomainConfig, error)
	ListNonOfficial(ctx context.Context) ([]entity.VipDomainConfig, error)
}
