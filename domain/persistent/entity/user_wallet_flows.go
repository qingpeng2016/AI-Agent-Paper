package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type UserWalletFlows struct {
	ID           uint             `gorm:"primaryKey;column:id"`
	UserID       uint             `gorm:"column:user_id;not null"`
	Type         string           `gorm:"column:type;size:32;not null"`
	Amount       decimal.Decimal  `gorm:"column:amount;type:decimal(16,2);not null"`
	BalanceAfter *decimal.Decimal `gorm:"column:balance_after;type:decimal(16,2)"`
	Currency     string           `gorm:"column:currency;size:3;not null;default:CNY"`
	RefType      *string          `gorm:"column:ref_type;size:32"`
	RefID        *uint            `gorm:"column:ref_id"`
	Remark       *string          `gorm:"column:remark;size:512"`
	CreatedAt    time.Time        `gorm:"column:created_at"`
}

func (UserWalletFlows) TableName() string { return "user_wallet_flows" }
