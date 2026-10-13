package request

type PaperLiteratureReviewListQuery struct {
	ManuscriptID uint64 `form:"manuscript_id" binding:"required"`
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
}
