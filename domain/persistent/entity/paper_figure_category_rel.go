package entity

import "time"

type PaperFigureCategoryRel struct {
	ID         uint64    `gorm:"primaryKey;column:id"`
	FigureID   uint64    `gorm:"column:figure_id;not null"`
	CategoryID uint64    `gorm:"column:category_id;not null"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (PaperFigureCategoryRel) TableName() string { return "paper_user_figure_category_rel" }
