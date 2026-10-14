package request

type PaperManuscriptFigureIDsQuery struct {
	ManuscriptID string `form:"manuscript_id"`
}

type PaperManuscriptSetFigureMembershipRequest struct {
	ManuscriptID string   `json:"manuscript_id"`
	FigureIDs    []string `json:"figure_ids"`
}

type PaperManuscriptWithFiguresListQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
