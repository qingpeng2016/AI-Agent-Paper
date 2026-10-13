package entity

import (
	"time"

	"gorm.io/datatypes"
)

type PaperOutputLiteratureReview struct {
	ID                           uint64         `gorm:"primaryKey;column:id"`
	ManuscriptID                 uint64         `gorm:"column:manuscript_id;not null"`
	UserID                       uint64         `gorm:"column:user_id;not null"`
	Version                      int            `gorm:"column:version;not null;default:1"`
	Status                       string         `gorm:"column:status;size:16;not null;default:completed"`
	From                         string         `gorm:"column:from;size:16;not null;default:generated"`
	Title                        *string        `gorm:"column:title;size:256"`
	ContentMedium                *string        `gorm:"column:content_medium"`
	PdfURI                       *string        `gorm:"column:pdf_uri;size:1024"`
	WordURI                      *string        `gorm:"column:word_uri;size:1024"`
	PptURI                       *string        `gorm:"column:ppt_uri;size:1024"`
	Format                       string         `gorm:"column:format;size:16;not null;default:md"`
	ExperimentPlanID             *uint64        `gorm:"column:experiment_plan_id"`
	ExperimentPlanLLMRequest     datatypes.JSON `gorm:"column:experiment_plan_llm_request"`
	ExperimentPlanLLMResponse    datatypes.JSON `gorm:"column:experiment_plan_llm_response"`
	CreatedAt                    time.Time      `gorm:"column:created_at"`
	UpdatedAt                    time.Time      `gorm:"column:updated_at"`
}

func (PaperOutputLiteratureReview) TableName() string { return "paper_output_literature_review" }
