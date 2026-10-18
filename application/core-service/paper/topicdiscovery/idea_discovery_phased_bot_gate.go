package topicdiscovery

import (
	"context"
	"fmt"
	"strings"

	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

// PrepareIdeaDiscoveryPhasedBotInvoke：按 DB runtime_state 找 executing 段；不依赖 .aris/runs JSON 决定是否 invoke。
func (s *TopicDiscoveryRunService) PrepareIdeaDiscoveryPhasedBotInvoke(
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
	syncIdeaDiscoveryPhasedProgressMeta(meta)

	cfg := mustIdeaDiscoveryConfig()
	phase := executingIdeaDiscoveryPhaseFromMeta(meta)
	if phase == "" {
		delete(meta, "async_llm")
		paperstep.PersistStepExtra(step, meta)
		_ = s.steps.SaveStep(ctx, step)
		return false, nil
	}

	paperRoot := ARISRunPaperRoot(uint64(userID), step.ManuscriptID, runVersion)
	runID := resolveIdeaDiscoveryRunID(paperRoot, step, input)
	meta["aris_run_id"] = runID
	meta["native_run_id"] = runID

	if ideaDiscoveryNeedsUserConfirm(meta, phase, cfg) {
		paperstep.PersistStepExtra(step, meta)
		_ = s.steps.SaveStep(ctx, step)
		return false, s.saveIdeaDiscoveryAwaitingUserContinue(ctx, step, paperRoot, runID, phase)
	}

	if phaseStatusFromRuntimeMeta(meta, phase) != IdeaDiscoveryPhaseExecuting {
		delete(meta, "async_llm")
		paperstep.PersistStepExtra(step, meta)
		_ = s.steps.SaveStep(ctx, step)
		return false, nil
	}

	if ideaDiscoveryPhasedInvokeCount(meta, phase) >= 1 {
		if strings.TrimSpace(phase) == "research-lit" && researchLitLandscapePresent(paperRoot) {
			if err := s.advanceResearchLitPhasedWhenDeliverablePresent(
				ctx, step, userID, runVersion, paperRoot, runID, phase, cfg,
			); err != nil {
				return false, err
			}
			return false, nil
		}
		reason := ideaDiscoveryPhasedReinvokeFailureReason(paperRoot, phase)
		if err := s.persistIdeaDiscoveryPhasedStepFailed(ctx, step, reason); err != nil {
			return false, err
		}
		return false, fmt.Errorf("%s", reason)
	}

	paperstep.PersistStepExtra(step, meta)
	step.ExecProgressPct = ideaDiscoveryPhasedProgressPct(meta)
	_ = s.steps.SaveStep(ctx, step)
	return true, nil
}
