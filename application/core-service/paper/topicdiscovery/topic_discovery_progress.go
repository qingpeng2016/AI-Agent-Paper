package topicdiscovery

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/internal/jsonx"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/literature"
	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

func topicDiscoveryPlannedCalls(maxRounds int) int {
	if maxRounds < 1 {
		maxRounds = 1
	}
	return maxRounds * 2
}

func (s *TopicDiscoveryRunService) saveTopicDiscoveryExecProgress(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	completedCalls, plannedTotal int,
	beforeFirstInvoke bool,
) error {
	if s == nil || s.steps == nil || step == nil {
		return nil
	}
	step.ExecProgressPct = literature.ExecProgressPct(completedCalls, plannedTotal, beforeFirstInvoke)
	meta := paperstep.LoadStepExtra(step)
	meta["exec_progress_planned_calls"] = plannedTotal
	meta["exec_progress_completed_calls"] = completedCalls
	step.Extra = jsonx.MustJSON(meta)
	return s.steps.SaveStep(ctx, step)
}
