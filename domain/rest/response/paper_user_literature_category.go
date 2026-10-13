package response

type PaperUserLiteratureCategoryItemView struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Sort      int    `json:"sort"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at,omitempty"`
}

type PaperUserLiteratureCategoryListView struct {
	Items    []PaperUserLiteratureCategoryItemView `json:"items"`
	Total    int64                                 `json:"total,omitempty"`
	Page     int                                   `json:"page,omitempty"`
	PageSize int                                   `json:"page_size,omitempty"`
}
