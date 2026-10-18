package ideadiscoveryphases

import (
	"context"
	"fmt"
	"strings"

	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/topicdiscovery"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"go.uber.org/zap"
)

const (
	researchRefineStageCode   = "research_refine"
	researchRefinePhaseID     = "research-refine-pipeline"
	researchRefineBehaviorKey = "topic.stage.research_refine"
)

var researchRefinePriorStages = []string{"research_lit", "idea_creator", "novelty_check", "research_review"}

// RunResearchRefine 扫 research_refine running 行，组装 /research-refine-pipeline 命令并 invoke ARIS。
func (j *IdeaDiscoveryPhasesJob) RunResearchRefine(ctx context.Context) (scanned, invoked, skipped, failed int) {
	rows, err := j.steps.ListByStatusAndStageOnMaxRunPerManuscript(ctx, "running", researchRefineStageCode)
	if err != nil {
		logger.ErrorZ(ctx, "idea-discovery-phases-list-failed",
			zap.String("stage_code", researchRefineStageCode), zap.Error(err))
		return 0, 0, 0, 0
	}
	for i := range rows {
		if !researchRefineQueueStep(&rows[i]) {
			continue
		}
		scanned++
		switch j.processResearchRefineStep(ctx, &rows[i]) {
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

// researchRefineQueueStep 是否为本段待 bot 处理的 running 行（stage=research_refine 且 async_llm）。
func researchRefineQueueStep(step *entity.PaperOutputTopicStep) bool {
	if step == nil || strings.TrimSpace(step.Status) != "running" {
		return false
	}
	if strings.TrimSpace(step.StageCode) != researchRefineStageCode {
		return false
	}
	meta := paperstep.LoadStepExtra(step)
	async, ok := meta["async_llm"].(bool)
	return ok && async
}

// researchRefinePriorCompleted 前序四段是否均已 completed。
func (j *IdeaDiscoveryPhasesJob) researchRefinePriorCompleted(ctx context.Context, userID uint, runVersion int) (bool, error) {
	for _, code := range researchRefinePriorStages {
		step, err := j.steps.GetStep(ctx, uint64(userID), runVersion, code)
		if err != nil {
			return false, err
		}
		if step == nil || strings.TrimSpace(step.Status) != "completed" {
			return false, nil
		}
	}
	return true, nil
}

// researchRefinePrepareInvoke 本 tick 是否允许调 ARIS（async_llm、前序、状态、非人工暂停）。
func (j *IdeaDiscoveryPhasesJob) researchRefinePrepareInvoke(ctx context.Context, userID uint, runVersion int, step *entity.PaperOutputTopicStep) (bool, error) {
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
	ok, err := j.researchRefinePriorCompleted(ctx, userID, runVersion)
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

// processResearchRefineStep 处理单条 research_refine：invoke、通知，返回 completed/failed/skipped。
func (j *IdeaDiscoveryPhasesJob) processResearchRefineStep(ctx context.Context, step *entity.PaperOutputTopicStep) string {
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
	meta := paperstep.LoadStepExtra(step)
	if async, ok := meta["async_llm"].(bool); !ok || !async {
		return "skipped"
	}
	userID := uint(step.UserID)
	runVersion := step.RunVersion
	invoke, err := j.researchRefinePrepareInvoke(ctx, userID, runVersion, step)
	if err != nil {
		researchRefineMarkFailed(j, ctx, step, err.Error(), 0, 0)
		return "failed"
	}
	if !invoke {
		return "skipped"
	}
	meta = paperstep.LoadStepExtra(step)
	beforePrompt := researchRefineIntMeta(meta["tokens_prompt"])
	beforeCompletion := researchRefineIntMeta(meta["tokens_completion"])
	if err = j.runResearchRefineTick(ctx, userID, runVersion, step); err != nil {
		if strings.TrimSpace(step.Status) != "failed" {
			researchRefineMarkFailed(j, ctx, step, err.Error(), 0, 0)
		} else {
			meta = paperstep.LoadStepExtra(step)
			researchRefineNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed, err.Error(),
				researchRefineIntMeta(meta["tokens_prompt"])-beforePrompt, researchRefineIntMeta(meta["tokens_completion"])-beforeCompletion)
		}
		return "failed"
	}
	meta = paperstep.LoadStepExtra(step)
	deltaPrompt := researchRefineIntMeta(meta["tokens_prompt"]) - beforePrompt
	deltaCompletion := researchRefineIntMeta(meta["tokens_completion"]) - beforeCompletion
	if strings.TrimSpace(step.Status) == "completed" {
		researchRefineNotify(j, ctx, step, papersvc.TopicStepOutcomeSuccess,
			fmt.Sprintf("Idea Discovery · %s 已完成", researchRefinePhaseID), deltaPrompt, deltaCompletion)
		if topicdiscovery.TopicStageNeedsUserCheckpoint(researchRefinePhaseID) {
			researchRefineNotify(j, ctx, step, papersvc.TopicStepOutcomeCheckpoint,
				topicdiscovery.TopicStageCheckpointDetail(researchRefinePhaseID), deltaPrompt, deltaCompletion)
		}
		return "completed"
	}
	if strings.TrimSpace(step.Status) == "failed" {
		researchRefineNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed,
			fmt.Sprintf("Idea Discovery · %s 失败", researchRefinePhaseID), deltaPrompt, deltaCompletion)
		return "failed"
	}
	logger.InfoZ(ctx, "idea-discovery-phases-invoke-tick",
		zap.Uint64("step_id", step.ID),
		zap.String("stage_code", researchRefineStageCode),
		zap.String("phase_id", researchRefinePhaseID))
	return "skipped"
}

// runResearchRefineTick 组装工作区与 /research-refine-pipeline 请求，执行一次 ARIS 并写回 step。
func (j *IdeaDiscoveryPhasesJob) runResearchRefineTick(ctx context.Context, userID uint, runVersion int, step *entity.PaperOutputTopicStep) error {
	input, err := j.runs.LoadRunInputForUser(ctx, userID, runVersion)
	if err != nil {
		return err
	}
	if step == nil || strings.TrimSpace(step.StageCode) != researchRefineStageCode {
		return fmt.Errorf("非 research_refine step: %q", step.StageCode)
	}
	svc := j.runs
	cfg := topicdiscovery.MustIdeaDiscoveryConfig()
	route, err := svc.PhasedBotRequireRoute(ctx, researchRefineBehaviorKey)
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
	meta["phase_id"] = researchRefinePhaseID
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
	req := topicdiscovery.BuildTopicDiscoveryInvokeRequest(researchRefineBehaviorKey, route.CapabilityID, input, invokeCtx)
	raw, usage, err := svc.PhasedBotInvokeBehavior(ctx, step, req, "")
	if err != nil {
		return svc.PhasedBotHandleInvokeFailure(ctx, step, paperRoot, cfg, err)
	}
	meta = paperstep.LoadStepExtra(step)
	meta["aris_invoke_count"] = topicdiscovery.IdeaDiscoveryInvokeCount(meta) + 1
	meta["last_phased_invoke"] = researchRefinePhaseID
	meta["last_invoke_preview"] = topicdiscovery.TruncateUTF8(string(raw), 4096)
	meta["tokens_prompt"] = usage.PromptTokens
	meta["tokens_completion"] = usage.CompletionTokens
	delete(meta, "aris_invoke_last_error")
	delete(meta, "aris_invoke_retries")
	paperstep.PersistStepExtra(step, meta)
	return j.advanceResearchRefineAfterInvoke(ctx, step, userID, runVersion, route, paperRoot, runID, cfg)
}

// advanceResearchRefineAfterInvoke 校验 refine 产物；成功则走 idea_discovery gate 收尾（topic_confirm）。
func (j *IdeaDiscoveryPhasesJob) advanceResearchRefineAfterInvoke(
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
	if !topicdiscovery.IdeaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, researchRefinePhaseID, cfg) {
		if topicdiscovery.IdeaDiscoveryPhasedBotMayRetryPhase(cfg) {
			meta["async_llm"] = true
			paperstep.PersistStepExtra(step, meta)
			step.Status = "running"
			return j.steps.SaveStep(ctx, step)
		}
		reason := fmt.Sprintf("Idea Discovery 阶段 /%s 未产出约定产物", researchRefinePhaseID)
		return svc.PhasedBotPersistStageFailed(ctx, step, reason)
	}
	delete(meta, "async_llm")
	delete(meta, "awaiting_user_confirm")
	meta["async_llm_done"] = true
	meta["aris_pending_phase"] = researchRefinePhaseID
	paperstep.PersistStepExtra(step, meta)
	// 五段末段：下一段为 topic_confirm，由 gate 收尾并标 completed（与原先 advance 一致）。
	return svc.PhasedBotFinishAfterGate(ctx, step, userID, runVersion, route, paperRoot, runID, cfg)
}

// researchRefineIntMeta 从 step meta 读取 int（兼容 JSON float64）。
func researchRefineIntMeta(v any) int {
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

// researchRefineMarkFailed 落库 failed 并推送失败通知。
func researchRefineMarkFailed(j *IdeaDiscoveryPhasesJob, ctx context.Context, step *entity.PaperOutputTopicStep, reason string, promptTokens, completionTokens int) {
	papersvc.FailTopicStep(step, reason)
	if err := j.steps.SaveStep(ctx, step); err != nil {
		logger.ErrorZ(ctx, "idea-discovery-phases-fail-save", zap.Uint64("step_id", step.ID), zap.Error(err))
		return
	}
	researchRefineNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed, reason, promptTokens, completionTokens)
}

// researchRefineNotify 推送本 step 执行结果（成功/失败/检查点）。
func researchRefineNotify(j *IdeaDiscoveryPhasesJob, ctx context.Context, step *entity.PaperOutputTopicStep, outcome, detail string, promptTokens, completionTokens int) {
	if step == nil {
		return
	}
	papersvc.NotifyTopicStepExecutionResult(ctx, j.manuscripts, papersvc.TopicStepExecNotify{
		StageCode: step.StageCode, Outcome: outcome, UserID: uint(step.UserID),
		ManuscriptID: step.ManuscriptID, RunVersion: step.RunVersion, StepID: step.ID,
		PromptTokens: promptTokens, CompletionTokens: completionTokens, Detail: detail,
	})
}
