package entity

import "time"

type PaperOutputFigure struct {
	ID             uint64    `gorm:"primaryKey;column:id"`
	ManuscriptID   uint64    `gorm:"column:manuscript_id;not null"`
	UserID         uint64    `gorm:"column:user_id;not null"`
	Origin         string    `gorm:"column:origin;size:16;not null"`
	FileKind       string    `gorm:"column:file_kind;size:16;not null"`
	FileName       string    `gorm:"column:file_name;size:512;not null"`
	Title          string    `gorm:"column:title;size:512;not null"`
	Note           *string   `gorm:"column:note"`
	SizeBytes      uint64    `gorm:"column:size_bytes;not null;default:0"`
	StorageURI     *string   `gorm:"column:storage_uri;size:1024"`
	ContentHash    *string   `gorm:"column:content_hash;size:64"`
	PromptSnapshot *string   `gorm:"column:prompt_snapshot"`
	Status         string    `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (PaperOutputFigure) TableName() string { return "paper_output_figure" }
