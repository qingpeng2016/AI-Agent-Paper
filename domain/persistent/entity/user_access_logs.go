package entity

import "time"

type UserAccessLogs struct {
	ID                   uint      `gorm:"primaryKey;column:id"`
	RequestID            string    `gorm:"column:request_id;size:64;not null"`
	UserID               *uint     `gorm:"column:user_id"`
	UserSubscriptionID   *uint     `gorm:"column:user_subscription_id"`
	APIKeyID             *uint     `gorm:"column:api_key_id"`
	ProductsCategoryName string    `gorm:"column:products_category_name;size:32;not null"`
	SKUProductName       *string   `gorm:"column:sku_product_name;size:128"`
	UpstreamInfoID       *uint     `gorm:"column:upstream_info_id"`
	Model                *string   `gorm:"column:model;size:128"`
	HTTPMethod           string    `gorm:"column:http_method;size:16;not null"`
	Path                 string    `gorm:"column:path;size:512;not null"`
	ClientIP             *string   `gorm:"column:client_ip;size:64"`
	GatewayStatus        int16     `gorm:"column:gateway_status;not null"`
	UpstreamStatus       *int16    `gorm:"column:upstream_status"`
	LatencyMs            int       `gorm:"column:latency_ms;not null;default:0"`
	TokensPrompt         *int      `gorm:"column:tokens_prompt"`
	TokensCompletion     *int      `gorm:"column:tokens_completion"`
	TokensTotal          *int      `gorm:"column:tokens_total"`
	IsStream             int       `gorm:"column:is_stream;not null;default:0"`
	ErrorCode            *string   `gorm:"column:error_code;size:32"`
	CreatedAt            time.Time `gorm:"column:created_at"`
}

func (UserAccessLogs) TableName() string { return "user_access_logs" }
