package request

type SubmitEnterpriseInquiryReq struct {
	CompanyName string `json:"company_name" binding:"required,max=256"`
	ContactName string `json:"contact_name" binding:"required,max=128"`
	Phone       string `json:"phone" binding:"required,max=32"`
	Email       string `json:"email" binding:"omitempty,max=255,email"`
}
