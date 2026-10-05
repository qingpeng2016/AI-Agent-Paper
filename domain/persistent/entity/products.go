package entity

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

type Products struct {
	ID                   uint            `gorm:"primaryKey;column:id"`
	SKUCode              string          `gorm:"column:sku_code;size:64;not null"`
	CardTitle            string          `gorm:"column:card_title;size:128;not null"`
	CardSubtitle         string          `gorm:"column:card_subtitle;size:512;not null"`
	CardFeatures         datatypes.JSON  `gorm:"column:card_features;type:json;not null"`
	ShareSeats           int             `gorm:"column:share_seats;not null;default:1"`
	ProductsCategoryID   *uint           `gorm:"column:products_category_id"`
	ProductsCategoryName string          `gorm:"column:products_category_name;size:32;not null"`
	SKUProductName       string          `gorm:"column:sku_product_name;size:128;not null"`
	LimitTokens          int64           `gorm:"column:limit_tokens;not null"`
	RPMLimit             int             `gorm:"column:rpm_limit;not null;default:0"`
	TPMLimit             *int            `gorm:"column:tpm_limit"`
	AllowedModels        datatypes.JSON  `gorm:"column:allowed_models;type:json;not null"`
	BillingPeriod        string          `gorm:"column:billing_period;size:16;not null;default:month"`
	PeriodDays           int             `gorm:"column:period_days;not null;default:30"`
	Price                decimal.Decimal `gorm:"column:price;type:decimal(16,2);not null"`
	Currency             string          `gorm:"column:currency;size:3;not null;default:CNY"`
	IsHot                int             `gorm:"column:is_hot;not null;default:0"`
	HotTagName           string          `gorm:"column:hot_tag_name;size:32"`
	SortOrder            int             `gorm:"column:sort_order;not null;default:0"`
	Status               string          `gorm:"column:status;size:16;not null;default:on_sale"`
	CreatedAt            time.Time       `gorm:"column:created_at"`
	UpdatedAt            time.Time       `gorm:"column:updated_at"`
}

func (Products) TableName() string { return "products" }
