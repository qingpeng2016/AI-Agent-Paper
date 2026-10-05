package response

type InvoiceConfigItem struct {
	ID          uint   `json:"id"`
	ProfileType string `json:"profile_type"`
	Title       string `json:"title"`
	TaxNo       string `json:"tax_no,omitempty"`
	BankName    string `json:"bank_name,omitempty"`
	BankAccount string `json:"bank_account,omitempty"`
	Address     string `json:"address,omitempty"`
	Phone       string `json:"phone,omitempty"`
	IsDefault   bool   `json:"is_default"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
