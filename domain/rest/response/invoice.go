package response

type UserInvoiceListPageResp struct {
	Items    []UserInvoiceListItem `json:"items"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

type UserInvoiceListItem struct {
	ID        uint   `json:"id"`
	OrderNo   string `json:"order_no"`
	Title     string `json:"title"`
	Amount    Money  `json:"amount"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}
