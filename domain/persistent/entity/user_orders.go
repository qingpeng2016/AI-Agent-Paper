package entity

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)

type UserOrders struct {
	ID                 uint            `gorm:"primaryKey;column:id"`
	OrderNo            string          `gorm:"column:order_no;size:64;not null"`
	UserID             uint            `gorm:"column:user_id;not null"`
	ProductID          uint            `gorm:"column:product_id;not null"`
	OrderType          string          `gorm:"column:order_type;size:32;not null;default:purchase"`
	UserSubscriptionID uint            `gorm:"column:user_subscription_id;not null;default:0"`
	Quantity           int             `gorm:"column:quantity;not null;default:1"`
	UnitPrice          decimal.Decimal `gorm:"column:unit_price;type:decimal(16,2);not null"`
	Status             string          `gorm:"column:status;size:32;not null"`
	TotalAmount        decimal.Decimal `gorm:"column:total_amount;type:decimal(16,2);not null"`
	Currency           string          `gorm:"column:currency;size:3;not null;default:CNY"`
	EnterpriseInvoice  int             `gorm:"column:enterprise_invoice;not null;default:0"`
	PayChannel         string          `gorm:"column:pay_channel;size:32"`
	OutTradeNo         string          `gorm:"column:out_trade_no;size:64"`
	ThirdTradeNo       *string         `gorm:"column:third_trade_no;size:128"`
	RawRequestJSON     datatypes.JSON  `gorm:"column:raw_request_json;type:json"`
	RawNotifyJSON      datatypes.JSON  `gorm:"column:raw_notify_json;type:json"`
	PaidAt             *time.Time      `gorm:"column:paid_at"`
	ExpireAt           time.Time       `gorm:"column:expire_at;not null"`
	ClosedAt           *time.Time      `gorm:"column:closed_at"`
	FailReason         *string         `gorm:"column:fail_reason;size:512"`
	CreatedAt          time.Time       `gorm:"column:created_at"`
	UpdatedAt          time.Time       `gorm:"column:updated_at"`
}

func (UserOrders) TableName() string { return "user_orders" }
