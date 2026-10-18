package topicdiscovery

import (
	"context"
	"strings"
	"time"

	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

func (s *TopicDiscoveryRunService) stageBeginPhasedTopicStageAsync(
	ctx context.Context,
	userID uint,
	runVersion int,
	step *entity.PaperOutputTopicStep,
) error {
	if step == nil {
		return errorx.ErrTopicStepInvalid
	}
	if len(step.Result) > 0 && strings.TrimSpace(step.Status) == "completed" {
		return nil
	}
	phaseID := topicStageCodeToPhaseID(step.StageCode)

	step.Status = "running"
	step.CompletedAt = nil
	step.ExecProgressPct = 1
	meta := paperstep.LoadStepExtra(step)
	meta["async_llm"] = true
	meta["async_llm_done"] = false
	meta["idea_discovery_invoke_mode"] = "phased_per_step"
	meta["phase_id"] = phaseID
	delete(meta, "awaiting_user_confirm")
	delete(meta, "exec_progress_completed_calls")
	delete(meta, "exec_progress_planned_calls")
	paperstep.PersistStepExtra(step, meta)
	return s.steps.SaveStep(ctx, step)
}

func (s *TopicDiscoveryRunService) beginPhasedTopicStageAsync(ctx context.Context, userID uint, runVersion int, stageCode string) error {
	step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, stageCode)
	if err != nil || step == nil {
		return errorx.ErrTopicStepInvalid
	}
	if strings.TrimSpace(step.Status) == "completed" {
		return nil
	}
	now := time.Now()
	step.Status = "running"
	if step.StartedAt == nil {
		step.StartedAt = &now
	}
	return s.stageBeginPhasedTopicStageAsync(ctx, userID, runVersion, step)
}
