package entity

import (
	"time"

	"gorm.io/datatypes"
)

type Payments struct {
	ID             uint           `gorm:"primaryKey;column:id"`
	Code           string         `gorm:"column:code;size:32;not null"`
	DisplayName    string         `gorm:"column:display_name;size:64;not null"`
	Driver         string         `gorm:"column:driver;size:32;not null;default:epay"`
	APIBaseURL     string         `gorm:"column:api_base_url;size:512;not null"`
	MerchantID     string         `gorm:"column:merchant_id;size:64;not null"`
	MerchantSecret string         `gorm:"column:merchant_secret;size:256;not null"`
	AppID          *string        `gorm:"column:app_id;size:64"`
	ExtraJSON      datatypes.JSON `gorm:"column:extra_json;type:json"`
	NotifyPath     *string        `gorm:"column:notify_path;size:128"`
	IsEnabled      int            `gorm:"column:is_enabled;not null;default:0"`
	SortOrder      int            `gorm:"column:sort_order;not null;default:0"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
}

func (Payments) TableName() string { return "payments" }
