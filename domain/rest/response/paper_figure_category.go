package response

type PaperFigureCategoryItemView struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at,omitempty"`
}

type PaperFigureCategoryListView struct {
	Items    []PaperFigureCategoryItemView `json:"items"`
	Total    int64                         `json:"total,omitempty"`
	Page     int                           `json:"page,omitempty"`
	PageSize int                           `json:"page_size,omitempty"`
}
