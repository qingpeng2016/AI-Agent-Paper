package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type CouponCampaigns struct {
	ID                   uint            `gorm:"primaryKey;column:id"`
	Code                 string          `gorm:"column:code;size:32;not null"`
	Name                 string          `gorm:"column:name;size:128;not null"`
	Title                string          `gorm:"column:title;size:128;not null"`
	Subtitle             *string         `gorm:"column:subtitle;size:512"`
	DiscountType         string          `gorm:"column:discount_type;size:16;not null"`
	DiscountValue        decimal.Decimal `gorm:"column:discount_value;type:decimal(16,2);not null"`
	MinOrderAmount       decimal.Decimal `gorm:"column:min_order_amount;type:decimal(16,2);not null;default:0"`
	ValidDays            uint            `gorm:"column:valid_days;not null;default:30"`
	AutoGrantOnRegister  bool            `gorm:"column:auto_grant_on_register;not null;default:0"`
	Enabled              bool            `gorm:"column:enabled;not null;default:1"`
	CreatedAt            time.Time       `gorm:"column:created_at"`
	UpdatedAt            time.Time       `gorm:"column:updated_at"`
}

func (CouponCampaigns) TableName() string { return "coupon_campaigns" }
