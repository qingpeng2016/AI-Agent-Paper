package request

type PaperLiteratureReviewSoftDeleteBody struct {
	ManuscriptID       uint64 `json:"manuscript_id" binding:"required"`
	LiteratureReviewID uint64 `json:"literature_review_id" binding:"required"`
}
