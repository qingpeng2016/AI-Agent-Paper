package entity

import "time"

type EnterpriseInquiry struct {
	ID          uint      `gorm:"primaryKey;column:id"`
	OwnerUserID *uint     `gorm:"column:owner_user_id"`
	CompanyName string    `gorm:"column:company_name;size:256;not null"`
	ContactName string    `gorm:"column:contact_name;size:128;not null"`
	Phone       string    `gorm:"column:phone;size:32;not null"`
	Email       *string   `gorm:"column:email;size:255"`
	Status      string    `gorm:"column:status;size:32;not null;default:pending"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (EnterpriseInquiry) TableName() string { return "enterprise_inquiry" }
