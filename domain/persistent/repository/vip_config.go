package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
)

type VipConfigRepo interface {
	FindByID(ctx context.Context, id uint) (*entity.VipConfig, error)
	ListEnabled(ctx context.Context) ([]entity.VipConfig, error)
	ListAll(ctx context.Context) ([]entity.VipConfig, error)
	FindDefault(ctx context.Context) (*entity.VipConfig, error)
}
