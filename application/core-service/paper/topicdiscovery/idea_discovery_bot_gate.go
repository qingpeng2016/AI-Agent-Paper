package topicdiscovery

import (
	"context"

	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

// PrepareIdeaDiscoveryBotInvoke：run_state 该段未 done，且 runtime_state 里该段为 true（用户已确认）→ 允许打一条 aris。
func (s *TopicDiscoveryRunService) PrepareIdeaDiscoveryBotInvoke(
	ctx context.Context,
	userID uint,
	runVersion int,
	step *entity.PaperOutputTopicStep,
	input topicRunInput,
) (invoke bool, err error) {
	if step == nil {
		return false, nil
	}
	meta := paperstep.LoadStepExtra(step)
	if v, ok := meta["async_llm"].(bool); !ok || !v {
		return false, nil
	}

	_ = s.ReconcileIdeaDiscoveryBotAwaiting(ctx, userID, runVersion, step, input)
	meta = paperstep.LoadStepExtra(step)
	normalizeIdeaDiscoveryRuntimeMeta(meta)

	cfg := mustIdeaDiscoveryConfig()
	paperRoot := ARISRunPaperRoot(uint64(userID), step.ManuscriptID, runVersion)
	runID := resolveIdeaDiscoveryRunID(paperRoot, step, input)
	_ = seedIdeaDiscoveryRunStateFromWorkspace(paperRoot, runID)

	pending, err := ideaDiscoveryResumePhase(paperRoot, runID)
	if err != nil {
		return false, err
	}
	syncIdeaDiscoveryRunProgressToMeta(meta, paperRoot, runID)

	if pending == "" {
		paperstep.PersistStepExtra(step, meta)
		_ = s.steps.SaveStep(ctx, step)
		return false, nil
	}

	if !IdeaDiscoveryBotMayInvokePending(meta, pending) {
		paperstep.PersistStepExtra(step, meta)
		_ = s.steps.SaveStep(ctx, step)
		if ideaDiscoveryNeedsUserConfirm(meta, pending, cfg) {
			return false, s.saveIdeaDiscoveryAwaitingUserContinue(ctx, step, paperRoot, runID, pending)
		}
		delete(meta, "async_llm")
		paperstep.PersistStepExtra(step, meta)
		_ = s.steps.SaveStep(ctx, step)
		return false, nil
	}

	paperstep.PersistStepExtra(step, meta)
	_ = s.steps.SaveStep(ctx, step)
	return true, nil
}
