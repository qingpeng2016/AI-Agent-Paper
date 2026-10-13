package response

type PaperLiteratureReviewItemView struct {
	ID                         string `json:"id"`
	Version                    int    `json:"version"`
	Status                     string `json:"status"`
	From                       string `json:"from"`
	ExperimentPlanID string `json:"experiment_plan_id,omitempty"`
	Title            string `json:"title,omitempty"`
	ContentMedium    string `json:"content_medium,omitempty"`
	PdfURI        string `json:"pdf_uri,omitempty"`
	WordURI       string `json:"word_uri,omitempty"`
	PptURI        string `json:"ppt_uri,omitempty"`
	Format        string `json:"format"`
	CreatedAt     string `json:"created_at"`
}

type PaperLiteratureReviewListView struct {
	ManuscriptID uint64                          `json:"manuscript_id"`
	Items        []PaperLiteratureReviewItemView `json:"items"`
	Total        int64                           `json:"total,omitempty"`
	Page         int                             `json:"page,omitempty"`
	PageSize     int                             `json:"page_size,omitempty"`
}
