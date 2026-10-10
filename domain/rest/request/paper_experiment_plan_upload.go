package request

type PaperExperimentPlanUploadForm struct {
	ManuscriptID     uint64 `form:"manuscript_id" binding:"required"`
	ExperimentPlanID uint64 `form:"experiment_plan_id" binding:"required"`
}
