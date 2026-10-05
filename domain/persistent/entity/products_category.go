package entity

import "time"

type ProductsCategory struct {
	ID         uint      `gorm:"primaryKey;column:id"`
	Name       string    `gorm:"column:name;size:64;not null"`
	DotColor   string    `gorm:"column:dot_color;size:16"`
	ActiveBg   string    `gorm:"column:active_bg;size:16"`
	Sort       int       `gorm:"column:sort;not null;default:0"`
	Status     string    `gorm:"column:status;size:16;not null;default:active"`
	HotTagName string    `gorm:"column:hot_tag_name;size:32"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (ProductsCategory) TableName() string { return "products_category" }
