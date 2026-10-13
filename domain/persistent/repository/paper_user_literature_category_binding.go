package repository

// UserLiteratureCategoryBinding 某条文献绑定的分类（列表展示用）。
type UserLiteratureCategoryBinding struct {
	UserLiteratureID uint64 `gorm:"column:user_literature_id"`
	CategoryID       uint64 `gorm:"column:category_id"`
	CategoryName     string `gorm:"column:category_name"`
}
