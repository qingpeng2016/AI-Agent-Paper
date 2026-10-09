package entity

import "time"

type PaperLLMCallLog struct {
	ID                 uint64    `gorm:"primaryKey;column:id"`
	ManuscriptID       uint64    `gorm:"column:manuscript_id;not null"`
	StageCode          *string   `gorm:"column:stage_code;size:64"`
	ModelConfigID      *uint     `gorm:"column:model_config_id"`
	WorkflowBindingID  *uint     `gorm:"column:workflow_binding_id"`
	PromptTemplateID   *uint     `gorm:"column:prompt_template_id"`
	ModelName          string    `gorm:"column:model_name;size:128;not null"`
	PromptTokens       *int      `gorm:"column:prompt_tokens"`
	CompletionTokens   *int      `gorm:"column:completion_tokens"`
	LatencyMs          int       `gorm:"column:latency_ms;not null;default:0"`
	Status             int16     `gorm:"column:status;not null"`
	CreatedAt          time.Time `gorm:"column:created_at"`
}

func (PaperLLMCallLog) TableName() string { return "paper_llm_call_logs" }
