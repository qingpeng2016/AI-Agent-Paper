package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

type UserSubscriptionsListRow struct {
	entity.UserSubscriptions
	ProductCardTitle string `gorm:"column:product_card_title"`
}

type UserSubscriptionsRepo interface {
	ListByUserID(ctx context.Context, userID uint) ([]UserSubscriptionsListRow, error)
	ListActiveIDs(ctx context.Context) ([]uint, error)
	FindOne(ctx context.Context, tx *gorm.DB, where map[string]interface{}, orderBy string) (entity.UserSubscriptions, error)
	FindActiveForUpdate(ctx context.Context, tx *gorm.DB, id uint) (*entity.UserSubscriptions, error)
	Create(ctx context.Context, tx *gorm.DB, m *entity.UserSubscriptions) error
	Update(ctx context.Context, tx *gorm.DB, where map[string]interface{}, updateData map[string]interface{}) error
	Reload(ctx context.Context, tx *gorm.DB, id uint) (*entity.UserSubscriptions, error)
}
