package entity

import "time"

type PaperLLMPromptTemplate struct {
	ID               uint                  `gorm:"primaryKey;column:id"`
	StageCode        string                `gorm:"column:stage_code;size:64;not null"`
	StageName        string                `gorm:"column:stage_name;size:128;not null"`
	LLMDescriptionCode uint64               `gorm:"column:llm_description_code;not null"`
	Description        *PaperLLMDescription `gorm:"foreignKey:LLMDescriptionCode;references:DescriptionCode"`
	Status           string                `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt        time.Time             `gorm:"column:created_at"`
	UpdatedAt        time.Time             `gorm:"column:updated_at"`
}

func (p *PaperLLMPromptTemplate) TemplateBodyForLocale(locale string) string {
	if p == nil || p.Description == nil {
		return ""
	}
	return p.Description.BodyForLocale(locale)
}

func (PaperLLMPromptTemplate) TableName() string { return "paper_llm_template_topic" }
