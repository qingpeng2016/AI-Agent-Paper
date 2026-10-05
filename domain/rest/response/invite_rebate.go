package response

type InviteRebateOverviewResp struct {
	PromoDomainURL      string                  `json:"promo_domain_url"`
	CurrentLevelLabel   string                  `json:"current_level_label"`
	CurrentRatePercent  Money                   `json:"current_rate_percent"`
	ValidInviteCount    int64                   `json:"valid_invite_count"`
	InviteePaidTotal    Money                   `json:"invitee_paid_total"`
	CommissionBalance   Money                   `json:"commission_balance"`
	Tiers               []InviteRebateTierItem  `json:"tiers"`
	Notes               []string                `json:"notes"`
}

type InviteRebateTierItem struct {
	LevelLabel      string `json:"level_label"`
	MinValidInvites      uint  `json:"min_valid_invites"`
	MinInviteePaidAmount Money `json:"min_invitee_paid_amount"`
	RatePercent          Money `json:"rate_percent"`
	IsCurrent       bool   `json:"is_current"`
}

type InviteRebateMemberItem struct {
	ID           uint   `json:"id"`
	Nickname     string `json:"nickname"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Status       string `json:"status"`
	RegisteredAt string `json:"registered_at"`
}

type InviteRebateMemberListResp struct {
	Items []InviteRebateMemberItem `json:"items"`
	Total int64                    `json:"total"`
}

type InviteCommissionRecordItem struct {
	ID              uint   `json:"id"`
	InviteeNickname string `json:"invitee_nickname"`
	InviteeEmail    string `json:"invitee_email"`
	OrderNo         string `json:"order_no"`
	ProductName     string `json:"product_name"`
	OrderAmount     Money  `json:"order_amount"`
	RebateAmount    Money  `json:"rebate_amount"`
	CreatedAt       string `json:"created_at"`
}

type InviteCommissionRecordListPageResp struct {
	Items    []InviteCommissionRecordItem `json:"items"`
	Total    int64                        `json:"total"`
	Page     int                          `json:"page"`
	PageSize int                          `json:"page_size"`
}

type InviteWithdrawalItem struct {
	ID        uint   `json:"id"`
	Amount    Money  `json:"amount"`
	Channel   string `json:"channel"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type InviteWithdrawalListPageResp struct {
	Items    []InviteWithdrawalItem `json:"items"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

type InvitePayoutConfigResp struct {
	AlipayQrDataURL  string `json:"alipay_qr_data_url"`
	WechatQrDataURL  string `json:"wechat_qr_data_url"`
	AlipayConfigured bool   `json:"alipay_configured"`
	WechatConfigured bool   `json:"wechat_configured"`
}

type CommissionTransferToBalanceResp struct {
	TransferredAmount Money `json:"transferred_amount"`
	CommissionBalance Money `json:"commission_balance"`
	WalletBalance     Money `json:"wallet_balance"`
}

type CreateCommissionWithdrawalResp struct {
	Withdrawal        InviteWithdrawalItem `json:"withdrawal"`
	CommissionBalance Money                `json:"commission_balance"`
}
