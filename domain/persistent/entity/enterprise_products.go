package entity

import (
	"time"

	"gorm.io/datatypes"
)

type EnterpriseProducts struct {
	ID          uint           `gorm:"primaryKey;column:id"`
	Code        string         `gorm:"column:code;size:32;not null"`
	Name        string         `gorm:"column:name;size:64;not null"`
	Badge       *string        `gorm:"column:badge;size:32"`
	PriceHint   string         `gorm:"column:price_hint;size:128;not null"`
	Seats       string         `gorm:"column:seats;size:128;not null"`
	Features    datatypes.JSON `gorm:"column:features;type:json;not null"`
	Tagline     string         `gorm:"column:tagline;size:256;not null"`
	ButtonLabel string         `gorm:"column:button_label;size:64;not null;default:获取报价"`
	IsFeatured  int            `gorm:"column:is_featured;not null;default:0"`
	Sort        int            `gorm:"column:sort;not null;default:0"`
	Status      string         `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
}

func (EnterpriseProducts) TableName() string { return "enterprise_products" }
