package request

type PaperExperimentPlanListQuery struct {
	ManuscriptID uint64 `form:"manuscript_id" binding:"required"`
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
}

type PaperLiteratureReviewDetailQuery struct {
	ManuscriptID       uint64 `form:"manuscript_id" binding:"required"`
	LiteratureReviewID uint64 `form:"literature_review_id" binding:"required"`
}

type PaperExperimentPlanDetailQuery struct {
	ManuscriptID     uint64 `form:"manuscript_id" binding:"required"`
	ExperimentPlanID uint64 `form:"experiment_plan_id" binding:"required"`
}
