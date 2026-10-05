package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type Users struct {
	ID            uint       `gorm:"primaryKey;column:id"`
	Email         *string    `gorm:"column:email;size:255"`
	Phone         *string    `gorm:"column:phone;size:32"`
	PasswordHash  string     `gorm:"column:password_hash;size:255;not null"`
	PasswordPlain string     `gorm:"column:password_plain;size:255;not null"`
	Nickname       *string    `gorm:"column:nickname;size:64"`
	ParentUserID   uint       `gorm:"column:parent_user_id;not null;default:0"`
	VipConfigID    uint       `gorm:"column:vip_config_id;not null;default:1"`
	VipDomain      *string    `gorm:"column:vip_domain;size:255"`
	Status        string          `gorm:"column:status;size:32;not null;default:active"`
	WalletBalance      decimal.Decimal `gorm:"column:wallet_balance;type:decimal(16,2);not null;default:0"`
	CommissionBalance  decimal.Decimal `gorm:"column:commission_balance;type:decimal(16,2);not null;default:0"`
	LastLoginAt        *time.Time      `gorm:"column:last_login_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (Users) TableName() string { return "users" }
