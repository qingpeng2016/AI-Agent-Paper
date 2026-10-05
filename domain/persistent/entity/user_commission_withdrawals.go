package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type UserCommissionWithdrawals struct {
	ID           uint            `gorm:"primaryKey;column:id"`
	UserID       uint            `gorm:"column:user_id;not null"`
	Amount       decimal.Decimal `gorm:"column:amount;type:decimal(16,2);not null"`
	Channel      string          `gorm:"column:channel;size:16;not null"`
	PayoutQrURL  string          `gorm:"column:payout_qr_url;size:512;not null;default:''"`
	Status       string          `gorm:"column:status;size:32;not null;default:pending"`
	FailReason   *string         `gorm:"column:fail_reason;size:512"`
	ProcessedAt  *time.Time      `gorm:"column:processed_at"`
	CreatedAt    time.Time       `gorm:"column:created_at"`
	UpdatedAt    time.Time       `gorm:"column:updated_at"`
}

func (UserCommissionWithdrawals) TableName() string { return "user_commission_withdrawals" }
