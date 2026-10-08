package entity

import (
	"time"

	"gorm.io/datatypes"
)

type PaperLLMModelConfig struct {
	ID                uint           `gorm:"primaryKey;column:id"`
	Label             string         `gorm:"column:label;size:128;not null"`
	ProviderCode      string         `gorm:"column:provider_code;size:32;not null"`
	ModelName         string         `gorm:"column:model_name;size:128;not null"`
	APIBaseURL        string         `gorm:"column:api_base_url;size:512;not null"`
	APIKey            *string        `gorm:"column:api_key;size:512"`
	TimeoutMs         uint           `gorm:"column:timeout_ms;not null;default:120000"`
	MaxRetries        uint8          `gorm:"column:max_retries;not null;default:2"`
	SupportsVision    bool           `gorm:"column:supports_vision;not null;default:0"`
	ContextWindowHint *uint          `gorm:"column:context_window_hint"`
	Extra             datatypes.JSON `gorm:"column:extra"`
	Status            string         `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt         time.Time      `gorm:"column:created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at"`
}

func (PaperLLMModelConfig) TableName() string { return "paper_llm_model_config" }
