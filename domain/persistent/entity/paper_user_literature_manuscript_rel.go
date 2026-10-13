package entity

import "time"

type PaperUserLiteratureManuscriptRel struct {
	ID               uint64    `gorm:"primaryKey;column:id"`
	UserLiteratureID uint64    `gorm:"column:user_literature_id;not null"`
	ManuscriptID     uint64    `gorm:"column:manuscript_id;not null"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

func (PaperUserLiteratureManuscriptRel) TableName() string {
	return "paper_user_literature_manuscript_rel"
}
