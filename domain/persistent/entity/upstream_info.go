package entity

import "time"

type UpstreamInfo struct {
	ID                   uint       `gorm:"primaryKey;column:id"`
	ProductsCategoryName string     `gorm:"column:products_category_name;size:32;not null"`
	SKUProductName       string     `gorm:"column:sku_product_name;size:128;not null"`
	AccountLabel         string     `gorm:"column:account_label;size:128;not null"`
	APIKeyCiphertext     []byte     `gorm:"column:api_key_ciphertext;not null"`
	APISecretCiphertext  []byte     `gorm:"column:api_secret_ciphertext"`
	ProcurementCostNote  *string    `gorm:"column:procurement_cost_note;size:256"`
	ExpiresAt            *time.Time `gorm:"column:expires_at"`
	CapTokens            *int64     `gorm:"column:cap_tokens"`
	UsedTokens           int64      `gorm:"column:used_tokens;not null;default:0"`
	QuotaResetAt         *time.Time `gorm:"column:quota_reset_at"`
	Weight               int        `gorm:"column:weight;not null;default:100"`
	Status               string     `gorm:"column:status;size:32;not null;default:active"`
	FailRate             *float64   `gorm:"column:fail_rate"`
	CreatedAt            time.Time  `gorm:"column:created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at"`
}

func (UpstreamInfo) TableName() string { return "upstream_info" }
