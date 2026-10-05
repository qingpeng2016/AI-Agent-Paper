package entity

import "time"

type UserInvoiceConfig struct {
	ID                  uint      `gorm:"primaryKey;column:id"`
	UserID              uint      `gorm:"column:user_id;not null"`
	EnterpriseInquiryID uint      `gorm:"column:enterprise_inquiry_id;not null;default:0"`
	ProfileType string    `gorm:"column:profile_type;size:16;not null;default:enterprise"`
	Title       string    `gorm:"column:title;size:256;not null"`
	TaxNo       *string   `gorm:"column:tax_no;size:64"`
	BankName    *string   `gorm:"column:bank_name;size:128"`
	BankAccount *string   `gorm:"column:bank_account;size:64"`
	Address     *string   `gorm:"column:address;size:512"`
	Phone       *string   `gorm:"column:phone;size:32"`
	IsDefault   int       `gorm:"column:is_default;not null;default:0"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (UserInvoiceConfig) TableName() string { return "user_invoice_config" }
