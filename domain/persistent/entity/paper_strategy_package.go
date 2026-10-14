package entity

import "time"

type PaperStrategyPackage struct {
	ID        uint      `gorm:"primaryKey;column:id"`
	Code      string    `gorm:"column:code;size:64;not null"`
	Name      string    `gorm:"column:name;size:128;not null"`
	Status    string    `gorm:"column:status;size:16;not null;default:active"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (PaperStrategyPackage) TableName() string { return "paper_llm_strategy_package" }
