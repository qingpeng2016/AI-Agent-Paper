package entity

import "time"

type PaperManuscript struct {
	ID           uint      `gorm:"primaryKey;column:id"`
	UserID       uint      `gorm:"column:user_id;not null"`
	Title        string    `gorm:"column:title;size:256;not null"`
	DisciplineID     *uint64 `gorm:"column:discipline_id"`
	ContentLanguage  string  `gorm:"column:content_language;size:8;not null;default:en"`
	Status           string  `gorm:"column:status;size:16;not null;default:active"`
	IsCurrent    bool      `gorm:"column:is_current;not null;default:0"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (PaperManuscript) TableName() string { return "paper_manuscript" }
