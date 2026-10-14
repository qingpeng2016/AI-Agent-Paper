package response

type PaperManuscriptFigureIDsView struct {
	ManuscriptID string   `json:"manuscript_id"`
	FigureIDs    []string `json:"figure_ids"`
}

type PaperManuscriptWithFiguresItemView struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	IsCurrent    bool     `json:"is_current"`
	FigureIDs    []string `json:"figure_ids,omitempty"`
	FigureTitles []string `json:"figure_titles,omitempty"`
	CreatedAt    string   `json:"created_at,omitempty"`
}

type PaperManuscriptWithFiguresListView struct {
	Items    []PaperManuscriptWithFiguresItemView `json:"items"`
	Total    int64                                `json:"total,omitempty"`
	Page     int                                  `json:"page,omitempty"`
	PageSize int                                  `json:"page_size,omitempty"`
}
