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

const (
	researchLitStageCode   = "research_lit"
	researchLitPhaseID     = "research-lit"
	researchLitBehaviorKey = "topic.stage.research_lit"
	researchLitNextStage   = "idea_creator"
)

// RunResearchLit 扫 research_lit running 行，组装 /research-lit 命令并 invoke ARIS。
func (j *IdeaDiscoveryPhasesJob) RunResearchLit(ctx context.Context) (scanned, invoked, skipped, failed int) {
	rows, err := j.steps.ListByStatusAndStageOnMaxRunPerManuscript(ctx, "running", researchLitStageCode)
	if err != nil {
		logger.ErrorZ(ctx, "idea-discovery-phases-list-failed",
			zap.String("stage_code", researchLitStageCode), zap.Error(err))
		return 0, 0, 0, 0
	}
	for i := range rows {
		if !researchLitQueueStep(&rows[i]) {
			continue
		}
		scanned++
		switch j.processResearchLitStep(ctx, &rows[i]) {
		case "completed":
			invoked++
		case "failed":
			failed++
		default:
			skipped++
		}
	}
	return scanned, invoked, skipped, failed
}

// researchLitQueueStep 是否为本段待 bot 处理的 running 行（stage=research_lit 且 async_llm）。
func researchLitQueueStep(step *entity.PaperOutputTopicStep) bool {
	if step == nil || strings.TrimSpace(step.Status) != "running" {
		return false
	}
	if strings.TrimSpace(step.StageCode) != researchLitStageCode {
		return false
	}
	meta := paperstep.LoadStepExtra(step)
	async, ok := meta["async_llm"].(bool)
	return ok && async
}

// researchLitPrepareInvoke 本 tick 是否允许调 ARIS（async_llm、状态、非人工暂停）。
func (j *IdeaDiscoveryPhasesJob) researchLitPrepareInvoke(ctx context.Context, userID uint, runVersion int, step *entity.PaperOutputTopicStep) (bool, error) {
	meta := paperstep.LoadStepExtra(step)
	if v, ok := meta["async_llm"].(bool); !ok || !v {
		return false, nil
	}
	if strings.TrimSpace(step.Status) != "running" {
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

// processResearchLitStep 处理单条 research_lit：invoke、通知，返回 completed/failed/skipped。
func (j *IdeaDiscoveryPhasesJob) processResearchLitStep(ctx context.Context, step *entity.PaperOutputTopicStep) string {
	if step == nil || j.runs == nil {
		return "skipped"
	}
	if strings.TrimSpace(step.Status) == "failed" {
		return "skipped"
	}
	if step.ErrorMsg != nil && strings.TrimSpace(*step.ErrorMsg) != "" {
		if strings.TrimSpace(step.Status) != "failed" {
			papersvc.FailTopicStep(step, *step.ErrorMsg)
			_ = j.steps.SaveStep(ctx, step)
		}
		return "skipped"
	}
	meta := papersvc.LoadStepExtra(step)
	if async, ok := meta["async_llm"].(bool); !ok || !async {
		return "skipped"
	}
	userID := uint(step.UserID)
	runVersion := step.RunVersion
	invoke, err := j.researchLitPrepareInvoke(ctx, userID, runVersion, step)
	if err != nil {
		researchLitMarkFailed(j, ctx, step, err.Error(), 0, 0)
		return "failed"
	}
	if !invoke {
		return "skipped"
	}
	meta = papersvc.LoadStepExtra(step)
	beforePrompt := researchLitIntMeta(meta["tokens_prompt"])
	beforeCompletion := researchLitIntMeta(meta["tokens_completion"])
	if err = j.runResearchLitTick(ctx, userID, runVersion, step); err != nil {
		if strings.TrimSpace(step.Status) != "failed" {
			researchLitMarkFailed(j, ctx, step, err.Error(), 0, 0)
		} else {
			meta = papersvc.LoadStepExtra(step)
			researchLitNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed, err.Error(),
				researchLitIntMeta(meta["tokens_prompt"])-beforePrompt, researchLitIntMeta(meta["tokens_completion"])-beforeCompletion)
		}
		return "failed"
	}
	meta = papersvc.LoadStepExtra(step)
	deltaPrompt := researchLitIntMeta(meta["tokens_prompt"]) - beforePrompt
	deltaCompletion := researchLitIntMeta(meta["tokens_completion"]) - beforeCompletion
	if strings.TrimSpace(step.Status) == "completed" {
		researchLitNotify(j, ctx, step, papersvc.TopicStepOutcomeSuccess,
			fmt.Sprintf("Idea Discovery · %s 已完成", researchLitPhaseID), deltaPrompt, deltaCompletion)
		if topicdiscovery.TopicStageNeedsUserCheckpoint(researchLitPhaseID) {
			researchLitNotify(j, ctx, step, papersvc.TopicStepOutcomeCheckpoint,
				topicdiscovery.TopicStageCheckpointDetail(researchLitPhaseID), deltaPrompt, deltaCompletion)
		}
		return "completed"
	}
	if strings.TrimSpace(step.Status) == "failed" {
		researchLitNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed,
			fmt.Sprintf("Idea Discovery · %s 失败", researchLitPhaseID), deltaPrompt, deltaCompletion)
		return "failed"
	}
	logger.InfoZ(ctx, "idea-discovery-phases-invoke-tick",
		zap.Uint64("step_id", step.ID),
		zap.String("stage_code", researchLitStageCode),
		zap.String("phase_id", researchLitPhaseID))
	return "skipped"
}

// runResearchLitTick 组装工作区与 /research-lit 请求，执行一次 ARIS 并写回 step。
func (j *IdeaDiscoveryPhasesJob) runResearchLitTick(ctx context.Context, userID uint, runVersion int, step *entity.PaperOutputTopicStep) error {
	input, err := j.runs.LoadRunInputForUser(ctx, userID, runVersion)
	if err != nil {
		return err
	}
	if step == nil || strings.TrimSpace(step.StageCode) != researchLitStageCode {
		return fmt.Errorf("非 research_lit step: %q", step.StageCode)
	}
	svc := j.runs
	cfg := topicdiscovery.MustIdeaDiscoveryConfig()
	route, err := svc.PhasedBotRequireRoute(ctx, researchLitBehaviorKey)
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
	meta["phase_id"] = researchLitPhaseID
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
	req := topicdiscovery.BuildTopicDiscoveryInvokeRequest(researchLitBehaviorKey, route.CapabilityID, input, invokeCtx)
	raw, usage, err := svc.PhasedBotInvokeBehavior(ctx, step, req, "")
	if err != nil {
		return svc.PhasedBotHandleInvokeFailure(ctx, step, paperRoot, cfg, err)
	}
	meta = paperstep.LoadStepExtra(step)
	meta["aris_invoke_count"] = topicdiscovery.IdeaDiscoveryInvokeCount(meta) + 1
	meta["last_phased_invoke"] = researchLitPhaseID
	meta["last_invoke_preview"] = topicdiscovery.TruncateUTF8(string(raw), 4096)
	meta["tokens_prompt"] = usage.PromptTokens
	meta["tokens_completion"] = usage.CompletionTokens
	delete(meta, "aris_invoke_last_error")
	delete(meta, "aris_invoke_retries")
	paperstep.PersistStepExtra(step, meta)
	if topicdiscovery.IdeaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, researchLitPhaseID, cfg) {
		_ = svc.PhasedBotIngestLiterature(ctx, userID, runVersion, step.ManuscriptID, paperRoot, cfg)
	}
	return j.advanceResearchLitAfterInvoke(ctx, step, userID, runVersion, route, paperRoot, runID, cfg)
}

// advanceResearchLitAfterInvoke 校验文献阶段产物，completed 本段或失败，并按配置开 idea_creator。
func (j *IdeaDiscoveryPhasesJob) advanceResearchLitAfterInvoke(
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
	if !topicdiscovery.IdeaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, researchLitPhaseID, cfg) {
		if topicdiscovery.IdeaDiscoveryPhasedBotMayRetryPhase(cfg) {
			meta["async_llm"] = true
			paperstep.PersistStepExtra(step, meta)
			step.Status = "running"
			return j.steps.SaveStep(ctx, step)
		}
		reason := fmt.Sprintf("Idea Discovery 阶段 /%s 未产出约定产物", researchLitPhaseID)
		return svc.PhasedBotPersistStageFailed(ctx, step, reason)
	}
	delete(meta, "async_llm")
	delete(meta, "awaiting_user_confirm")
	meta["async_llm_done"] = true
	meta["aris_pending_phase"] = researchLitPhaseID
	paperstep.PersistStepExtra(step, meta)
	if err := researchLitCompleteStep(j, ctx, step); err != nil {
		return err
	}
	if cfg.PhaseNeedsUserCheckpoint(researchLitPhaseID) {
		return nil
	}
	return svc.PhasedBotBeginNextStageAsync(ctx, userID, runVersion, researchLitNextStage)
}

// researchLitCompleteStep 将 research_lit 标为 completed（100%）。
func researchLitCompleteStep(j *IdeaDiscoveryPhasesJob, ctx context.Context, step *entity.PaperOutputTopicStep) error {
	if step == nil {
		return nil
	}
	now := time.Now()
	step.Status = "completed"
	step.ExecProgressPct = 100
	step.CompletedAt = &now
	return j.steps.SaveStep(ctx, step)
}

// researchLitIntMeta 从 step meta 读取 int（兼容 JSON float64）。
func researchLitIntMeta(v any) int {
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

// researchLitMarkFailed 落库 failed 并推送失败通知。
func researchLitMarkFailed(j *IdeaDiscoveryPhasesJob, ctx context.Context, step *entity.PaperOutputTopicStep, reason string, promptTokens, completionTokens int) {
	papersvc.FailTopicStep(step, reason)
	if err := j.steps.SaveStep(ctx, step); err != nil {
		logger.ErrorZ(ctx, "idea-discovery-phases-fail-save", zap.Uint64("step_id", step.ID), zap.Error(err))
		return
	}
	researchLitNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed, reason, promptTokens, completionTokens)
}

// researchLitNotify 推送本 step 执行结果（成功/失败/检查点）。
func researchLitNotify(j *IdeaDiscoveryPhasesJob, ctx context.Context, step *entity.PaperOutputTopicStep, outcome, detail string, promptTokens, completionTokens int) {
	if step == nil {
		return
	}
	papersvc.NotifyTopicStepExecutionResult(ctx, j.manuscripts, papersvc.TopicStepExecNotify{
		StageCode: step.StageCode, Outcome: outcome, UserID: uint(step.UserID),
		ManuscriptID: step.ManuscriptID, RunVersion: step.RunVersion, StepID: step.ID,
		PromptTokens: promptTokens, CompletionTokens: completionTokens, Detail: detail,
	})
}
