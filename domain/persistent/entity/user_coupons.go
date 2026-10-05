package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type UserCoupons struct {
	ID             uint            `gorm:"primaryKey;column:id"`
	UserID         uint            `gorm:"column:user_id;not null"`
	CampaignID     uint            `gorm:"column:campaign_id;not null"`
	CouponCode     string          `gorm:"column:coupon_code;size:40;not null"`
	DiscountType   string          `gorm:"column:discount_type;size:16;not null"`
	DiscountValue  decimal.Decimal `gorm:"column:discount_value;type:decimal(16,2);not null"`
	MinOrderAmount decimal.Decimal `gorm:"column:min_order_amount;type:decimal(16,2);not null;default:0"`
	Status         string          `gorm:"column:status;size:16;not null;default:available"`
	ValidFrom      time.Time       `gorm:"column:valid_from;not null"`
	ValidUntil     time.Time       `gorm:"column:valid_until;not null"`
	UsedAt         *time.Time      `gorm:"column:used_at"`
	OrderID        *uint           `gorm:"column:order_id"`
	CreatedAt      time.Time       `gorm:"column:created_at"`
	UpdatedAt      time.Time       `gorm:"column:updated_at"`
}

func (UserCoupons) TableName() string { return "user_coupons" }
