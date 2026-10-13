package response

type PaperManuscriptItemView struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Status          string `json:"status"`
	IsCurrent       bool   `json:"is_current"`
	ContentLanguage string `json:"content_language"`
	DisciplineCode  string `json:"discipline_code,omitempty"`
	DisciplineLabel string `json:"discipline_label,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
}

type PaperManuscriptListView struct {
	CurrentManuscriptID uint64                    `json:"current_manuscript_id,omitempty"`
	Items               []PaperManuscriptItemView `json:"items"`
	Total               int64                     `json:"total,omitempty"`
	Page                int                       `json:"page,omitempty"`
	PageSize            int                       `json:"page_size,omitempty"`
}
