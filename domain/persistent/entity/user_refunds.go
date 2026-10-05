package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type UserRefunds struct {
	ID         uint            `gorm:"primaryKey;column:id"`
	OrderID    uint            `gorm:"column:order_id;not null"`
	RefundNo   string          `gorm:"column:refund_no;size:64;not null"`
	Amount     decimal.Decimal `gorm:"column:amount;type:decimal(16,2);not null"`
	Reason     *string         `gorm:"column:reason;size:512"`
	Status     string          `gorm:"column:status;size:32;not null"`
	RefundedAt *time.Time      `gorm:"column:refunded_at"`
	CreatedAt  time.Time       `gorm:"column:created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at"`
}

func (UserRefunds) TableName() string { return "user_refunds" }
