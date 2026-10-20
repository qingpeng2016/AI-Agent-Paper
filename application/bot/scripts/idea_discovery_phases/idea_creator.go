package ideadiscoveryphases

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/lang"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/topicdiscovery"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"go.uber.org/zap"
)

// RunIdeaCreator 对 running 的 idea_creator step 执行一次 /idea-creator ARIS。
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
		papersvc.FailTopicStep(step, err.Error())
		if saveErr := j.steps.SaveStep(ctx, step); saveErr != nil {
			logger.ErrorZ(ctx, "idea-discovery-phases-fail-save", zap.Uint64("step_id", step.ID), zap.Error(saveErr))
		}
		ideaCreatorNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed, err.Error(), 0, 0)
		return 1, 0, 0, 1
	}
	if !priorOK {
		return 0, 0, 0, 0
	}
	scanned = 1

	phaseID := topicdiscovery.TopicStageCodeToPhaseID(ideaCreatorStageCode)
	if err := j.invokeIdeaCreator(ctx, step, phaseID); err != nil {
		if strings.TrimSpace(step.Status) != topicStepFailed {
			papersvc.FailTopicStep(step, err.Error())
			if saveErr := j.steps.SaveStep(ctx, step); saveErr != nil {
				logger.ErrorZ(ctx, "idea-discovery-phases-fail-save", zap.Uint64("step_id", step.ID), zap.Error(saveErr))
			}
		}
		ideaCreatorNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed, err.Error(), 0, 0)
		return scanned, 0, 0, 1
	}

	switch strings.TrimSpace(step.Status) {
	case topicStepCompleted:
		ideaCreatorNotify(j, ctx, step, papersvc.TopicStepOutcomeSuccess,
			fmt.Sprintf("Idea Discovery · %s 已完成", phaseID), 0, 0)
		if topicdiscovery.TopicStageNeedsUserCheckpoint(phaseID) {
			ideaCreatorNotify(j, ctx, step, papersvc.TopicStepOutcomeCheckpoint,
				topicdiscovery.TopicStageCheckpointDetail(phaseID), 0, 0)
		}
		return scanned, 1, 0, 0
	case topicStepFailed:
		return scanned, 0, 0, 1
	default:
		return scanned, 0, 1, 0
	}
}

func (j *IdeaDiscoveryPhasesJob) ideaCreatorPriorCompleted(ctx context.Context, userID uint, runVersion int) (bool, error) {
	for _, code := range ideaCreatorPriorStages {
		st, err := j.steps.GetStep(ctx, uint64(userID), runVersion, code)
		if err != nil {
			return false, err
		}
		if st == nil || strings.TrimSpace(st.Status) != topicStepCompleted {
			return false, nil
		}
	}
	return true, nil
}

type ideaCreatorInvokePayload struct {
	Req              topicdiscovery.StrategyInvokeRequest
	PlatformArgsLine string
	Language         string
}

// invokeIdeaCreator 本段固定 ARIS /idea-creator，一次 tick 走完四步（只改 step.status，不写 runtime_state/extra 编排字段）。
func (j *IdeaDiscoveryPhasesJob) invokeIdeaCreator(ctx context.Context, step *entity.PaperOutputTopicStep, phaseID string) error {
	userID := uint(step.UserID)
	runVersion := step.RunVersion
	svc := j.runs
	cfg := topicdiscovery.MustIdeaDiscoveryConfig()

	input, err := svc.LoadRunInputForUser(ctx, userID, runVersion)
	if err != nil {
		return err
	}
	behaviorKey := topicdiscovery.TopicStageBehaviorKey(ideaCreatorStageCode)
	route, err := svc.PhasedBotRequireRoute(ctx, behaviorKey)
	if err != nil {
		return err
	}
	paperRoot := topicdiscovery.ARISRunPaperRoot(uint64(userID), step.ManuscriptID, runVersion)
	locale := svc.PhasedBotBuildInvokeContext(ctx, userID, runVersion, step, input).Language
	payload := ideaCreatorBuildInvokePayload(step, userID, runVersion, input, route, cfg, locale, paperRoot)
	runID := topicdiscovery.PhasedBotResolveRunID(paperRoot, step, input)

	// 1. 准备 workspace：清 CLI session、落 RESEARCH_BRIEF/参数，供 /idea-creator 读取。
	cfg.ClearCLIWorkspaceSessions(paperRoot)
	if err = topicdiscovery.PrepareNativeIdeaDiscoveryWorkspace(paperRoot, input, payload.Req.Context.Language, cfg); err != nil {
		return err
	}
	if err = topicdiscovery.WriteNativePlatformArguments(paperRoot, payload.PlatformArgsLine); err != nil {
		return err
	}

	// 2. invoke：req 与 platform 参数均来自 payload。
	if err = j.ideaCreatorInvokeBehavior(ctx, step, svc, paperRoot, cfg, payload); err != nil {
		return err
	}

	// 3. 校验产物：idea-creator 需有约定文件；失败则 status=failed 或保持 running 下 tick 重试。
	if !topicdiscovery.IdeaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, phaseID, cfg) {
		if topicdiscovery.IdeaDiscoveryPhasedBotMayRetryPhase(cfg) {
			step.Status = topicStepRunning
			return j.steps.SaveStep(ctx, step)
		}
		reason := fmt.Sprintf("Idea Discovery 阶段 /%s 未产出约定产物", phaseID)
		return svc.PhasedBotPersistStageFailed(ctx, step, reason)
	}

	// 4. 更新 status：本段 completed；无段间 pause 则把下一段（novelty_check）标 running，否则下一段保持 pending 等 API 继续。
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
	return svc.PhasedBotBeginNextStageAsync(ctx, userID, runVersion, topicdiscovery.NextPhasedTopicStageCode(ideaCreatorStageCode))
}

func (j *IdeaDiscoveryPhasesJob) ideaCreatorInvokeBehavior(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	svc *papersvc.TopicDiscoveryRunService,
	paperRoot string,
	cfg topicdiscovery.ARISIdeaDiscoveryConfig,
	payload ideaCreatorInvokePayload,
) error {
	_, _, err := svc.PhasedBotInvokeBehavior(ctx, step, payload.Req, payload.PlatformArgsLine)
	if err != nil {
		return svc.PhasedBotHandleInvokeFailure(ctx, step, paperRoot, cfg, err)
	}
	return nil
}

func ideaCreatorBuildInvokePayload(
	step *entity.PaperOutputTopicStep,
	userID uint,
	runVersion int,
	input topicdiscovery.TopicRunInput,
	route *entity.PaperStrategyRoute,
	cfg topicdiscovery.ARISIdeaDiscoveryConfig,
	language string,
	paperRoot string,
) ideaCreatorInvokePayload {
	behaviorKey := topicdiscovery.TopicStageBehaviorKey(ideaCreatorStageCode)

	litDir := strings.TrimSpace(cfg.Artifacts.LiteratureDir)
	if litDir == "" {
		litDir = "literature"
	}
	paperLibrary := filepath.Join(paperRoot, litDir)
	if !strings.HasSuffix(paperLibrary, string(filepath.Separator)) {
		paperLibrary += string(filepath.Separator)
	}

	language = lang.NormalizeContentLanguage(language)
	if language == "" {
		language = "zh"
	}

	platformArgs := fmt.Sprintf(
		`%s — paper library: %s — language: %s`,
		ideaCreatorQuotedDirection(input),
		paperLibrary,
		language,
	)

	req := topicdiscovery.StrategyInvokeRequest{
		BehaviorKey:  behaviorKey,
		CapabilityID: strings.TrimSpace(route.CapabilityID),
		Context: topicdiscovery.StrategyInvokeContext{
			ManuscriptID: step.ManuscriptID,
			UserID:       uint64(userID),
			RunVersion:   runVersion,
			Language:     language,
			Direction:    strings.TrimSpace(input.Direction),
			Keywords:     input.Keywords,
			VenueCodes:   input.VenueCodes,
			VenueLabel:   strings.TrimSpace(input.Venue),
			Round:        1,
			MaxRounds:    1,
		},
		Options: topicdiscovery.StrategyInvokeOptions{
			AuditLevelCode:   strings.TrimSpace(input.AuditLevel),
			IntensityCode:    strings.TrimSpace(input.Intensity),
			ExecRounds:       input.ExecRounds,
			MaxIdeas:         input.MaxIdeas,
			MaxPapers:        input.MaxPapers,
			SourceCodes:      input.SourceCodes,
			RandomReferences: input.RandomReferences,
		},
	}

	return ideaCreatorInvokePayload{
		Req:              req,
		PlatformArgsLine: platformArgs,
		Language:         language,
	}
}

func ideaCreatorQuotedDirection(input topicdiscovery.TopicRunInput) string {
	inner := strings.TrimSpace(input.Description)
	if inner == "" {
		body := strings.TrimSpace(input.Direction)
		kw := strings.Join(input.Keywords, ", ")
		kw = strings.TrimSpace(kw)
		if kw != "" {
			inner = fmt.Sprintf("研究内容：%s + 关键词：%s", body, kw)
		} else {
			inner = body
		}
	}
	inner = strings.ReplaceAll(inner, `"`, `\"`)
	return `"` + inner + `"`
}

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
