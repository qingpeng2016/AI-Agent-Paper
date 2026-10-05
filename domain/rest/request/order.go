package request

type ListOrdersQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type CreateOrderReq struct {
	ProductID         uint   `json:"product_id" binding:"required"`
	OrderType          string `json:"order_type" binding:"omitempty,oneof=purchase renewal upgrade quota_addon"`
	UserSubscriptionID uint   `json:"user_subscription_id" binding:"omitempty"`
	Quantity           int    `json:"quantity" binding:"required,min=1,max=99"`
	Channel           string `json:"channel" binding:"required"`
	EnterpriseInvoice bool   `json:"enterprise_invoice"`
	UserCouponID      uint   `json:"user_coupon_id" binding:"omitempty"`
}

type PaymentNotifyReq struct {
	OutTradeNo   string `json:"out_trade_no" binding:"required"`
	TradeNo      string `json:"trade_no"`
	TradeStatus  string `json:"trade_status"`
}
