package entity

import "time"

type TutorialCategory struct {
	ID        uint      `gorm:"primaryKey;column:id"`
	Code      string    `gorm:"column:code;size:32;not null"`
	Name      string    `gorm:"column:name;size:64;not null"`
	Sort      int       `gorm:"column:sort;not null;default:0"`
	Status    string    `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (TutorialCategory) TableName() string { return "tutorial_category" }
