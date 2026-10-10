package response

type PaperExperimentPlanItemView struct {
	ID                 string `json:"id"`
	Version            int    `json:"version"`
	Status             string `json:"status"`
	Title              string `json:"title,omitempty"`
	Summary            string `json:"summary,omitempty"`
	ContentMedium      string `json:"content_medium,omitempty"`
	Format             string `json:"format"`
	LiteratureReviewID string `json:"literature_review_id,omitempty"`
	ExperimentDataURI  string `json:"experiment_data_uri,omitempty"`
	CreatedAt          string `json:"created_at"`
}

type PaperExperimentPlanListView struct {
	ManuscriptID uint64                        `json:"manuscript_id"`
	Items        []PaperExperimentPlanItemView `json:"items"`
}
