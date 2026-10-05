package response

type UserAPIKeyItem struct {
	ID                   uint   `json:"id"`
	KeyType              string `json:"key_type"`
	KeyMasked            string `json:"key_masked"`
	Status               string `json:"status"`
	UserSubscriptionID   uint   `json:"user_subscription_id"`
	SubscriptionName     string `json:"subscription_name"`
	LimitTokens          int64  `json:"limit_tokens"`
	UsedTokens           int64  `json:"used_tokens"`
	MemberUserID         uint   `json:"member_user_id,omitempty"`
	MemberNickname       string `json:"member_nickname,omitempty"`
	MemberEmail          string `json:"member_email,omitempty"`
	// APIKey 仅创建子 Key 时返回一次，不入库。
	APIKey string `json:"api_key,omitempty"`
}

type ApiTeamMemberItem struct {
	ID       uint   `json:"id"`
	UserID   uint   `json:"user_id"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Status   string `json:"status"`
}

type ApiTeamAddableInviteeItem struct {
	UserID     uint   `json:"user_id"`
	Nickname   string `json:"nickname"`
	Email      string `json:"email"`
	RegisteredAt string `json:"registered_at"`
}

type ApiTeamMemberListResp struct {
	Items []ApiTeamMemberItem `json:"items"`
}

type ApiTeamAddableListResp struct {
	Items []ApiTeamAddableInviteeItem `json:"items"`
}

type ApiTeamEnterpriseInquiryStatusResp struct {
	HasInquiry bool                           `json:"has_inquiry"`
	Inquiry    *ApiTeamEnterpriseInquiryBrief `json:"inquiry,omitempty"`
}

type ApiTeamEnterpriseInquiryBrief struct {
	ID          uint   `json:"id"`
	CompanyName string `json:"company_name"`
	ContactName string `json:"contact_name"`
	Phone       string `json:"phone"`
	Email       string `json:"email,omitempty"`
	Status      string `json:"status"`
}

type ApiTeamEnterpriseInquirySubmitResp struct {
	ID      uint   `json:"id"`
	Message string `json:"message"`
}
