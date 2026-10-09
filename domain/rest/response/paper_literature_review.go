package response

type PaperLiteratureReviewItemView struct {
	ID                         string `json:"id"`
	Version                    int    `json:"version"`
	Status                     string `json:"status"`
	ExperimentPlanID string `json:"experiment_plan_id,omitempty"`
	Structure     string `json:"structure,omitempty"`
	Title         string `json:"title,omitempty"`
	Summary       string `json:"summary,omitempty"`
	ContentMedium string `json:"content_medium,omitempty"`
	Format        string `json:"format"`
	CreatedAt     string `json:"created_at"`
}

type PaperLiteratureReviewListView struct {
	ManuscriptID uint64                        `json:"manuscript_id"`
	Items        []PaperLiteratureReviewItemView `json:"items"`
}
