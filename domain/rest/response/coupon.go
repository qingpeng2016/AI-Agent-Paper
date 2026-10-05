package response

type RegisterCouponPromoResp struct {
	Active      bool    `json:"active"`
	Title       string  `json:"title"`
	Subtitle    string  `json:"subtitle"`
	DiscountLabel string `json:"discount_label"`
	ValidDays   uint    `json:"valid_days"`
	CampaignCode string `json:"campaign_code"`
}

type UserCouponItem struct {
	ID             uint   `json:"id"`
	CouponCode     string `json:"coupon_code"`
	CampaignCode   string `json:"campaign_code"`
	Title          string `json:"title"`
	DiscountType   string `json:"discount_type"`
	DiscountValue  string `json:"discount_value"`
	MinOrderAmount string `json:"min_order_amount"`
	Status         string `json:"status"`
	ValidFrom      string `json:"valid_from"`
	ValidUntil     string `json:"valid_until"`
}

type UserCouponListResp struct {
	Items []UserCouponItem `json:"items"`
}

type GrantRegisterCouponsResp struct {
	GrantedCount int `json:"granted_count"`
}
