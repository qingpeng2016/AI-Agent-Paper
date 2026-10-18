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

// RunResearchLit 处理 entry 传入的 research_lit step：组装 /research-lit 并 invoke ARIS。
func (j *IdeaDiscoveryPhasesJob) RunResearchLit(ctx context.Context, step *entity.PaperOutputTopicStep) (scanned, invoked, skipped, failed int) {
	if step == nil || j.runs == nil {
		return 0, 0, 0, 0
	}
	if strings.TrimSpace(step.Status) != topicStepRunning ||
		strings.TrimSpace(step.StageCode) != researchLitStageCode {
		return 0, 0, 0, 0
	}
	scanned = 1
	switch j.processResearchLitStep(ctx, step) {
	case topicStepCompleted:
		invoked++
	case topicStepFailed:
		failed++
	default:
		skipped++
	}
	return scanned, invoked, skipped, failed
}

// processResearchLitStep 处理单条 research_lit：invoke、通知，返回 completed/failed/skipped。
func (j *IdeaDiscoveryPhasesJob) processResearchLitStep(ctx context.Context, step *entity.PaperOutputTopicStep) string {
	userID := uint(step.UserID)
	runVersion := step.RunVersion
	meta := papersvc.LoadStepExtra(step)
	beforePrompt := researchLitIntMeta(meta["tokens_prompt"])
	beforeCompletion := researchLitIntMeta(meta["tokens_completion"])
	if err := j.runResearchLitTick(ctx, userID, runVersion, step); err != nil {
		if strings.TrimSpace(step.Status) != topicStepFailed {
			papersvc.FailTopicStep(step, err.Error())
			if saveErr := j.steps.SaveStep(ctx, step); saveErr != nil {
				logger.ErrorZ(ctx, "idea-discovery-phases-fail-save", zap.Uint64("step_id", step.ID), zap.Error(saveErr))
			} else {
				researchLitNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed, err.Error(), 0, 0)
			}
		} else {
			meta = papersvc.LoadStepExtra(step)
			researchLitNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed, err.Error(),
				researchLitIntMeta(meta["tokens_prompt"])-beforePrompt, researchLitIntMeta(meta["tokens_completion"])-beforeCompletion)
		}
		return topicStepFailed
	}
	meta = papersvc.LoadStepExtra(step)
	deltaPrompt := researchLitIntMeta(meta["tokens_prompt"]) - beforePrompt
	deltaCompletion := researchLitIntMeta(meta["tokens_completion"]) - beforeCompletion
	if strings.TrimSpace(step.Status) == topicStepCompleted {
		researchLitNotify(j, ctx, step, papersvc.TopicStepOutcomeSuccess,
			fmt.Sprintf("Idea Discovery · %s 已完成", topicdiscovery.TopicStageCodeToPhaseID(researchLitStageCode)), deltaPrompt, deltaCompletion)
		if topicdiscovery.TopicStageNeedsUserCheckpoint(topicdiscovery.TopicStageCodeToPhaseID(researchLitStageCode)) {
			researchLitNotify(j, ctx, step, papersvc.TopicStepOutcomeCheckpoint,
				topicdiscovery.TopicStageCheckpointDetail(topicdiscovery.TopicStageCodeToPhaseID(researchLitStageCode)), deltaPrompt, deltaCompletion)
		}
		return topicStepCompleted
	}
	if strings.TrimSpace(step.Status) == topicStepFailed {
		researchLitNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed,
			fmt.Sprintf("Idea Discovery · %s 失败", topicdiscovery.TopicStageCodeToPhaseID(researchLitStageCode)), deltaPrompt, deltaCompletion)
		return topicStepFailed
	}
	logger.InfoZ(ctx, "idea-discovery-phases-invoke-tick",
		zap.Uint64("step_id", step.ID),
		zap.String("stage_code", researchLitStageCode),
		zap.String("phase_id", topicdiscovery.TopicStageCodeToPhaseID(researchLitStageCode)))
	return topicStepSkipped
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
	route, err := svc.PhasedBotRequireRoute(ctx, topicdiscovery.TopicStageBehaviorKey(researchLitStageCode))
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
	meta["phase_id"] = topicdiscovery.TopicStageCodeToPhaseID(researchLitStageCode)
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
	req := topicdiscovery.BuildTopicDiscoveryInvokeRequest(topicdiscovery.TopicStageBehaviorKey(researchLitStageCode), route.CapabilityID, input, invokeCtx)
	raw, usage, err := svc.PhasedBotInvokeBehavior(ctx, step, req, "")
	if err != nil {
		return svc.PhasedBotHandleInvokeFailure(ctx, step, paperRoot, cfg, err)
	}
	meta = paperstep.LoadStepExtra(step)
	meta["aris_invoke_count"] = topicdiscovery.IdeaDiscoveryInvokeCount(meta) + 1
	meta["last_phased_invoke"] = topicdiscovery.TopicStageCodeToPhaseID(researchLitStageCode)
	meta["last_invoke_preview"] = topicdiscovery.TruncateUTF8(string(raw), 4096)
	meta["tokens_prompt"] = usage.PromptTokens
	meta["tokens_completion"] = usage.CompletionTokens
	delete(meta, "aris_invoke_last_error")
	delete(meta, "aris_invoke_retries")
	paperstep.PersistStepExtra(step, meta)
	if topicdiscovery.IdeaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, topicdiscovery.TopicStageCodeToPhaseID(researchLitStageCode), cfg) {
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
	if !topicdiscovery.IdeaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, topicdiscovery.TopicStageCodeToPhaseID(researchLitStageCode), cfg) {
		if topicdiscovery.IdeaDiscoveryPhasedBotMayRetryPhase(cfg) {
			step.Status = topicStepRunning
			return j.steps.SaveStep(ctx, step)
		}
		reason := fmt.Sprintf("Idea Discovery 阶段 /%s 未产出约定产物", topicdiscovery.TopicStageCodeToPhaseID(researchLitStageCode))
		return svc.PhasedBotPersistStageFailed(ctx, step, reason)
	}
	now := time.Now()
	step.Status = topicStepCompleted
	step.ExecProgressPct = 100
	step.CompletedAt = &now
	if err := j.steps.SaveStep(ctx, step); err != nil {
		return err
	}
	if cfg.PhaseNeedsUserCheckpoint(topicdiscovery.TopicStageCodeToPhaseID(researchLitStageCode)) {
		return nil
	}
	return svc.PhasedBotBeginNextStageAsync(ctx, userID, runVersion, topicdiscovery.NextPhasedTopicStageCode(researchLitStageCode))
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
