package ideadiscoveryphases

import (
	"context"
	"fmt"
	"strings"

	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/lang"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/topicdiscovery"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"go.uber.org/zap"
)

// RunResearchRefine 对 running 的 research_refine step 执行一次 /research-refine-pipeline ARIS。
func (j *IdeaDiscoveryPhasesJob) RunResearchRefine(ctx context.Context, step *entity.PaperOutputTopicStep) (scanned, invoked, skipped, failed int) {
	if step == nil || j.runs == nil {
		return 0, 0, 0, 0
	}
	if strings.TrimSpace(step.Status) != topicStepRunning ||
		strings.TrimSpace(step.StageCode) != researchRefineStageCode {
		return 0, 0, 0, 0
	}
	userID := uint(step.UserID)
	runVersion := step.RunVersion
	priorOK, err := j.researchRefinePriorCompleted(ctx, userID, runVersion)
	if err != nil {
		papersvc.FailTopicStep(step, err.Error())
		if saveErr := j.steps.SaveStep(ctx, step); saveErr != nil {
			logger.ErrorZ(ctx, "idea-discovery-phases-fail-save", zap.Uint64("step_id", step.ID), zap.Error(saveErr))
		}
		researchRefineNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed, err.Error(), 0, 0)
		return 1, 0, 0, 1
	}
	if !priorOK {
		return 0, 0, 0, 0
	}
	scanned = 1

	phaseID := topicdiscovery.TopicStageCodeToPhaseID(researchRefineStageCode)
	if err := j.invokeResearchRefine(ctx, step, phaseID); err != nil {
		if strings.TrimSpace(step.Status) != topicStepFailed {
			papersvc.FailTopicStep(step, err.Error())
			if saveErr := j.steps.SaveStep(ctx, step); saveErr != nil {
				logger.ErrorZ(ctx, "idea-discovery-phases-fail-save", zap.Uint64("step_id", step.ID), zap.Error(saveErr))
			}
		}
		researchRefineNotify(j, ctx, step, papersvc.TopicStepOutcomeFailed, err.Error(), 0, 0)
		return scanned, 0, 0, 1
	}

	switch strings.TrimSpace(step.Status) {
	case topicStepCompleted:
		researchRefineNotify(j, ctx, step, papersvc.TopicStepOutcomeSuccess,
			fmt.Sprintf("Idea Discovery · %s 已完成", phaseID), 0, 0)
		if topicdiscovery.TopicStageNeedsUserCheckpoint(phaseID) {
			researchRefineNotify(j, ctx, step, papersvc.TopicStepOutcomeCheckpoint,
				topicdiscovery.TopicStageCheckpointDetail(phaseID), 0, 0)
		}
		return scanned, 1, 0, 0
	case topicStepFailed:
		return scanned, 0, 0, 1
	default:
		return scanned, 0, 1, 0
	}
}

func (j *IdeaDiscoveryPhasesJob) researchRefinePriorCompleted(ctx context.Context, userID uint, runVersion int) (bool, error) {
	for _, code := range researchRefinePriorStages {
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

type researchRefineInvokePayload struct {
	Req              topicdiscovery.StrategyInvokeRequest
	PlatformArgsLine string
	Sources          string
	ArxivDownload    bool
	ArxivMaxDownload int
	Venue            string
	Language         string
}

// invokeResearchRefine 本段固定 ARIS /research-refine-pipeline，一次 tick 走完四步（只改 step.status，不写 runtime_state/extra 编排字段）。
func (j *IdeaDiscoveryPhasesJob) invokeResearchRefine(ctx context.Context, step *entity.PaperOutputTopicStep, phaseID string) error {
	userID := uint(step.UserID)
	runVersion := step.RunVersion
	svc := j.runs
	cfg := topicdiscovery.MustIdeaDiscoveryConfig()

	input, err := svc.LoadRunInputForUser(ctx, userID, runVersion)
	if err != nil {
		return err
	}
	behaviorKey := topicdiscovery.TopicStageBehaviorKey(researchRefineStageCode)
	route, err := svc.PhasedBotRequireRoute(ctx, behaviorKey)
	if err != nil {
		return err
	}
	locale := svc.PhasedBotBuildInvokeContext(ctx, userID, runVersion, step, input).Language
	payload := researchRefineBuildInvokePayload(step, userID, runVersion, input, route, cfg, locale)

	paperRoot := topicdiscovery.ARISRunPaperRoot(uint64(userID), step.ManuscriptID, runVersion)
	runID := topicdiscovery.PhasedBotResolveRunID(paperRoot, step, input)

	// 1. 准备 workspace：清 CLI session、落 RESEARCH_BRIEF/参数，供 /research-refine-pipeline 读取。
	cfg.ClearCLIWorkspaceSessions(paperRoot)
	if err = topicdiscovery.PrepareNativeIdeaDiscoveryWorkspace(paperRoot, input, payload.Req.Context.Language, cfg); err != nil {
		return err
	}
	if err = topicdiscovery.WriteNativePlatformArguments(paperRoot, payload.PlatformArgsLine); err != nil {
		return err
	}

	// 2. invoke：req 与 platform 参数均来自 payload。
	if err = j.researchRefineInvokeBehavior(ctx, step, svc, paperRoot, cfg, payload); err != nil {
		return err
	}

	// 3. 校验产物：research-refine 需有约定文件；失败则 status=failed 或保持 running 下 tick 重试。
	if !topicdiscovery.IdeaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, phaseID, cfg) {
		if topicdiscovery.IdeaDiscoveryPhasedBotMayRetryPhase(cfg) {
			step.Status = topicStepRunning
			return j.steps.SaveStep(ctx, step)
		}
		reason := fmt.Sprintf("Idea Discovery 阶段 /%s 未产出约定产物", phaseID)
		return svc.PhasedBotPersistStageFailed(ctx, step, reason)
	}

	// 4. 更新 status：五段末段，走 idea_discovery gate（topic_confirm）收尾并标本段 completed。
	return svc.PhasedBotFinishAfterGate(ctx, step, userID, runVersion, route, paperRoot, runID, cfg)
}

func (j *IdeaDiscoveryPhasesJob) researchRefineInvokeBehavior(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	svc *papersvc.TopicDiscoveryRunService,
	paperRoot string,
	cfg topicdiscovery.ARISIdeaDiscoveryConfig,
	payload researchRefineInvokePayload,
) error {
	_, _, err := svc.PhasedBotInvokeBehavior(ctx, step, payload.Req, payload.PlatformArgsLine)
	if err != nil {
		return svc.PhasedBotHandleInvokeFailure(ctx, step, paperRoot, cfg, err)
	}
	return nil
}

func researchRefineBuildInvokePayload(
	step *entity.PaperOutputTopicStep,
	userID uint,
	runVersion int,
	input topicdiscovery.TopicRunInput,
	route *entity.PaperStrategyRoute,
	cfg topicdiscovery.ARISIdeaDiscoveryConfig,
	language string,
) researchRefineInvokePayload {
	behaviorKey := topicdiscovery.TopicStageBehaviorKey(researchRefineStageCode)

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
		researchRefineQuotedDirection(input),
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

	return researchRefineInvokePayload{
		Req:              req,
		PlatformArgsLine: platformArgs,
		Sources:          sources,
		ArxivDownload:    arxivDownload,
		ArxivMaxDownload: arxivMaxDownload,
		Venue:            venue,
		Language:         language,
	}
}

func researchRefineQuotedDirection(input topicdiscovery.TopicRunInput) string {
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
