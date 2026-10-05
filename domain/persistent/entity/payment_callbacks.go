package entity

import (
	"time"

	"gorm.io/datatypes"
)

type PaymentCallbacks struct {
	ID             uint           `gorm:"primaryKey;column:id"`
	Channel        string         `gorm:"column:channel;size:32;not null"`
	IdempotencyKey string         `gorm:"column:idempotency_key;size:128;not null"`
	PayloadJSON    datatypes.JSON `gorm:"column:payload_json;type:json;not null"`
	SignatureOK    int            `gorm:"column:signature_ok;not null;default:0"`
	ProcessResult  string         `gorm:"column:process_result;size:32;not null"`
	ProcessedAt    time.Time      `gorm:"column:processed_at"`
}

func (PaymentCallbacks) TableName() string { return "payment_callbacks" }
