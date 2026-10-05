package request

type InviteRebatePageQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type CommissionTransferToBalanceReq struct {
	Amount string `json:"amount" binding:"required"`
}

type CreateCommissionWithdrawalReq struct {
	Amount  string `json:"amount" binding:"required"`
	Channel string `json:"channel" binding:"required"`
}

