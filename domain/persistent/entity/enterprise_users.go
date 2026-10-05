package entity

import "time"

type EnterpriseUsers struct {
	ID                  uint      `gorm:"primaryKey;column:id"`
	OwnerUserID         uint      `gorm:"column:owner_user_id;not null"`
	EnterpriseInquiryID uint      `gorm:"column:enterprise_inquiry_id;not null;default:0"`
	UserID              *uint     `gorm:"column:user_id"`
	MemberName          string    `gorm:"column:member_name;size:128;not null"`
	Email               *string   `gorm:"column:email;size:255"`
	Phone               *string   `gorm:"column:phone;size:32"`
	Department          *string   `gorm:"column:department;size:128"`
	JobTitle            *string   `gorm:"column:job_title;size:128"`
	Role                string    `gorm:"column:role;size:32;not null;default:member"`
	Status              string    `gorm:"column:status;size:32;not null;default:active"`
	Remark              *string   `gorm:"column:remark;size:512"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}

func (EnterpriseUsers) TableName() string { return "enterprise_users" }
