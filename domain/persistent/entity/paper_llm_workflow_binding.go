package entity

import "time"

type PaperLLMWorkflowBinding struct {
	ID            uint      `gorm:"primaryKey;column:id"`
	StageCode     string    `gorm:"column:stage_code;size:64;not null"`
	StageName     string    `gorm:"column:stage_name;size:128;not null"`
	ModelConfigID uint      `gorm:"column:model_config_id;not null"`
	Status        string    `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (PaperLLMWorkflowBinding) TableName() string { return "paper_llm_workflow_binding" }
