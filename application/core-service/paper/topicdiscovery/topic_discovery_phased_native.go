package topicdiscovery

import (
	"context"
	"fmt"
	"strings"

	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

// RunNativeIdeaDiscoveryPhaseLLM 分阶段 bot：每 tick 对 runtime_state=executing 的段打一条 /{phase} ARIS，不跑 bundled /idea-discovery。
func (s *TopicDiscoveryRunService) RunNativeIdeaDiscoveryPhaseLLM(
	ctx context.Context,
	userID uint,
	runVersion int,
	step *entity.PaperOutputTopicStep,
	input topicRunInput,
) error {
	cfg := mustIdeaDiscoveryConfig()

	route, err := s.requireTopicStrategyRoute(ctx, BehaviorIdeaDiscoveryRun)
	if err != nil {
		return err
	}

	paperRoot := ARISRunPaperRoot(uint64(userID), step.ManuscriptID, runVersion)
	invokeCtx := s.buildTopicStrategyInvokeContext(ctx, userID, runVersion, step, input, 1, 1, nil, nil)

	meta := paperstep.LoadStepExtra(step)
	normalizeIdeaDiscoveryRuntimeMeta(meta)
	phase := executingIdeaDiscoveryPhaseFromMeta(meta)
	if phase == "" {
		return fmt.Errorf("phased invoke: 无 executing 段")
	}

	runID := resolveIdeaDiscoveryRunID(paperRoot, step, input)
	meta["aris_run_id"] = runID
	meta["native_run_id"] = runID
	syncIdeaDiscoveryPhasedProgressMeta(meta)
	paperstep.PersistStepExtra(step, meta)
	step.ExecProgressPct = ideaDiscoveryPhasedProgressPct(meta)
	_ = s.steps.SaveStep(ctx, step)

	cfg.ClearCLIWorkspaceSessions(paperRoot)
	locale := invokeCtx.Language
	if err := PrepareNativeIdeaDiscoveryWorkspace(paperRoot, input, locale, cfg); err != nil {
		return err
	}
	argsLine := cfg.FormatIdeaDiscoveryArgsLine(input, strings.TrimSpace(input.Venue), locale)
	if err := writeNativePlatformArguments(paperRoot, argsLine); err != nil {
		return err
	}

	req := BuildTopicDiscoveryInvokeRequest(BehaviorIdeaDiscoveryPhase, phase, input, invokeCtx)
	raw, usage, err := s.invokeLocalARIS(ctx, step, req, route, "")
	if err != nil {
		return s.handleIdeaDiscoveryInvokeFailure(ctx, step, paperRoot, cfg, err)
	}

	meta = paperstep.LoadStepExtra(step)
	meta["aris_invoke_count"] = ideaDiscoveryInvokeCount(meta) + 1
	meta["last_phased_invoke"] = phase
	meta["last_invoke_preview"] = truncateUTF8(string(raw), 4096)
	meta["tokens_prompt"] = usage.PromptTokens
	meta["tokens_completion"] = usage.CompletionTokens
	delete(meta, "aris_invoke_last_error")
	delete(meta, "aris_invoke_retries")
	syncIdeaDiscoveryPhasedProgressMeta(meta)
	paperstep.PersistStepExtra(step, meta)

	if phase == "research-lit" && (ideaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, phase, cfg) || isResearchLitPhaseDone(paperRoot, runID)) {
		_, _ = s.ingestARISWorkspaceLiterature(ctx, userID, runVersion, step.ManuscriptID, paperRoot, cfg)
	}

	return s.advanceIdeaDiscoveryPhasedAfterInvoke(ctx, step, userID, runVersion, route, paperRoot, runID, phase, cfg)
}

func (s *TopicDiscoveryRunService) advanceIdeaDiscoveryPhasedAfterInvoke(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	userID uint,
	runVersion int,
	route *entity.PaperStrategyRoute,
	paperRoot, runID, phase string,
	cfg ARISIdeaDiscoveryConfig,
) error {
	meta := paperstep.LoadStepExtra(step)
	normalizeIdeaDiscoveryRuntimeMeta(meta)
	syncIdeaDiscoveryPhasedProgressMeta(meta)

	if !ideaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, phase, cfg) {
		if IdeaDiscoveryPhasedBotMayRetryPhase(cfg) {
			meta["async_llm"] = true
			meta["async_llm_done"] = false
			paperstep.PersistStepExtra(step, meta)
			step.Status = "running"
			step.ExecProgressPct = ideaDiscoveryPhasedProgressPct(meta)
			return s.steps.SaveStep(ctx, step)
		}
		delete(meta, "async_llm")
		meta["async_llm_done"] = false
		meta["aris_pending_phase"] = phase
		paperstep.PersistStepExtra(step, meta)
		step.Status = "running"
		step.ExecProgressPct = ideaDiscoveryPhasedProgressPct(meta)
		return s.steps.SaveStep(ctx, step)
	}

	setPhaseStatusInMeta(meta, phase, IdeaDiscoveryPhaseCompleted)
	meta["aris_pending_phase"] = phase

	next := nextIdeaDiscoveryPhaseID(phase)
	if next == "" {
		delete(meta, "async_llm")
		paperstep.PersistStepExtra(step, meta)
		return s.finishIdeaDiscoveryAfterGate(ctx, step, userID, runVersion, route, paperRoot, runID, cfg)
	}

	if cfg.PhaseNeedsUserCheckpoint(phase) {
		setPhaseStatusInMeta(meta, next, IdeaDiscoveryPhaseAwaitingConfirm)
		meta["aris_pending_phase"] = next
		delete(meta, "async_llm")
		meta["async_llm_done"] = false
	} else {
		setPhaseStatusInMeta(meta, next, IdeaDiscoveryPhaseExecuting)
		meta["aris_pending_phase"] = next
		meta["async_llm"] = true
		meta["async_llm_done"] = false
	}

	syncIdeaDiscoveryPhasedProgressMeta(meta)
	paperstep.PersistStepExtra(step, meta)
	step.Status = "running"
	step.ExecProgressPct = ideaDiscoveryPhasedProgressPct(meta)
	return s.steps.SaveStep(ctx, step)
}
