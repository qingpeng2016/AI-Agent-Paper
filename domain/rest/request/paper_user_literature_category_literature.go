package request

type PaperUserLiteratureCategoryLiteratureIDsQuery struct {
	CategoryID string `form:"category_id"`
}

type PaperUserLiteratureCategorySetLiteratureRequest struct {
	CategoryID    string   `json:"category_id"`
	LiteratureIDs []string `json:"literature_ids"`
}
