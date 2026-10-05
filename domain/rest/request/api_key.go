package request

type CreateSubAPIKeyReq struct {
	UserSubscriptionID uint  `json:"user_subscription_id" binding:"required"`
	MemberUserID       uint  `json:"member_user_id" binding:"required"`
	LimitTokens        int64 `json:"limit_tokens" binding:"required,min=1"`
}

type UpdateSubAPIKeyLimitReq struct {
	LimitTokens int64 `json:"limit_tokens" binding:"required,min=1"`
}

type AddApiTeamMemberReq struct {
	UserID uint `json:"user_id" binding:"required"`
}

// ApiTeamEnterpriseInquiryReq 会员中心 API 团队登记（仅企业名必填，联系人/手机/邮箱由 users 表补全）。
type ApiTeamEnterpriseInquiryReq struct {
	CompanyName string `json:"company_name" binding:"required,max=256"`
}
