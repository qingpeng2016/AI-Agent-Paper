package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type UserCommissionRecords struct {
	ID             uint            `gorm:"primaryKey;column:id"`
	InviterUserID  uint            `gorm:"column:inviter_user_id;not null"`
	InviteeUserID  uint            `gorm:"column:invitee_user_id;not null"`
	OrderID        uint            `gorm:"column:order_id;not null"`
	OrderNo        string          `gorm:"column:order_no;size:64;not null"`
	ProductName    string          `gorm:"column:product_name;size:128;not null;default:''"`
	OrderAmount    decimal.Decimal `gorm:"column:order_amount;type:decimal(16,2);not null"`
	RatePercent    decimal.Decimal `gorm:"column:rate_percent;type:decimal(5,2);not null"`
	RebateAmount   decimal.Decimal `gorm:"column:rebate_amount;type:decimal(16,2);not null"`
	Status         string          `gorm:"column:status;size:32;not null;default:settled"`
	CreatedAt      time.Time       `gorm:"column:created_at"`
	UpdatedAt      time.Time       `gorm:"column:updated_at"`
}

func (UserCommissionRecords) TableName() string { return "user_commission_records" }
