package entity

import "time"

type PaperFigureManuscriptRel struct {
	ID           uint64    `gorm:"primaryKey;column:id"`
	FigureID     uint64    `gorm:"column:figure_id;not null"`
	ManuscriptID uint64    `gorm:"column:manuscript_id;not null"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (PaperFigureManuscriptRel) TableName() string { return "paper_user_figure_manuscript_rel" }
