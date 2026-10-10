package entity

import (
	"strings"
	"time"
)

type PaperLLMPromptTemplate struct {
	ID             uint      `gorm:"primaryKey;column:id"`
	StageCode      string    `gorm:"column:stage_code;size:64;not null"`
	StageName      string    `gorm:"column:stage_name;size:128;not null"`
	TemplateBodyZh string    `gorm:"column:template_body_zh;type:mediumtext;not null"`
	TemplateBodyEn string    `gorm:"column:template_body_en;type:mediumtext;not null"`
	Status         string    `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

// TemplateBodyForLocale 返回渲染用语术；locale 为 en 时用英文，否则用中文。
func (p *PaperLLMPromptTemplate) TemplateBodyForLocale(locale string) string {
	if p == nil {
		return ""
	}
	if strings.EqualFold(strings.TrimSpace(locale), "en") {
		if s := strings.TrimSpace(p.TemplateBodyEn); s != "" {
			return p.TemplateBodyEn
		}
	}
	if s := strings.TrimSpace(p.TemplateBodyZh); s != "" {
		return p.TemplateBodyZh
	}
	return p.TemplateBodyEn
}

func (PaperLLMPromptTemplate) TableName() string { return "paper_template_llm_prompt" }
