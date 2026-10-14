package entity

import "time"

type PaperFigureCategory struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	Name      string    `gorm:"column:name;size:128;not null"`
	Status    string    `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (PaperFigureCategory) TableName() string { return "paper_user_figure_category" }
