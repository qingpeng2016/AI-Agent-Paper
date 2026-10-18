package ideadiscoveryphases

import (
	"context"
	"fmt"
	"strings"
	"time"

	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/lang"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/topicdiscovery"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"go.uber.org/zap"
)

// RunNoveltyCheck 对 running 的 novelty_check step 执行一次 /novelty-check ARIS。
func (j *IdeaDiscoveryPhasesJob) RunNoveltyCheck(ctx context.Context, step *entity.PaperOutputTopicStep) (scanned, invoked, skipped, failed int) {
	if step == nil || j.runs == nil {
		return 0, 0, 0, 0
	}
	if strings.TrimSpace(step.Status) != topicStepRunning ||
		strings.TrimSpace(step.StageCode) != noveltyCheckStageCode {
		return 0, 0, 0, 0
	}
	userID := uint(step.UserID)
	runVersion := step.RunVersion
	priorOK, err := j.noveltyCheckPriorCompleted(ctx, userID, runVersion)
	if err != nil {
		papersvc.FailTopicStep(step, err.Error())
		if saveErr := j.steps.SaveStep(ctx, step); saveErr != nil {
			logger.ErrorZ(ctx, "idea-discovery-phases-fail-save", zap.Uint64("step_id", step.ID), zap.Error(saveErr))
		}
		noveltyCheckNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed, err.Error(), 0, 0)
		return 1, 0, 0, 1
	}
	if !priorOK {
		return 0, 0, 0, 0
	}
	scanned = 1

	phaseID := topicdiscovery.TopicStageCodeToPhaseID(noveltyCheckStageCode)
	if err := j.invokeNoveltyCheck(ctx, step, phaseID); err != nil {
		if strings.TrimSpace(step.Status) != topicStepFailed {
			papersvc.FailTopicStep(step, err.Error())
			if saveErr := j.steps.SaveStep(ctx, step); saveErr != nil {
				logger.ErrorZ(ctx, "idea-discovery-phases-fail-save", zap.Uint64("step_id", step.ID), zap.Error(saveErr))
			}
		}
		noveltyCheckNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed, err.Error(), 0, 0)
		return scanned, 0, 0, 1
	}

	switch strings.TrimSpace(step.Status) {
	case topicStepCompleted:
		noveltyCheckNotify(j, ctx, step, papersvc.TopicStepOutcomeSuccess,
			fmt.Sprintf("Idea Discovery · %s 已完成", phaseID), 0, 0)
		if topicdiscovery.TopicStageNeedsUserCheckpoint(phaseID) {
			noveltyCheckNotify(j, ctx, step, papersvc.TopicStepOutcomeCheckpoint,
				topicdiscovery.TopicStageCheckpointDetail(phaseID), 0, 0)
		}
		return scanned, 1, 0, 0
	case topicStepFailed:
		return scanned, 0, 0, 1
	default:
		return scanned, 0, 1, 0
	}
}

func (j *IdeaDiscoveryPhasesJob) noveltyCheckPriorCompleted(ctx context.Context, userID uint, runVersion int) (bool, error) {
	for _, code := range noveltyCheckPriorStages {
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

type noveltyCheckInvokePayload struct {
	Req              topicdiscovery.StrategyInvokeRequest
	PlatformArgsLine string
	Sources          string
	ArxivDownload    bool
	ArxivMaxDownload int
	Venue            string
	Language         string
}

// invokeNoveltyCheck 本段固定 ARIS /novelty-check，一次 tick 走完四步（只改 step.status，不写 runtime_state/extra 编排字段）。
func (j *IdeaDiscoveryPhasesJob) invokeNoveltyCheck(ctx context.Context, step *entity.PaperOutputTopicStep, phaseID string) error {
	userID := uint(step.UserID)
	runVersion := step.RunVersion
	svc := j.runs
	cfg := topicdiscovery.MustIdeaDiscoveryConfig()

	input, err := svc.LoadRunInputForUser(ctx, userID, runVersion)
	if err != nil {
		return err
	}
	behaviorKey := topicdiscovery.TopicStageBehaviorKey(noveltyCheckStageCode)
	route, err := svc.PhasedBotRequireRoute(ctx, behaviorKey)
	if err != nil {
		return err
	}
	locale := svc.PhasedBotBuildInvokeContext(ctx, userID, runVersion, step, input).Language
	payload := noveltyCheckBuildInvokePayload(step, userID, runVersion, input, route, cfg, locale)

	paperRoot := topicdiscovery.ARISRunPaperRoot(uint64(userID), step.ManuscriptID, runVersion)
	runID := topicdiscovery.PhasedBotResolveRunID(paperRoot, step, input)

	// 1. 准备 workspace：清 CLI session、落 RESEARCH_BRIEF/参数，供 /novelty-check 读取。
	cfg.ClearCLIWorkspaceSessions(paperRoot)
	if err = topicdiscovery.PrepareNativeIdeaDiscoveryWorkspace(paperRoot, input, payload.Req.Context.Language, cfg); err != nil {
		return err
	}
	if err = topicdiscovery.WriteNativePlatformArguments(paperRoot, payload.PlatformArgsLine); err != nil {
		return err
	}

	// 2. invoke：req 与 platform 参数均来自 payload。
	if err = j.noveltyCheckInvokeBehavior(ctx, step, svc, paperRoot, cfg, payload); err != nil {
		return err
	}

	// 3. 校验产物：novelty-check 需有约定文件；失败则 status=failed 或保持 running 下 tick 重试。
	if !topicdiscovery.IdeaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, phaseID, cfg) {
		if topicdiscovery.IdeaDiscoveryPhasedBotMayRetryPhase(cfg) {
			step.Status = topicStepRunning
			return j.steps.SaveStep(ctx, step)
		}
		reason := fmt.Sprintf("Idea Discovery 阶段 /%s 未产出约定产物", phaseID)
		return svc.PhasedBotPersistStageFailed(ctx, step, reason)
	}

	// 4. 更新 status：本段 completed；无段间 pause 则把下一段（research_review）标 running，否则下一段保持 pending 等 API 继续。
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
	return svc.PhasedBotBeginNextStageAsync(ctx, userID, runVersion, topicdiscovery.NextPhasedTopicStageCode(noveltyCheckStageCode))
}

func (j *IdeaDiscoveryPhasesJob) noveltyCheckInvokeBehavior(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	svc *papersvc.TopicDiscoveryRunService,
	paperRoot string,
	cfg topicdiscovery.ARISIdeaDiscoveryConfig,
	payload noveltyCheckInvokePayload,
) error {
	_, _, err := svc.PhasedBotInvokeBehavior(ctx, step, payload.Req, payload.PlatformArgsLine)
	if err != nil {
		return svc.PhasedBotHandleInvokeFailure(ctx, step, paperRoot, cfg, err)
	}
	return nil
}

func noveltyCheckBuildInvokePayload(
	step *entity.PaperOutputTopicStep,
	userID uint,
	runVersion int,
	input topicdiscovery.TopicRunInput,
	route *entity.PaperStrategyRoute,
	cfg topicdiscovery.ARISIdeaDiscoveryConfig,
	language string,
) noveltyCheckInvokePayload {
	behaviorKey := topicdiscovery.TopicStageBehaviorKey(noveltyCheckStageCode)

	sources := strings.TrimSpace(cfg.Platform.Sources)
	if sources == "" {
		sources = "web"
	}
	arxivDownload := cfg.Platform.ArxivDownload
	arxivDownloadCLI := "false"
	if arxivDownload {
		arxivDownloadCLI = "true"
	}
	arxivMaxDownload := cfg.Platform.ArxivMaxDownload
	if arxivMaxDownload <= 0 {
		arxivMaxDownload = 5
	}
	venue := strings.TrimSpace(input.Venue)
	if venue == "" {
		venue = "NeurIPS"
	}
	language = lang.NormalizeContentLanguage(language)
	if language == "" {
		language = "zh"
	}

	platformArgs := fmt.Sprintf(
		`%s — venue: %s — sources: %s — ARXIV_DOWNLOAD: %s`,
		noveltyCheckQuotedDirection(input),
		venue,
		sources,
		arxivDownloadCLI,
	)
	if arxivDownload {
		platformArgs += fmt.Sprintf(` — ARXIV_MAX_DOWNLOAD: %d`, arxivMaxDownload)
	}
	platformArgs += fmt.Sprintf(` — language: %s`, language)

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

	return noveltyCheckInvokePayload{
		Req:              req,
		PlatformArgsLine: platformArgs,
		Sources:          sources,
		ArxivDownload:    arxivDownload,
		ArxivMaxDownload: arxivMaxDownload,
		Venue:            venue,
		Language:         language,
	}
}

func noveltyCheckQuotedDirection(input topicdiscovery.TopicRunInput) string {
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
