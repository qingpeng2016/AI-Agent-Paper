package repository

type ManuscriptLiteratureBinding struct {
	ManuscriptID     uint64 `gorm:"column:manuscript_id"`
	UserLiteratureID uint64 `gorm:"column:user_literature_id"`
	LiteratureTitle  string `gorm:"column:literature_title"`
}

// LiteratureManuscriptBinding 某条文献关联的论文（列表展示用）。
type LiteratureManuscriptBinding struct {
	UserLiteratureID uint64 `gorm:"column:user_literature_id"`
	ManuscriptID     uint64 `gorm:"column:manuscript_id"`
	ManuscriptTitle  string `gorm:"column:manuscript_title"`
}
