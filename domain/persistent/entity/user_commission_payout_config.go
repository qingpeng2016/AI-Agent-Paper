package entity

import "time"

type UserCommissionPayoutConfig struct {
	ID        uint      `gorm:"primaryKey;column:id"`
	UserID    uint      `gorm:"column:user_id;not null"`
	Channel   string    `gorm:"column:channel;size:16;not null"`
	QrMime    *string   `gorm:"column:qr_mime;size:64"`
	QrImage   []byte    `gorm:"column:qr_image;type:mediumblob"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (UserCommissionPayoutConfig) TableName() string { return "user_commission_payout_config" }

func (c *UserCommissionPayoutConfig) HasQR() bool {
	return len(c.QrImage) > 0 && c.QrMime != nil && *c.QrMime != ""
}
