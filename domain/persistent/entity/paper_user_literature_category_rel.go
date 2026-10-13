package entity

import "time"

type PaperUserLiteratureCategoryRel struct {
	ID               uint64    `gorm:"primaryKey;column:id"`
	UserLiteratureID uint64    `gorm:"column:user_literature_id;not null"`
	CategoryID       uint64    `gorm:"column:category_id;not null"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

func (PaperUserLiteratureCategoryRel) TableName() string {
	return "paper_user_literature_category_rel"
}
