package ideadiscoveryphases

import (
	"context"
	"fmt"
	"strings"
	"time"

	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/topicdiscovery"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"go.uber.org/zap"
)

// RunResearchLit 对 running 的 research_lit step 执行一次 /research-lit ARIS。
func (j *IdeaDiscoveryPhasesJob) RunResearchLit(ctx context.Context, step *entity.PaperOutputTopicStep) (scanned, invoked, skipped, failed int) {
	if step == nil || j.runs == nil {
		return 0, 0, 0, 0
	}
	if strings.TrimSpace(step.Status) != topicStepRunning ||
		strings.TrimSpace(step.StageCode) != researchLitStageCode {
		return 0, 0, 0, 0
	}
	scanned = 1

	phaseID := topicdiscovery.TopicStageCodeToPhaseID(researchLitStageCode)
	if err := j.invokeResearchLit(ctx, step, phaseID); err != nil {
		if strings.TrimSpace(step.Status) != topicStepFailed {
			papersvc.FailTopicStep(step, err.Error())
			if saveErr := j.steps.SaveStep(ctx, step); saveErr != nil {
				logger.ErrorZ(ctx, "idea-discovery-phases-fail-save", zap.Uint64("step_id", step.ID), zap.Error(saveErr))
			}
		}
		researchLitNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed, err.Error(), 0, 0)
		return scanned, 0, 0, 1
	}

	switch strings.TrimSpace(step.Status) {
	case topicStepCompleted:
		researchLitNotify(j, ctx, step, papersvc.TopicStepOutcomeSuccess,
			fmt.Sprintf("Idea Discovery · %s 已完成", phaseID), 0, 0)
		if topicdiscovery.TopicStageNeedsUserCheckpoint(phaseID) {
			researchLitNotify(j, ctx, step, papersvc.TopicStepOutcomeCheckpoint,
				topicdiscovery.TopicStageCheckpointDetail(phaseID), 0, 0)
		}
		return scanned, 1, 0, 0
	case topicStepFailed:
		return scanned, 0, 0, 1
	default:
		return scanned, 0, 1, 0
	}
}

// invokeResearchLit 本段固定 ARIS /research-lit，一次 tick 走完四步（只改 step.status，不写 runtime_state/extra 编排字段）。
func (j *IdeaDiscoveryPhasesJob) invokeResearchLit(ctx context.Context, step *entity.PaperOutputTopicStep, phaseID string) error {
	userID := uint(step.UserID)
	runVersion := step.RunVersion
	svc := j.runs
	cfg := topicdiscovery.MustIdeaDiscoveryConfig()

	input, err := svc.LoadRunInputForUser(ctx, userID, runVersion)
	if err != nil {
		return err
	}
	paperRoot := topicdiscovery.ARISRunPaperRoot(uint64(userID), step.ManuscriptID, runVersion)
	invokeCtx := svc.PhasedBotBuildInvokeContext(ctx, userID, runVersion, step, input)
	runID := topicdiscovery.PhasedBotResolveRunID(paperRoot, step, input)

	// 1. 准备 workspace：清 CLI session、落 RESEARCH_BRIEF/参数，供 /research-lit 读取。
	cfg.ClearCLIWorkspaceSessions(paperRoot)
	if err = topicdiscovery.PrepareNativeIdeaDiscoveryWorkspace(paperRoot, input, invokeCtx.Language, cfg); err != nil {
		return err
	}
	argsLine := cfg.FormatIdeaDiscoveryArgsLine(input, strings.TrimSpace(input.Venue), invokeCtx.Language)
	if err = topicdiscovery.WriteNativePlatformArguments(paperRoot, argsLine); err != nil {
		return err
	}

	// 2. invoke：按 paper_llm_strategy_route 打一次 /research-lit。
	if err = j.researchLitInvokeBehavior(ctx, step, svc, input, invokeCtx, paperRoot, cfg); err != nil {
		return err
	}

	// 3. 校验产物：research-lit 需有约定文件（如 LITERATURE_LANDSCAPE.md）；失败则 status=failed 或保持 running 下 tick 重试。
	if !topicdiscovery.IdeaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, phaseID, cfg) {
		if topicdiscovery.IdeaDiscoveryPhasedBotMayRetryPhase(cfg) {
			step.Status = topicStepRunning
			return j.steps.SaveStep(ctx, step)
		}
		reason := fmt.Sprintf("Idea Discovery 阶段 /%s 未产出约定产物", phaseID)
		return svc.PhasedBotPersistStageFailed(ctx, step, reason)
	}

	_ = svc.PhasedBotIngestLiterature(ctx, userID, runVersion, step.ManuscriptID, paperRoot, cfg)

	// 4. 更新 status：本段 completed；无段间 pause 则把下一段（idea_creator）标 running，否则下一段保持 pending 等 API 继续。
	now := time.Now()
	step.Status = topicStepCompleted
	step.ExecProgressPct = 100
	step.CompletedAt = &now
	if err := j.steps.SaveStep(ctx, step); err != nil {
		return err
	}
	if cfg.PhaseNeedsUserCheckpoint(phaseID) {
		return nil
	}
	return svc.PhasedBotBeginNextStageAsync(ctx, userID, runVersion, topicdiscovery.NextPhasedTopicStageCode(researchLitStageCode))
}

// researchLitInvokeBehavior 查 research_lit 的 strategy route，组装 Invoke 请求并走本地 ARIS CLI 一次。
func (j *IdeaDiscoveryPhasesJob) researchLitInvokeBehavior(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	svc *papersvc.TopicDiscoveryRunService,
	input topicdiscovery.TopicRunInput,
	invokeCtx topicdiscovery.StrategyInvokeContext,
	paperRoot string,
	cfg topicdiscovery.ARISIdeaDiscoveryConfig,
) error {
	behaviorKey := topicdiscovery.TopicStageBehaviorKey(researchLitStageCode)
	route, err := svc.PhasedBotRequireRoute(ctx, behaviorKey)
	if err != nil {
		return err
	}
	req := topicdiscovery.StrategyInvokeRequest{
		BehaviorKey:  behaviorKey,
		CapabilityID: strings.TrimSpace(route.CapabilityID),
		Context:      invokeCtx,
		Options: topicdiscovery.StrategyInvokeOptions{
			AuditLevelCode:   input.AuditLevel,
			IntensityCode:    input.Intensity,
			ExecRounds:       input.ExecRounds,
			MaxIdeas:         input.MaxIdeas,
			MaxPapers:        input.MaxPapers,
			SourceCodes:      input.SourceCodes,
			RandomReferences: input.RandomReferences,
		},
	}
	_, _, err = svc.PhasedBotInvokeBehavior(ctx, step, req, "")
	if err != nil {
		return svc.PhasedBotHandleInvokeFailure(ctx, step, paperRoot, cfg, err)
	}
	return nil
}

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
