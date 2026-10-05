package entity

import (
	"time"

	"gorm.io/datatypes"
)

type UserSubscriptions struct {
	ID                   uint           `gorm:"primaryKey;column:id"`
	UserID               uint           `gorm:"column:user_id;not null"`
	ProductID            uint           `gorm:"column:product_id;not null"`
	Orders               datatypes.JSON `gorm:"column:orders;type:json;not null"`
	ProductsCategoryName string         `gorm:"column:products_category_name;size:32;not null"`
	SKUProductName       string         `gorm:"column:sku_product_name;size:128;not null"`
	BaseLimitTokens      int64          `gorm:"column:base_limit_tokens;not null"`
	LimitTokens          int64          `gorm:"column:limit_tokens;not null"`
	UsedTokens           int64          `gorm:"column:used_tokens;not null;default:0"`
	StartedAt            time.Time      `gorm:"column:started_at;not null"`
	ExpiresAt            time.Time      `gorm:"column:expires_at;not null"`
	PeriodStart          time.Time      `gorm:"column:period_start;not null"`
	PeriodEnd            time.Time      `gorm:"column:period_end;not null"`
	Status               string         `gorm:"column:status;size:32;not null;default:active"`
	CreatedAt            time.Time      `gorm:"column:created_at"`
	UpdatedAt            time.Time      `gorm:"column:updated_at"`
}

func (UserSubscriptions) TableName() string { return "user_subscriptions" }
