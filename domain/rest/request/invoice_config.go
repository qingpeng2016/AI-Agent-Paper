package request

type SaveInvoiceConfigReq struct {
	ProfileType string  `json:"profile_type" binding:"omitempty,oneof=enterprise personal"`
	Title       string  `json:"title" binding:"max=256"`
	TaxNo       *string `json:"tax_no" binding:"omitempty,max=64"`
	BankName    *string `json:"bank_name" binding:"omitempty,max=128"`
	BankAccount *string `json:"bank_account" binding:"omitempty,max=64"`
	Address     *string `json:"address" binding:"omitempty,max=512"`
	Phone       *string `json:"phone" binding:"omitempty,max=32"`
	IsDefault   bool    `json:"is_default"`
}
