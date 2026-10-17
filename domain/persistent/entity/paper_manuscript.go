package entity

import (
	"time"
)

type PaperManuscript struct {
	ID           uint      `gorm:"primaryKey;column:id"`
	UserID       uint      `gorm:"column:user_id;not null"`
	Title        string    `gorm:"column:title;size:256;not null"`
	DisciplineID     *uint64 `gorm:"column:discipline_id"`
	Language         string  `gorm:"column:language;size:8;not null;default:en"`
	Status           string  `gorm:"column:status;size:16;not null;default:active"`
	IsCurrent    bool      `gorm:"column:is_current;not null;default:0"`
	DraftVersion uint      `gorm:"column:draft_version;not null;default:0"`
	DraftStatus  string    `gorm:"column:draft_status;size:16;not null;default:writing"`
	SubmissionStatus   string     `gorm:"column:submission_status;size:24;not null;default:not_submitted"`
	SubmissionLabel    *string    `gorm:"column:submission_label;size:128"`
	SubmissionDeadline *time.Time `gorm:"column:submission_deadline"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (PaperManuscript) TableName() string { return "paper_manuscript" }
