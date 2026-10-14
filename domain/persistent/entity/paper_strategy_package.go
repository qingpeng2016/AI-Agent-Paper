package entity

import (
	"time"

	"gorm.io/datatypes"
)

type PaperStrategyPackage struct {
	ID              uint           `gorm:"primaryKey;column:id"`
	Code            string         `gorm:"column:code;size:64;not null"`
	Name            string         `gorm:"column:name;size:128;not null"`
	ProviderKind    string         `gorm:"column:provider_kind;size:16;not null;default:inline"`
	BaseURL         *string        `gorm:"column:base_url;size:512"`
	AuthType        string         `gorm:"column:auth_type;size:32;not null;default:none"`
	AuthSecretRef   *string        `gorm:"column:auth_secret_ref;size:256"`
	ManifestVersion *string        `gorm:"column:manifest_version;size:32"`
	Capabilities    datatypes.JSON `gorm:"column:capabilities;type:json"`
	Status          string         `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
}

func (PaperStrategyPackage) TableName() string { return "paper_llm_strategy_package" }

const (
	StrategyProviderInline = "inline"
	StrategyProviderHTTP   = "http"
)
