package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type VipConfig struct {
	ID              uint            `gorm:"primaryKey;column:id"`
	LevelLabel      string          `gorm:"column:level_label;size:64;not null"`
	MinValidInvites      uint            `gorm:"column:min_valid_invites;not null;default:0"`
	MinInviteePaidAmount decimal.Decimal `gorm:"column:min_invitee_paid_amount;type:decimal(16,2);not null;default:0"`
	RatePercent          decimal.Decimal `gorm:"column:rate_percent;type:decimal(5,2);not null"`
	SortOrder       int             `gorm:"column:sort_order;not null;default:0"`
	Enabled         bool            `gorm:"column:enabled;not null;default:1"`
	IsDefault       bool            `gorm:"column:is_default;not null;default:0"`
	CreatedAt       time.Time       `gorm:"column:created_at"`
	UpdatedAt       time.Time       `gorm:"column:updated_at"`
}

func (VipConfig) TableName() string { return "vip_config" }
