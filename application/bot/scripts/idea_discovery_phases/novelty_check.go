package ideadiscoveryphases

import (
	"context"
	"fmt"
	"strings"
	"time"

	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/topicdiscovery"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"go.uber.org/zap"
)

// RunNoveltyCheck 处理 entry 传入的 novelty_check step：组装 /novelty-check 并 invoke ARIS。
func (j *IdeaDiscoveryPhasesJob) RunNoveltyCheck(ctx context.Context, step *entity.PaperOutputTopicStep) (scanned, invoked, skipped, failed int) {
	if step == nil || strings.TrimSpace(step.Status) != topicStepRunning ||
		strings.TrimSpace(step.StageCode) != noveltyCheckStageCode {
		return 0, 0, 0, 0
	}
	meta := paperstep.LoadStepExtra(step)
	async, ok := meta["async_llm"].(bool)
	if !ok || !async {
		return 0, 0, 0, 0
	}
	scanned = 1
	switch j.processNoveltyCheckStep(ctx, step) {
	case topicStepCompleted:
		invoked++
	case topicStepFailed:
		failed++
	default:
		skipped++
	}
	return scanned, invoked, skipped, failed
}

// noveltyCheckPriorCompleted 前序 research_lit、idea_creator 是否均已 completed。
func (j *IdeaDiscoveryPhasesJob) noveltyCheckPriorCompleted(ctx context.Context, userID uint, runVersion int) (bool, error) {
	for _, code := range noveltyCheckPriorStages {
		step, err := j.steps.GetStep(ctx, uint64(userID), runVersion, code)
		if err != nil {
			return false, err
		}
		if step == nil || strings.TrimSpace(step.Status) != topicStepCompleted {
			return false, nil
		}
	}
	return true, nil
}

// noveltyCheckPrepareInvoke 本 tick 是否允许调 ARIS（async_llm、前序、状态、非人工暂停）。
func (j *IdeaDiscoveryPhasesJob) noveltyCheckPrepareInvoke(ctx context.Context, userID uint, runVersion int, step *entity.PaperOutputTopicStep) (bool, error) {
	meta := paperstep.LoadStepExtra(step)
	if v, ok := meta["async_llm"].(bool); !ok || !v {
		return false, nil
	}
	if strings.TrimSpace(step.Status) != topicStepRunning {
		delete(meta, "async_llm")
		paperstep.PersistStepExtra(step, meta)
		_ = j.steps.SaveStep(ctx, step)
		return false, nil
	}
	ok, err := j.noveltyCheckPriorCompleted(ctx, userID, runVersion)
	if err != nil {
		return false, err
	}
	if !ok {
		delete(meta, "async_llm")
		paperstep.PersistStepExtra(step, meta)
		_ = j.steps.SaveStep(ctx, step)
		return false, nil
	}
	if awaiting, _ := meta["awaiting_user_confirm"].(bool); awaiting {
		return false, nil
	}
	return true, nil
}

// processNoveltyCheckStep 处理单条 novelty_check：invoke、通知，返回 completed/failed/skipped。
func (j *IdeaDiscoveryPhasesJob) processNoveltyCheckStep(ctx context.Context, step *entity.PaperOutputTopicStep) string {
	if step == nil || j.runs == nil {
		return topicStepSkipped
	}
	if strings.TrimSpace(step.Status) == topicStepFailed {
		return topicStepSkipped
	}
	if step.ErrorMsg != nil && strings.TrimSpace(*step.ErrorMsg) != "" {
		if strings.TrimSpace(step.Status) != topicStepFailed {
			papersvc.FailTopicStep(step, *step.ErrorMsg)
			_ = j.steps.SaveStep(ctx, step)
		}
		return topicStepSkipped
	}
	meta := paperstep.LoadStepExtra(step)
	if async, ok := meta["async_llm"].(bool); !ok || !async {
		return topicStepSkipped
	}
	userID := uint(step.UserID)
	runVersion := step.RunVersion
	invoke, err := j.noveltyCheckPrepareInvoke(ctx, userID, runVersion, step)
	if err != nil {
		noveltyCheckMarkFailed(j, ctx, step, err.Error(), 0, 0)
		return topicStepFailed
	}
	if !invoke {
		return topicStepSkipped
	}
	meta = paperstep.LoadStepExtra(step)
	beforePrompt := noveltyCheckIntMeta(meta["tokens_prompt"])
	beforeCompletion := noveltyCheckIntMeta(meta["tokens_completion"])
	if err = j.runNoveltyCheckTick(ctx, userID, runVersion, step); err != nil {
		if strings.TrimSpace(step.Status) != topicStepFailed {
			noveltyCheckMarkFailed(j, ctx, step, err.Error(), 0, 0)
		} else {
			meta = paperstep.LoadStepExtra(step)
			noveltyCheckNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed, err.Error(),
				noveltyCheckIntMeta(meta["tokens_prompt"])-beforePrompt, noveltyCheckIntMeta(meta["tokens_completion"])-beforeCompletion)
		}
		return topicStepFailed
	}
	meta = paperstep.LoadStepExtra(step)
	deltaPrompt := noveltyCheckIntMeta(meta["tokens_prompt"]) - beforePrompt
	deltaCompletion := noveltyCheckIntMeta(meta["tokens_completion"]) - beforeCompletion
	if strings.TrimSpace(step.Status) == topicStepCompleted {
		noveltyCheckNotify(j, ctx, step, papersvc.TopicStepOutcomeSuccess,
			fmt.Sprintf("Idea Discovery · %s 已完成", topicdiscovery.TopicStageCodeToPhaseID(noveltyCheckStageCode)), deltaPrompt, deltaCompletion)
		if topicdiscovery.TopicStageNeedsUserCheckpoint(topicdiscovery.TopicStageCodeToPhaseID(noveltyCheckStageCode)) {
			noveltyCheckNotify(j, ctx, step, papersvc.TopicStepOutcomeCheckpoint,
				topicdiscovery.TopicStageCheckpointDetail(topicdiscovery.TopicStageCodeToPhaseID(noveltyCheckStageCode)), deltaPrompt, deltaCompletion)
		}
		return topicStepCompleted
	}
	if strings.TrimSpace(step.Status) == topicStepFailed {
		noveltyCheckNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed,
			fmt.Sprintf("Idea Discovery · %s 失败", topicdiscovery.TopicStageCodeToPhaseID(noveltyCheckStageCode)), deltaPrompt, deltaCompletion)
		return topicStepFailed
	}
	logger.InfoZ(ctx, "idea-discovery-phases-invoke-tick",
		zap.Uint64("step_id", step.ID),
		zap.String("stage_code", noveltyCheckStageCode),
		zap.String("phase_id", topicdiscovery.TopicStageCodeToPhaseID(noveltyCheckStageCode)))
	return topicStepSkipped
}

// runNoveltyCheckTick 组装工作区与 /novelty-check 请求，执行一次 ARIS 并写回 step。
func (j *IdeaDiscoveryPhasesJob) runNoveltyCheckTick(ctx context.Context, userID uint, runVersion int, step *entity.PaperOutputTopicStep) error {
	input, err := j.runs.LoadRunInputForUser(ctx, userID, runVersion)
	if err != nil {
		return err
	}
	if step == nil || strings.TrimSpace(step.StageCode) != noveltyCheckStageCode {
		return fmt.Errorf("非 novelty_check step: %q", step.StageCode)
	}
	svc := j.runs
	cfg := topicdiscovery.MustIdeaDiscoveryConfig()
	route, err := svc.PhasedBotRequireRoute(ctx, topicdiscovery.TopicStageBehaviorKey(noveltyCheckStageCode))
	if err != nil {
		return err
	}
	paperRoot := topicdiscovery.ARISRunPaperRoot(uint64(userID), step.ManuscriptID, runVersion)
	invokeCtx := svc.PhasedBotBuildInvokeContext(ctx, userID, runVersion, step, input)
	meta := paperstep.LoadStepExtra(step)
	runID := topicdiscovery.PhasedBotResolveRunID(paperRoot, step, input)
	meta["aris_run_id"] = runID
	meta["native_run_id"] = runID
	meta["idea_discovery_invoke_mode"] = "phased_per_step"
	meta["phase_id"] = topicdiscovery.TopicStageCodeToPhaseID(noveltyCheckStageCode)
	paperstep.PersistStepExtra(step, meta)
	_ = j.steps.SaveStep(ctx, step)
	cfg.ClearCLIWorkspaceSessions(paperRoot)
	if err := topicdiscovery.PrepareNativeIdeaDiscoveryWorkspace(paperRoot, input, invokeCtx.Language, cfg); err != nil {
		return err
	}
	argsLine := cfg.FormatIdeaDiscoveryArgsLine(input, strings.TrimSpace(input.Venue), invokeCtx.Language)
	if err := topicdiscovery.WriteNativePlatformArguments(paperRoot, argsLine); err != nil {
		return err
	}
	req := topicdiscovery.BuildTopicDiscoveryInvokeRequest(topicdiscovery.TopicStageBehaviorKey(noveltyCheckStageCode), route.CapabilityID, input, invokeCtx)
	raw, usage, err := svc.PhasedBotInvokeBehavior(ctx, step, req, "")
	if err != nil {
		return svc.PhasedBotHandleInvokeFailure(ctx, step, paperRoot, cfg, err)
	}
	meta = paperstep.LoadStepExtra(step)
	meta["aris_invoke_count"] = topicdiscovery.IdeaDiscoveryInvokeCount(meta) + 1
	meta["last_phased_invoke"] = topicdiscovery.TopicStageCodeToPhaseID(noveltyCheckStageCode)
	meta["last_invoke_preview"] = topicdiscovery.TruncateUTF8(string(raw), 4096)
	meta["tokens_prompt"] = usage.PromptTokens
	meta["tokens_completion"] = usage.CompletionTokens
	delete(meta, "aris_invoke_last_error")
	delete(meta, "aris_invoke_retries")
	paperstep.PersistStepExtra(step, meta)
	return j.advanceNoveltyCheckAfterInvoke(ctx, step, userID, runVersion, route, paperRoot, runID, cfg)
}

// advanceNoveltyCheckAfterInvoke 校验新颖性阶段产物，completed 本段或失败，并按配置开 research_review。
func (j *IdeaDiscoveryPhasesJob) advanceNoveltyCheckAfterInvoke(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	userID uint,
	runVersion int,
	route *entity.PaperStrategyRoute,
	paperRoot, runID string,
	cfg topicdiscovery.ARISIdeaDiscoveryConfig,
) error {
	svc := j.runs
	meta := paperstep.LoadStepExtra(step)
	if !topicdiscovery.IdeaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, topicdiscovery.TopicStageCodeToPhaseID(noveltyCheckStageCode), cfg) {
		if topicdiscovery.IdeaDiscoveryPhasedBotMayRetryPhase(cfg) {
			meta["async_llm"] = true
			paperstep.PersistStepExtra(step, meta)
			step.Status = topicStepRunning
			return j.steps.SaveStep(ctx, step)
		}
		reason := fmt.Sprintf("Idea Discovery 阶段 /%s 未产出约定产物", topicdiscovery.TopicStageCodeToPhaseID(noveltyCheckStageCode))
		return svc.PhasedBotPersistStageFailed(ctx, step, reason)
	}
	delete(meta, "async_llm")
	delete(meta, "awaiting_user_confirm")
	meta["async_llm_done"] = true
	meta["aris_pending_phase"] = topicdiscovery.TopicStageCodeToPhaseID(noveltyCheckStageCode)
	paperstep.PersistStepExtra(step, meta)
	if err := noveltyCheckCompleteStep(j, ctx, step); err != nil {
		return err
	}
	if cfg.PhaseNeedsUserCheckpoint(topicdiscovery.TopicStageCodeToPhaseID(noveltyCheckStageCode)) {
		return nil
	}
	return svc.PhasedBotBeginNextStageAsync(ctx, userID, runVersion, topicdiscovery.NextPhasedTopicStageCode(noveltyCheckStageCode))
}

// noveltyCheckCompleteStep 将 novelty_check 标为 completed（100%）。
func noveltyCheckCompleteStep(j *IdeaDiscoveryPhasesJob, ctx context.Context, step *entity.PaperOutputTopicStep) error {
	if step == nil {
		return nil
	}
	now := time.Now()
	step.Status = topicStepCompleted
	step.ExecProgressPct = 100
	step.CompletedAt = &now
	return j.steps.SaveStep(ctx, step)
}

// noveltyCheckIntMeta 从 step meta 读取 int（兼容 JSON float64）。
func noveltyCheckIntMeta(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	default:
		return 0
	}
}

// noveltyCheckMarkFailed 落库 failed 并推送失败通知。
func noveltyCheckMarkFailed(j *IdeaDiscoveryPhasesJob, ctx context.Context, step *entity.PaperOutputTopicStep, reason string, promptTokens, completionTokens int) {
	papersvc.FailTopicStep(step, reason)
	if err := j.steps.SaveStep(ctx, step); err != nil {
		logger.ErrorZ(ctx, "idea-discovery-phases-fail-save", zap.Uint64("step_id", step.ID), zap.Error(err))
		return
	}
	noveltyCheckNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed, reason, promptTokens, completionTokens)
}

// noveltyCheckNotify 推送本 step 执行结果（成功/失败/检查点）。
func noveltyCheckNotify(j *IdeaDiscoveryPhasesJob, ctx context.Context, step *entity.PaperOutputTopicStep, outcome, detail string, promptTokens, completionTokens int) {
	if step == nil {
		return
	}
	papersvc.NotifyTopicStepExecutionResult(ctx, j.manuscripts, papersvc.TopicStepExecNotify{
		StageCode: step.StageCode, Outcome: outcome, UserID: uint(step.UserID),
		ManuscriptID: step.ManuscriptID, RunVersion: step.RunVersion, StepID: step.ID,
		PromptTokens: promptTokens, CompletionTokens: completionTokens, Detail: detail,
	})
}
