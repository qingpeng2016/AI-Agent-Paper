package request

type RegisterUserReq struct {
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	Password          string `json:"password" binding:"required,min=6"`
	ConfirmPassword   string `json:"confirm_password" binding:"required,min=6"`
	RegistrationHost  string `json:"registration_host"`
}

type LoginUserReq struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordReq struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=6"`
}

type ListWalletFlowsQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type ListInvoicesQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
