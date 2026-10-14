package response

type PaperFigureCategoryFigureIDsView struct {
	CategoryID string   `json:"category_id"`
	FigureIDs  []string `json:"figure_ids"`
}
