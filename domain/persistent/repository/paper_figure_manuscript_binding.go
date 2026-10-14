package repository

type ManuscriptFigureBinding struct {
	ManuscriptID uint64 `gorm:"column:manuscript_id"`
	FigureID     uint64 `gorm:"column:figure_id"`
	FigureTitle  string `gorm:"column:figure_title"`
}

// FigureManuscriptBinding 某条图表关联的论文（列表展示用）。
type FigureManuscriptBinding struct {
	FigureID        uint64 `gorm:"column:figure_id"`
	ManuscriptID    uint64 `gorm:"column:manuscript_id"`
	ManuscriptTitle string `gorm:"column:manuscript_title"`
}
