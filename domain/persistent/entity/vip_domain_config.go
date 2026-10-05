package entity

import "time"

type VipDomainConfig struct {
	ID         uint      `gorm:"primaryKey;column:id"`
	Domain     string    `gorm:"column:domain;size:255;not null"`
	IsOfficial bool      `gorm:"column:is_official;not null;default:0"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (VipDomainConfig) TableName() string { return "vip_domain_config" }
