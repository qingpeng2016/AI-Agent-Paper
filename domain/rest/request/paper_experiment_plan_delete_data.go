package request

type PaperExperimentPlanDeleteExperimentDataBody struct {
	ManuscriptID     uint64 `json:"manuscript_id" binding:"required"`
	ExperimentPlanID uint64 `json:"experiment_plan_id" binding:"required"`
}
