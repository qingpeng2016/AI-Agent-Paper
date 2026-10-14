package entity

import (
	"time"

	"gorm.io/datatypes"
)

type PaperLLMCallLog struct {
	ID                 uint64         `gorm:"primaryKey;column:id"`
	ManuscriptID       uint64         `gorm:"column:manuscript_id;not null"`
	StageCode          *string        `gorm:"column:stage_code;size:64"`
	BehaviorKey        *string        `gorm:"column:behavior_key;size:128"`
	StrategyPackageID  *uint64        `gorm:"column:strategy_package_id"`
	CapabilityID       *string        `gorm:"column:capability_id;size:128"`
	ExternalTraceID    *string        `gorm:"column:external_trace_id;size:128"`
	ModelConfigID      *uint          `gorm:"column:model_config_id"`
	ModelName          string         `gorm:"column:model_name;size:128;not null"`
	RequestJSON        datatypes.JSON `gorm:"column:request_json;not null"`
	ResponseJSON       datatypes.JSON `gorm:"column:response_json;not null"`
	PromptTokens       *int           `gorm:"column:prompt_tokens"`
	CompletionTokens   *int           `gorm:"column:completion_tokens"`
	LatencyMs          int            `gorm:"column:latency_ms;not null;default:0"`
	Status             int16          `gorm:"column:status;not null"`
	CreatedAt          time.Time      `gorm:"column:created_at"`
}

func (PaperLLMCallLog) TableName() string { return "paper_llm_call_logs" }
