package request

type PaperLiteratureReviewListQuery struct {
	ManuscriptID uint64 `form:"manuscript_id" binding:"required"`
}
