package topicdiscovery

import (
	"context"
	"fmt"
	"strings"
	"time"

	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

// PrepareTopicStageBotInvoke 分 step bot：当前 stage running 且 async_llm，前序 step 均已 completed。
func (s *TopicDiscoveryRunService) PrepareTopicStageBotInvoke(
	ctx context.Context,
	userID uint,
	runVersion int,
	step *entity.PaperOutputTopicStep,
	input topicRunInput,
) (invoke bool, err error) {
	if step == nil || !isPhasedTopicStage(step.StageCode) {
		return false, nil
	}
	meta := paperstep.LoadStepExtra(step)
	if v, ok := meta["async_llm"].(bool); !ok || !v {
		return false, nil
	}
	if strings.TrimSpace(step.Status) != "running" {
		delete(meta, "async_llm")
		paperstep.PersistStepExtra(step, meta)
		_ = s.steps.SaveStep(ctx, step)
		return false, nil
	}
	ok, err := s.phasedStagesPriorCompleted(ctx, userID, runVersion, step.StageCode)
	if err != nil {
		return false, err
	}
	if !ok {
		delete(meta, "async_llm")
		paperstep.PersistStepExtra(step, meta)
		_ = s.steps.SaveStep(ctx, step)
		return false, nil
	}
	if awaiting, _ := meta["awaiting_user_confirm"].(bool); awaiting {
		return false, nil
	}
	return true, nil
}

// RunNativeTopicStageLLM 对单个 paper_output_topic_step 打一条 /research-lit 等 ARIS 命令。
func (s *TopicDiscoveryRunService) RunNativeTopicStageLLM(
	ctx context.Context,
	userID uint,
	runVersion int,
	step *entity.PaperOutputTopicStep,
	input topicRunInput,
) error {
	if step == nil || !isPhasedTopicStage(step.StageCode) {
		return fmt.Errorf("非分阶段 topic step: %q", step.StageCode)
	}
	cfg := mustIdeaDiscoveryConfig()
	stageCode := strings.TrimSpace(step.StageCode)
	phaseID := topicStageCodeToPhaseID(stageCode)
	behaviorKey := topicStageBehaviorKey(stageCode)

	route, err := s.requireTopicStrategyRoute(ctx, behaviorKey)
	if err != nil {
		return err
	}

	paperRoot := ARISRunPaperRoot(uint64(userID), step.ManuscriptID, runVersion)
	invokeCtx := s.buildTopicStrategyInvokeContext(ctx, userID, runVersion, step, input, 1, 1, nil, nil)

	meta := paperstep.LoadStepExtra(step)
	runID := resolveIdeaDiscoveryRunID(paperRoot, step, input)
	meta["aris_run_id"] = runID
	meta["native_run_id"] = runID
	meta["idea_discovery_invoke_mode"] = "phased_per_step"
	meta["phase_id"] = phaseID
	paperstep.PersistStepExtra(step, meta)
	_ = s.steps.SaveStep(ctx, step)

	cfg.ClearCLIWorkspaceSessions(paperRoot)
	if err := PrepareNativeIdeaDiscoveryWorkspace(paperRoot, input, invokeCtx.Language, cfg); err != nil {
		return err
	}
	argsLine := cfg.FormatIdeaDiscoveryArgsLine(input, strings.TrimSpace(input.Venue), invokeCtx.Language)
	if err := writeNativePlatformArguments(paperRoot, argsLine); err != nil {
		return err
	}

	req := BuildTopicDiscoveryInvokeRequest(behaviorKey, route.CapabilityID, input, invokeCtx)
	raw, usage, err := s.invokeTopicDiscoveryBehavior(ctx, step, req, "")
	if err != nil {
		return s.handleIdeaDiscoveryInvokeFailure(ctx, step, paperRoot, cfg, err)
	}

	meta = paperstep.LoadStepExtra(step)
	meta["aris_invoke_count"] = ideaDiscoveryInvokeCount(meta) + 1
	meta["last_phased_invoke"] = phaseID
	meta["last_invoke_preview"] = truncateUTF8(string(raw), 4096)
	meta["tokens_prompt"] = usage.PromptTokens
	meta["tokens_completion"] = usage.CompletionTokens
	delete(meta, "aris_invoke_last_error")
	delete(meta, "aris_invoke_retries")
	paperstep.PersistStepExtra(step, meta)

	if phaseID == "research-lit" && ideaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, phaseID, cfg) {
		_, _ = s.ingestARISWorkspaceLiterature(ctx, userID, runVersion, step.ManuscriptID, paperRoot, cfg)
	}

	return s.advanceTopicStageAfterInvoke(ctx, step, userID, runVersion, route, paperRoot, runID, stageCode, phaseID, cfg)
}

func (s *TopicDiscoveryRunService) advanceTopicStageAfterInvoke(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	userID uint,
	runVersion int,
	route *entity.PaperStrategyRoute,
	paperRoot, runID, stageCode, phaseID string,
	cfg ARISIdeaDiscoveryConfig,
) error {
	meta := paperstep.LoadStepExtra(step)

	if !ideaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, phaseID, cfg) {
		if IdeaDiscoveryPhasedBotMayRetryPhase(cfg) {
			meta["async_llm"] = true
			paperstep.PersistStepExtra(step, meta)
			step.Status = "running"
			return s.steps.SaveStep(ctx, step)
		}
		reason := fmt.Sprintf("Idea Discovery 阶段 /%s 未产出约定产物", phaseID)
		return s.persistIdeaDiscoveryPhasedStepFailed(ctx, step, reason)
	}

	delete(meta, "async_llm")
	delete(meta, "awaiting_user_confirm")
	meta["async_llm_done"] = true
	meta["aris_pending_phase"] = phaseID
	paperstep.PersistStepExtra(step, meta)

	nextCode := nextPhasedTopicStageCode(stageCode)
	if nextCode == "topic_confirm" || nextCode == "" {
		return s.finishIdeaDiscoveryAfterGate(ctx, step, userID, runVersion, route, paperRoot, runID, cfg)
	}

	if err := s.completeTopicStageStep(ctx, step); err != nil {
		return err
	}

	if cfg.PhaseNeedsUserCheckpoint(phaseID) {
		return nil
	}
	return s.beginPhasedTopicStageAsync(ctx, userID, runVersion, nextCode)
}

func (s *TopicDiscoveryRunService) completeTopicStageStep(ctx context.Context, step *entity.PaperOutputTopicStep) error {
	if step == nil {
		return nil
	}
	now := time.Now()
	step.Status = "completed"
	step.ExecProgressPct = 100
	step.CompletedAt = &now
	return s.steps.SaveStep(ctx, step)
}
