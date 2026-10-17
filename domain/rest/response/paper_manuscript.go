package response

type PaperManuscriptItemView struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Status          string `json:"status"`
	IsCurrent       bool   `json:"is_current"`
	Language string `json:"language"`
	DisciplineCode  string `json:"discipline_code,omitempty"`
	DisciplineLabel string `json:"discipline_label,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	DraftVersion    uint   `json:"draft_version,omitempty"`
	DraftStatus     string `json:"draft_status,omitempty"`
	SubmissionStatus string `json:"submission_status,omitempty"`
	SubmissionLabel  string `json:"submission_label,omitempty"`
	SubmissionDeadline string `json:"submission_deadline,omitempty"`
}

type PaperManuscriptListView struct {
	CurrentManuscriptID uint64                    `json:"current_manuscript_id,omitempty"`
	Items               []PaperManuscriptItemView `json:"items"`
	Total               int64                     `json:"total,omitempty"`
	Page                int                       `json:"page,omitempty"`
	PageSize            int                       `json:"page_size,omitempty"`
}
