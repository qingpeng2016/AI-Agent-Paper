package response

type PaperManuscriptItemView struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Status          string `json:"status"`
	IsCurrent       bool   `json:"is_current"`
	DisciplineCode  string `json:"discipline_code,omitempty"`
	DisciplineLabel string `json:"discipline_label,omitempty"`
}

type PaperManuscriptListView struct {
	CurrentManuscriptID uint64                    `json:"current_manuscript_id,omitempty"`
	Items               []PaperManuscriptItemView `json:"items"`
}
