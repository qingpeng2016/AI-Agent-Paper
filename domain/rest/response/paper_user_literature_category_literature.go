package response

type PaperUserLiteratureCategoryLiteratureIDsView struct {
	CategoryID    string   `json:"category_id"`
	LiteratureIDs []string `json:"literature_ids"`
}
