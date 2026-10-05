package entity

import "time"

type UserAPIKeys struct {
	ID                   uint       `gorm:"primaryKey;column:id"`
	UserID               uint       `gorm:"column:user_id;not null"`
	OwnerUserID          uint       `gorm:"column:owner_user_id;not null;default:0"`
	EnterpriseInquiryID  uint       `gorm:"column:enterprise_inquiry_id;not null;default:0"`
	UserSubscriptionID   uint       `gorm:"column:user_subscription_id;not null"`
	KeyType              string     `gorm:"column:key_type;size:16;not null;default:main"`
	KeyHash              string     `gorm:"column:key_hash;size:64;not null"`
	ProductsCategoryName string     `gorm:"column:products_category_name;size:32;not null"`
	LimitTokens          int64      `gorm:"column:limit_tokens;not null"`
	UsedTokens           int64      `gorm:"column:used_tokens;not null;default:0"`
	Status               string     `gorm:"column:status;size:32;not null;default:active"`
	RotatedAt            *time.Time `gorm:"column:rotated_at"`
	CreatedAt            time.Time  `gorm:"column:created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at"`
}

func (UserAPIKeys) TableName() string { return "user_api_keys" }
