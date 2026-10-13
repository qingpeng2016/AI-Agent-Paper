package response

type PaperManuscriptLiteratureIDsView struct {
	ManuscriptID  string   `json:"manuscript_id"`
	LiteratureIDs []string `json:"literature_ids"`
}

type PaperManuscriptWithLiteratureItemView struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	IsCurrent        bool     `json:"is_current"`
	LiteratureIDs    []string `json:"literature_ids,omitempty"`
	LiteratureTitles []string `json:"literature_titles,omitempty"`
	CreatedAt        string   `json:"created_at,omitempty"`
}

type PaperManuscriptWithLiteratureListView struct {
	Items    []PaperManuscriptWithLiteratureItemView `json:"items"`
	Total    int64                                   `json:"total,omitempty"`
	Page     int                                     `json:"page,omitempty"`
	PageSize int                                     `json:"page_size,omitempty"`
}
