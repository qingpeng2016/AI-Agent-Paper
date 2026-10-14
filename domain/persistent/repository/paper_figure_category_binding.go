package repository

// FigureCategoryBinding 某条图表绑定的分类（列表展示用）。
type FigureCategoryBinding struct {
	FigureID     uint64 `gorm:"column:figure_id"`
	CategoryID   uint64 `gorm:"column:category_id"`
	CategoryName string `gorm:"column:category_name"`
}
