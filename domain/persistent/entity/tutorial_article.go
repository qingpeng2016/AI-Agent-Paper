package entity

import (
	"time"

	"gorm.io/datatypes"
)

type TutorialArticle struct {
	ID          uint           `gorm:"primaryKey;column:id"`
	CategoryID  uint           `gorm:"column:category_id;not null"`
	Slug        string         `gorm:"column:slug;size:128;not null"`
	Title       string         `gorm:"column:title;size:256;not null"`
	Excerpt     string         `gorm:"column:excerpt;size:512;not null"`
	Body        datatypes.JSON `gorm:"column:body;type:json;not null"`
	PublishedAt time.Time      `gorm:"column:published_at;type:date"`
	Sort        int            `gorm:"column:sort;not null;default:0"`
	Status      string         `gorm:"column:status;size:16;not null;default:published"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
}

func (TutorialArticle) TableName() string { return "tutorial_article" }
