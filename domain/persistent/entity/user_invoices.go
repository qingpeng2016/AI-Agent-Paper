package entity

import (
	"time"

	"github.com/shopspring/decimal"
)

type UserInvoices struct {
	ID                  uint            `gorm:"primaryKey;column:id"`
	OrderID             uint            `gorm:"column:order_id;not null"`
	EnterpriseInquiryID uint            `gorm:"column:enterprise_inquiry_id;not null;default:0"`
	UserID              uint            `gorm:"column:user_id;not null"`
	InvoiceType string          `gorm:"column:invoice_type;size:32;not null"`
	Title       string          `gorm:"column:title;size:256;not null"`
	TaxNo       *string         `gorm:"column:tax_no;size:64"`
	Amount      decimal.Decimal `gorm:"column:amount;type:decimal(16,2);not null"`
	Status      string          `gorm:"column:status;size:32;not null"`
	FileURL     *string         `gorm:"column:file_url;size:512"`
	IssuedAt    *time.Time      `gorm:"column:issued_at"`
	CreatedAt   time.Time       `gorm:"column:created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at"`
}

func (UserInvoices) TableName() string { return "user_invoices" }
