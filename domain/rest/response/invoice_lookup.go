package response

type EnterpriseInvoiceLookupItem struct {
	Title       string `json:"title"`
	TaxNo       string `json:"tax_no"`
	BankName    string `json:"bank_name,omitempty"`
	BankAccount string `json:"bank_account,omitempty"`
	Address     string `json:"address,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

type EnterpriseInvoiceLookupResult struct {
	Match      EnterpriseInvoiceLookupItem   `json:"match"`
	Candidates []EnterpriseInvoiceLookupItem `json:"candidates,omitempty"`
}
