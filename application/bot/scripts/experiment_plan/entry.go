package experimentplan

import (
	"context"

	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"go.uber.org/zap"
)

const (
	ModuleAIAgentPaper      = "ai_agent_paper"
	TaskExperimentPlanLLM   = "experiment_plan_llm"
	scanLimit               = 20
)

type ExperimentPlanJob struct {
	reviews repository.PaperOutputLiteratureReviewRepo
	plans   *papersvc.ExperimentPlanService
}

func NewExperimentPlanJob(
	reviews repository.PaperOutputLiteratureReviewRepo,
	plans *papersvc.ExperimentPlanService,
) *ExperimentPlanJob {
	return &ExperimentPlanJob{reviews: reviews, plans: plans}
}

func (j *ExperimentPlanJob) Run(ctx context.Context) {
	rows, err := j.reviews.ListByStatus(ctx, papersvc.LitReviewStatusGeneratingExperimentPlan, scanLimit)
	if err != nil {
		logger.ErrorZ(ctx, "experiment-plan-list-failed", zap.Error(err))
		return
	}
	if len(rows) == 0 {
		return
	}
	var okN, failN int
	for i := range rows {
		if err := j.plans.RunExperimentPlanLLM(ctx, &rows[i]); err != nil {
			failN++
			logger.WarnZ(ctx, "experiment-plan-llm-failed",
				zap.Uint64("review_id", rows[i].ID),
				zap.Error(err))
			continue
		}
		okN++
	}
	logger.InfoZ(ctx, "experiment-plan-finished",
		zap.Int("scanned", len(rows)),
		zap.Int("completed", okN),
		zap.Int("failed", failN))
}
