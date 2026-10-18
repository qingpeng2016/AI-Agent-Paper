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

// RunIdeaCreator 处理 entry 传入的 idea_creator step：组装 /idea-creator 并 invoke ARIS。
func (j *IdeaDiscoveryPhasesJob) RunIdeaCreator(ctx context.Context, step *entity.PaperOutputTopicStep) (scanned, invoked, skipped, failed int) {
	if step == nil || j.runs == nil {
		return 0, 0, 0, 0
	}
	if strings.TrimSpace(step.Status) != topicStepRunning ||
		strings.TrimSpace(step.StageCode) != ideaCreatorStageCode {
		return 0, 0, 0, 0
	}
	userID := uint(step.UserID)
	runVersion := step.RunVersion
	priorOK, err := j.ideaCreatorPriorCompleted(ctx, userID, runVersion)
	if err != nil {
		ideaCreatorMarkFailed(j, ctx, step, err.Error(), 0, 0)
		return 1, 0, 0, 1
	}
	if !priorOK {
		return 0, 0, 0, 0
	}
	scanned = 1
	switch j.processIdeaCreatorStep(ctx, step) {
	case topicStepCompleted:
		invoked++
	case topicStepFailed:
		failed++
	default:
		skipped++
	}
	return scanned, invoked, skipped, failed
}

// ideaCreatorPriorCompleted 前序 research_lit 是否均已 completed。
func (j *IdeaDiscoveryPhasesJob) ideaCreatorPriorCompleted(ctx context.Context, userID uint, runVersion int) (bool, error) {
	for _, code := range ideaCreatorPriorStages {
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

// processIdeaCreatorStep 处理单条 idea_creator：invoke、通知，返回 completed/failed/skipped。
func (j *IdeaDiscoveryPhasesJob) processIdeaCreatorStep(ctx context.Context, step *entity.PaperOutputTopicStep) string {
	userID := uint(step.UserID)
	runVersion := step.RunVersion
	meta := paperstep.LoadStepExtra(step)
	beforePrompt := ideaCreatorIntMeta(meta["tokens_prompt"])
	beforeCompletion := ideaCreatorIntMeta(meta["tokens_completion"])
	if err := j.runIdeaCreatorTick(ctx, userID, runVersion, step); err != nil {
		if strings.TrimSpace(step.Status) != topicStepFailed {
			ideaCreatorMarkFailed(j, ctx, step, err.Error(), 0, 0)
		} else {
			meta = paperstep.LoadStepExtra(step)
			ideaCreatorNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed, err.Error(),
				ideaCreatorIntMeta(meta["tokens_prompt"])-beforePrompt, ideaCreatorIntMeta(meta["tokens_completion"])-beforeCompletion)
		}
		return topicStepFailed
	}
	meta = paperstep.LoadStepExtra(step)
	deltaPrompt := ideaCreatorIntMeta(meta["tokens_prompt"]) - beforePrompt
	deltaCompletion := ideaCreatorIntMeta(meta["tokens_completion"]) - beforeCompletion
	if strings.TrimSpace(step.Status) == topicStepCompleted {
		ideaCreatorNotify(j, ctx, step, papersvc.TopicStepOutcomeSuccess,
			fmt.Sprintf("Idea Discovery · %s 已完成", topicdiscovery.TopicStageCodeToPhaseID(ideaCreatorStageCode)), deltaPrompt, deltaCompletion)
		if topicdiscovery.TopicStageNeedsUserCheckpoint(topicdiscovery.TopicStageCodeToPhaseID(ideaCreatorStageCode)) {
			ideaCreatorNotify(j, ctx, step, papersvc.TopicStepOutcomeCheckpoint,
				topicdiscovery.TopicStageCheckpointDetail(topicdiscovery.TopicStageCodeToPhaseID(ideaCreatorStageCode)), deltaPrompt, deltaCompletion)
		}
		return topicStepCompleted
	}
	if strings.TrimSpace(step.Status) == topicStepFailed {
		ideaCreatorNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed,
			fmt.Sprintf("Idea Discovery · %s 失败", topicdiscovery.TopicStageCodeToPhaseID(ideaCreatorStageCode)), deltaPrompt, deltaCompletion)
		return topicStepFailed
	}
	logger.InfoZ(ctx, "idea-discovery-phases-invoke-tick",
		zap.Uint64("step_id", step.ID),
		zap.String("stage_code", ideaCreatorStageCode),
		zap.String("phase_id", topicdiscovery.TopicStageCodeToPhaseID(ideaCreatorStageCode)))
	return topicStepSkipped
}

// runIdeaCreatorTick 组装工作区与 /idea-creator 请求，执行一次 ARIS 并写回 step。
func (j *IdeaDiscoveryPhasesJob) runIdeaCreatorTick(ctx context.Context, userID uint, runVersion int, step *entity.PaperOutputTopicStep) error {
	input, err := j.runs.LoadRunInputForUser(ctx, userID, runVersion)
	if err != nil {
		return err
	}
	if step == nil || strings.TrimSpace(step.StageCode) != ideaCreatorStageCode {
		return fmt.Errorf("非 idea_creator step: %q", step.StageCode)
	}
	svc := j.runs
	cfg := topicdiscovery.MustIdeaDiscoveryConfig()
	route, err := svc.PhasedBotRequireRoute(ctx, topicdiscovery.TopicStageBehaviorKey(ideaCreatorStageCode))
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
	meta["phase_id"] = topicdiscovery.TopicStageCodeToPhaseID(ideaCreatorStageCode)
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
	req := topicdiscovery.BuildTopicDiscoveryInvokeRequest(topicdiscovery.TopicStageBehaviorKey(ideaCreatorStageCode), route.CapabilityID, input, invokeCtx)
	raw, usage, err := svc.PhasedBotInvokeBehavior(ctx, step, req, "")
	if err != nil {
		return svc.PhasedBotHandleInvokeFailure(ctx, step, paperRoot, cfg, err)
	}
	meta = paperstep.LoadStepExtra(step)
	meta["aris_invoke_count"] = topicdiscovery.IdeaDiscoveryInvokeCount(meta) + 1
	meta["last_phased_invoke"] = topicdiscovery.TopicStageCodeToPhaseID(ideaCreatorStageCode)
	meta["last_invoke_preview"] = topicdiscovery.TruncateUTF8(string(raw), 4096)
	meta["tokens_prompt"] = usage.PromptTokens
	meta["tokens_completion"] = usage.CompletionTokens
	delete(meta, "aris_invoke_last_error")
	delete(meta, "aris_invoke_retries")
	paperstep.PersistStepExtra(step, meta)
	return j.advanceIdeaCreatorAfterInvoke(ctx, step, userID, runVersion, route, paperRoot, runID, cfg)
}

// advanceIdeaCreatorAfterInvoke 校验选题大纲产物，completed 本段或失败，并按配置开 novelty_check。
func (j *IdeaDiscoveryPhasesJob) advanceIdeaCreatorAfterInvoke(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	userID uint,
	runVersion int,
	route *entity.PaperStrategyRoute,
	paperRoot, runID string,
	cfg topicdiscovery.ARISIdeaDiscoveryConfig,
) error {
	svc := j.runs
	if !topicdiscovery.IdeaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, topicdiscovery.TopicStageCodeToPhaseID(ideaCreatorStageCode), cfg) {
		if topicdiscovery.IdeaDiscoveryPhasedBotMayRetryPhase(cfg) {
			step.Status = topicStepRunning
			return j.steps.SaveStep(ctx, step)
		}
		reason := fmt.Sprintf("Idea Discovery 阶段 /%s 未产出约定产物", topicdiscovery.TopicStageCodeToPhaseID(ideaCreatorStageCode))
		return svc.PhasedBotPersistStageFailed(ctx, step, reason)
	}
	if err := ideaCreatorCompleteStep(j, ctx, step); err != nil {
		return err
	}
	if cfg.PhaseNeedsUserCheckpoint(topicdiscovery.TopicStageCodeToPhaseID(ideaCreatorStageCode)) {
		return nil
	}
	return svc.PhasedBotBeginNextStageAsync(ctx, userID, runVersion, topicdiscovery.NextPhasedTopicStageCode(ideaCreatorStageCode))
}

// ideaCreatorCompleteStep 将 idea_creator 标为 completed（100%）。
func ideaCreatorCompleteStep(j *IdeaDiscoveryPhasesJob, ctx context.Context, step *entity.PaperOutputTopicStep) error {
	if step == nil {
		return nil
	}
	now := time.Now()
	step.Status = topicStepCompleted
	step.ExecProgressPct = 100
	step.CompletedAt = &now
	return j.steps.SaveStep(ctx, step)
}

// ideaCreatorIntMeta 从 step meta 读取 int（兼容 JSON float64）。
func ideaCreatorIntMeta(v any) int {
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

// ideaCreatorMarkFailed 落库 failed 并推送失败通知。
func ideaCreatorMarkFailed(j *IdeaDiscoveryPhasesJob, ctx context.Context, step *entity.PaperOutputTopicStep, reason string, promptTokens, completionTokens int) {
	papersvc.FailTopicStep(step, reason)
	if err := j.steps.SaveStep(ctx, step); err != nil {
		logger.ErrorZ(ctx, "idea-discovery-phases-fail-save", zap.Uint64("step_id", step.ID), zap.Error(err))
		return
	}
	ideaCreatorNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed, reason, promptTokens, completionTokens)
}

// ideaCreatorNotify 推送本 step 执行结果（成功/失败/检查点）。
func ideaCreatorNotify(j *IdeaDiscoveryPhasesJob, ctx context.Context, step *entity.PaperOutputTopicStep, outcome, detail string, promptTokens, completionTokens int) {
	if step == nil {
		return
	}
	papersvc.NotifyTopicStepExecutionResult(ctx, j.manuscripts, papersvc.TopicStepExecNotify{
		StageCode: step.StageCode, Outcome: outcome, UserID: uint(step.UserID),
		ManuscriptID: step.ManuscriptID, RunVersion: step.RunVersion, StepID: step.ID,
		PromptTokens: promptTokens, CompletionTokens: completionTokens, Detail: detail,
	})
}
