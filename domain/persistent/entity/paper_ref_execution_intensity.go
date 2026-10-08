package entity

import "github.com/shopspring/decimal"

type PaperRefExecutionIntensity struct {
	Code       string          `gorm:"primaryKey;column:code;size:16"`
	Name       string          `gorm:"column:name;size:64;not null"`
	Multiplier decimal.Decimal `gorm:"column:multiplier;type:decimal(4,2);not null"`
	MaxPapers  uint            `gorm:"column:max_papers;not null"`
	MaxIdeas   uint            `gorm:"column:max_ideas;not null"`
}

func (PaperRefExecutionIntensity) TableName() string { return "paper_ref_execution_intensity" }
