package response

type UserSubscriptionItem struct {
	ID           uint   `json:"id"`
	ProductID    uint   `json:"product_id"`
	ProductName  string `json:"product_name"`
	Status       string `json:"status"`
	LimitTokens  int64  `json:"limit_tokens"`
	UsedTokens   int64  `json:"used_tokens"`
	PeriodEnd    string `json:"period_end"`
	ExpiresAt    string `json:"expires_at"`
	SKUProductName         string `json:"sku_product_name"`
	ProductsCategoryName   string `json:"products_category_name"`
}
