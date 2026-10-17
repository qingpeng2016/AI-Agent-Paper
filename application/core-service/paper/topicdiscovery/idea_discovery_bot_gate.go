package topicdiscovery

import (
	"context"

	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

// PrepareIdeaDiscoveryBotInvoke：只读 ARIS 的 run_state JSON；无文件或 JSON 该 pending 未 done 且 DB 为 executing → 本 tick 打一条 aris。
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

	if !IdeaDiscoveryBotMayInvokePending(meta, pending, paperRoot, runID) {
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
