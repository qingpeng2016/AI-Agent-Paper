package request

type PaperFigureCategoryFigureIDsQuery struct {
	CategoryID string `form:"category_id"`
}

type PaperFigureCategorySetFigureRequest struct {
	CategoryID string   `json:"category_id"`
	FigureIDs  []string `json:"figure_ids"`
}
