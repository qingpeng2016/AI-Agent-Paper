package entity

import (
	"time"

	"gorm.io/datatypes"
)

type PaperRefLiteratureSource struct {
	ID             uint           `gorm:"primaryKey;column:id"`
	Code           string         `gorm:"column:code;size:64;not null"`
	Name           string         `gorm:"column:name;size:128;not null"`
	APIKind        string         `gorm:"column:api_kind;size:32;not null"`
	BaseURL        string         `gorm:"column:base_url;size:512;not null"`
	AuthType       string         `gorm:"column:auth_type;size:32;not null;default:none"`
	ConfigSchema   datatypes.JSON `gorm:"column:config_schema"`
	DefaultConfig  datatypes.JSON `gorm:"column:default_config"`
	RateLimitHint   *string        `gorm:"column:rate_limit_hint;size:256"`
	Priority        int            `gorm:"column:priority;not null;default:100"`
	DefaultSelected bool           `gorm:"column:default_selected;not null;default:0"`
	Status          string         `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
}

func (PaperRefLiteratureSource) TableName() string { return "paper_ref_literature_source" }
