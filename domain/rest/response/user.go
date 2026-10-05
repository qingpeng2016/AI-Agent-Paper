package response

type UserProfileResp struct {
	ID                uint   `json:"id"`
	Email             string `json:"email,omitempty"`
	Phone             string `json:"phone,omitempty"`
	Nickname          string `json:"nickname,omitempty"`
	Status            string `json:"status"`
	WalletBalance     Money  `json:"wallet_balance"`
	CommissionBalance Money  `json:"commission_balance"`
	CreatedAt         string `json:"created_at"`
	LastLoginAt       string `json:"last_login_at,omitempty"`
}

type UserWalletFlowListPageResp struct {
	Items    []UserWalletFlowItem `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

type UserWalletFlowItem struct {
	ID        uint   `json:"id"`
	Type      string `json:"type"`
	Amount    Money  `json:"amount"`
	Remark    string `json:"remark"`
	CreatedAt string `json:"created_at"`
}

type LoginUserResp struct {
	Token                  string          `json:"token"`
	User                   UserProfileResp `json:"user"`
	RegisterCouponsGranted int             `json:"register_coupons_granted,omitempty"`
}
