package entity

import "time"

type PaperLLMPromptTemplate struct {
	ID             uint      `gorm:"primaryKey;column:id"`
	StageCode      string    `gorm:"column:stage_code;size:64;not null"`
	StageName      string    `gorm:"column:stage_name;size:128;not null"`
	TemplateBody   string    `gorm:"column:template_body;type:mediumtext;not null"`
	Status         string    `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (PaperLLMPromptTemplate) TableName() string { return "paper_llm_prompt_template" }
