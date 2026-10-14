package topicdiscovery

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/internal/jsonx"
	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

func topicDiscoveryPlannedCalls(maxRounds int) int {
	if maxRounds < 1 {
		maxRounds = 1
	}
	return maxRounds * 2
}

func topicDiscoveryProgressPct(completedCalls, plannedTotal int, beforeFirstInvoke bool) uint8 {
	if plannedTotal <= 0 {
		return 0
	}
	if beforeFirstInvoke && completedCalls == 0 {
		return 1
	}
	if completedCalls >= plannedTotal {
		return 100
	}
	pct := completedCalls * 100 / plannedTotal
	if pct < 1 && completedCalls > 0 {
		return 1
	}
	if pct > 99 && completedCalls < plannedTotal {
		return 99
	}
	return uint8(pct)
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
	step.ExecProgressPct = topicDiscoveryProgressPct(completedCalls, plannedTotal, beforeFirstInvoke)
	meta := paperstep.LoadStepExtra(step)
	meta["exec_progress_planned_calls"] = plannedTotal
	meta["exec_progress_completed_calls"] = completedCalls
	step.Extra = jsonx.MustJSON(meta)
	return s.steps.SaveStep(ctx, step)
}
