package response

type UserOrderListPageResp struct {
	Items    []UserOrderListItem `json:"items"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

type UserOrderListItem struct {
	ID                uint   `json:"id"`
	OrderNo           string `json:"order_no"`
	OrderType         string `json:"order_type"`
	ProductName       string `json:"product_name"`
	Quantity          int    `json:"quantity"`
	TotalAmount       Money  `json:"total_amount"`
	Status            string `json:"status"`
	EnterpriseInvoice bool   `json:"enterprise_invoice"`
	OutTradeNo        string `json:"out_trade_no"`
	CreatedAt         string `json:"created_at"`
}

type CreateOrderResp struct {
	OrderNo          string `json:"order_no"`
	OutTradeNo       string `json:"out_trade_no"`
	OrderID          uint   `json:"order_id"`
	OrderType          string `json:"order_type"`
	UserSubscriptionID uint   `json:"user_subscription_id"`
	Channel            string `json:"channel"`
	Status           string `json:"status"`
	TotalAmount Money `json:"total_amount"`
	Currency         string `json:"currency"`
}
