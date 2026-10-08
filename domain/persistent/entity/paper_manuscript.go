package entity

import "time"

type PaperManuscript struct {
	ID        uint      `gorm:"primaryKey;column:id"`
	UserID    uint      `gorm:"column:user_id;not null"`
	Title     string    `gorm:"column:title;size:256;not null"`
	Status    string    `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (PaperManuscript) TableName() string { return "paper_manuscript" }
