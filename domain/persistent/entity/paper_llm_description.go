package entity

import (
	"strings"
	"time"
)

type PaperLLMDescription struct {
	ID              uint64 `gorm:"column:id;primaryKey;autoIncrement"`
	DescriptionCode uint64 `gorm:"column:description_code;not null;uniqueIndex:uk_paper_llm_description_code"`
	ModelConfigID   uint64 `gorm:"column:model_config_id;not null;default:0"`
	PromptCode    string `gorm:"column:prompt_code;size:64;not null"`
	PromptType     string    `gorm:"column:prompt_type;size:128;not null"`
	PromptName     string    `gorm:"column:prompt_name;size:256;not null;default:''"`
	LlmDescZh      string    `gorm:"column:llm_desc_zh;type:mediumtext;not null"`
	LlmDescEn      string    `gorm:"column:llm_desc_en;type:mediumtext;not null"`
	Status         string    `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (PaperLLMDescription) TableName() string { return "paper_llm_description" }

// BodyForLocale 返回注入 LLM 的正文；locale 为 en 时用英文，否则用中文。
func (d *PaperLLMDescription) BodyForLocale(locale string) string {
	if d == nil {
		return ""
	}
	if strings.EqualFold(strings.TrimSpace(locale), "en") {
		if s := strings.TrimSpace(d.LlmDescEn); s != "" {
			return d.LlmDescEn
		}
	}
	if s := strings.TrimSpace(d.LlmDescZh); s != "" {
		return d.LlmDescZh
	}
	return d.LlmDescEn
}
