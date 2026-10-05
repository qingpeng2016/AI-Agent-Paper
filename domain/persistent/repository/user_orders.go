package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

type UserOrdersListRow struct {
	entity.UserOrders
	ProductCardTitle string `gorm:"column:product_card_title"`
	ProductSKUName   string `gorm:"column:product_sku_name"`
}

type PaymentNotifyInput struct {
	Channel        string
	OutTradeNo     string
	ThirdTradeNo   string
	IdempotencyKey string
	PayloadJSON    []byte
}

type UserOrdersRepo interface {
	ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]UserOrdersListRow, error)
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	CreateOrder(ctx context.Context, order *entity.UserOrders) error
	CreateRenewalOrderReplacingPending(ctx context.Context, order *entity.UserOrders) error
	CreateUpgradeOrderReplacingPending(ctx context.Context, order *entity.UserOrders) error
	CreateQuotaAddonOrderReplacingPending(ctx context.Context, order *entity.UserOrders) error
	FindOrderByOutTradeNo(ctx context.Context, outTradeNo string) (*entity.UserOrders, error)
	FindOrderByID(ctx context.Context, id uint) (*entity.UserOrders, error)
	FindActiveSubscriptionByUserProduct(ctx context.Context, userID, productID uint) (*entity.UserSubscriptions, error)
	FindLatestSubscriptionByUserProduct(ctx context.Context, userID, productID uint) (*entity.UserSubscriptions, error)
	FindSubscriptionForUser(ctx context.Context, userID, subscriptionID uint) (*entity.UserSubscriptions, error)
	CreateTx(ctx context.Context, tx *gorm.DB, order *entity.UserOrders) error
	FindByOutTradeNoForUpdate(ctx context.Context, tx *gorm.DB, outTradeNo string) (*entity.UserOrders, error)
	UpdateFields(ctx context.Context, tx *gorm.DB, id uint, fields map[string]interface{}) error
}
