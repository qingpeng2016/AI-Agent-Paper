package entity

import (
	"time"

	"gorm.io/datatypes"
)

type PaperRefDiscipline struct {
	ID                    uint           `gorm:"primaryKey;column:id"`
	Code                  string         `gorm:"column:code;size:64;not null"`
	Name                  string         `gorm:"column:name;size:128;not null"`
	NameEn                *string        `gorm:"column:name_en;size:128"`
	Sort                  int            `gorm:"column:sort;not null;default:0"`
	LiteratureSourceCodes datatypes.JSON `gorm:"column:literature_source_codes;not null"`
	Status                string         `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt             time.Time      `gorm:"column:created_at"`
	UpdatedAt             time.Time      `gorm:"column:updated_at"`
}

func (PaperRefDiscipline) TableName() string { return "paper_ref_discipline" }
