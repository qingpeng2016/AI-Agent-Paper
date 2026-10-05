package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

type UserAPIKeysRepo interface {
	Create(ctx context.Context, tx *gorm.DB, m *entity.UserAPIKeys) error
	UpdateWhere(ctx context.Context, tx *gorm.DB, where map[string]interface{}, updateData map[string]interface{}) error
	FindByID(ctx context.Context, id uint) (*entity.UserAPIKeys, error)
	// ListMainByUserID：user_id 归属的 Key（main + 分配给自己的 sub）。
	ListMainByUserID(ctx context.Context, userID uint) ([]entity.UserAPIKeys, error)
	ListSubByOwnerUserID(ctx context.Context, ownerUserID uint) ([]entity.UserAPIKeys, error)
	SumActiveLimitTokensBySubscription(ctx context.Context, subscriptionID uint, excludeKeyID uint) (int64, error)
	SumActiveSubKeyLimitTokensBySubscription(ctx context.Context, subscriptionID uint, excludeKeyID uint) (int64, error)
	ExistsActiveSubKey(ctx context.Context, subscriptionID, memberUserID uint) (bool, error)
}
