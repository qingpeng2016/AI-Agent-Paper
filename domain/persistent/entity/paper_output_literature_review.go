package entity

import (
	"time"

	"gorm.io/datatypes"
)

type PaperOutputLiteratureReview struct {
	ID             uint64         `gorm:"primaryKey;column:id"`
	ManuscriptID   uint64         `gorm:"column:manuscript_id;not null"`
	UserID         uint64         `gorm:"column:user_id;not null"`
	Version        int            `gorm:"column:version;not null;default:1"`
	IsCurrent      bool           `gorm:"column:is_current;not null;default:1"`
	Status         string         `gorm:"column:status;size:16;not null;default:completed"`
	Structure      *string        `gorm:"column:structure;size:32"`
	Title          *string        `gorm:"column:title;size:256"`
	Summary        *string        `gorm:"column:summary"`
	ContentMedium  *string        `gorm:"column:content_medium"`
	StorageURI     *string        `gorm:"column:storage_uri;size:1024"`
	Format         string         `gorm:"column:format;size:16;not null;default:md"`
	Citations      datatypes.JSON `gorm:"column:citations"`
	InputParams    datatypes.JSON `gorm:"column:input_params"`
	Meta           datatypes.JSON `gorm:"column:meta"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
}

func (PaperOutputLiteratureReview) TableName() string { return "paper_output_literature_review" }
