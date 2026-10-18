package ideadiscoveryphases

import (
	"context"
	"fmt"
	"os"
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

// researchLitInvokePayload research_lit 单次 ARIS：StrategyInvokeRequest + CLI 参数字符串（本段所有开关在此显式赋值）。
type researchLitInvokePayload struct {
	Req topicdiscovery.StrategyInvokeRequest

	// PlatformArgsLine 写入 .paper/invoke/platform_user_prompt.txt；CLI 为 `/research-lit` + 本行。
	PlatformArgsLine string

	Sources          string
	ArxivDownload    bool
	ArxivMaxDownload int
	Venue            string
	Language         string
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
	behaviorKey := topicdiscovery.TopicStageBehaviorKey(researchLitStageCode)
	route, err := svc.PhasedBotRequireRoute(ctx, behaviorKey)
	if err != nil {
		return err
	}
	locale := svc.PhasedBotBuildInvokeContext(ctx, userID, runVersion, step, input).Language
	payload := researchLitBuildInvokePayload(step, userID, runVersion, input, route, cfg, locale)

	paperRoot := topicdiscovery.ARISRunPaperRoot(uint64(userID), step.ManuscriptID, runVersion)

	// 1. 准备 workspace：清 CLI session、落 RESEARCH_BRIEF/参数，供 /research-lit 读取。
	cfg.ClearCLIWorkspaceSessions(paperRoot)
	if err = topicdiscovery.PrepareNativeIdeaDiscoveryWorkspace(paperRoot, input, payload.Req.Context.Language, cfg); err != nil {
		return err
	}
	if err = topicdiscovery.WriteNativePlatformArguments(paperRoot, payload.PlatformArgsLine); err != nil {
		return err
	}

	// 2. invoke：req 与 platform 参数均来自 payload（见 researchLitBuildInvokePayload）。
	if err = j.researchLitInvokeBehavior(ctx, step, svc, paperRoot, cfg, payload); err != nil {
		return err
	}

	// 3. 校验产物：paperRoot/literature（见 yaml literature_dir）下至少有一个 PDF，否则本段 failed。
	if err = researchLitRequireLiteraturePDF(paperRoot, cfg); err != nil {
		papersvc.FailTopicStep(step, err.Error())
		if saveErr := j.steps.SaveStep(ctx, step); saveErr != nil {
			return saveErr
		}
		return err
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

// researchLitBuildInvokePayload 在本文件内显式组装 req 与 /research-lit CLI 参数（默认值来自 conf/aris_idea_discovery.yaml，赋值处可见）。
func researchLitBuildInvokePayload(
	step *entity.PaperOutputTopicStep,
	userID uint,
	runVersion int,
	input topicdiscovery.TopicRunInput,
	route *entity.PaperStrategyRoute,
	cfg topicdiscovery.ARISIdeaDiscoveryConfig,
	language string,
) researchLitInvokePayload {
	behaviorKey := topicdiscovery.TopicStageBehaviorKey(researchLitStageCode)

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
		researchLitQuotedDirection(input),
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

	return researchLitInvokePayload{
		Req:              req,
		PlatformArgsLine: platformArgs,
		Sources:          sources,
		ArxivDownload:    arxivDownload,
		ArxivMaxDownload: arxivMaxDownload,
		Venue:            venue,
		Language:         language,
	}
}

func researchLitQuotedDirection(input topicdiscovery.TopicRunInput) string {
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

// researchLitRequireLiteraturePDF 校验 …/paper/<literature_dir>/ 下存在至少一个 .pdf 文件。
func researchLitRequireLiteraturePDF(paperRoot string, cfg topicdiscovery.ARISIdeaDiscoveryConfig) error {
	paperRoot = strings.TrimSpace(paperRoot)
	litDir := strings.TrimSpace(cfg.Artifacts.LiteratureDir)
	if litDir == "" {
		litDir = "literature"
	}
	litPath := filepath.Join(paperRoot, litDir)
	entries, err := os.ReadDir(litPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Idea Discovery /research-lit 未产出文献 PDF（目录不存在: %s）", litPath)
		}
		return fmt.Errorf("校验文献目录 %s: %w", litPath, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(e.Name()), ".pdf") {
			return nil
		}
	}
	return fmt.Errorf("Idea Discovery /research-lit 未产出文献 PDF（%s 下无 .pdf 文件）", litPath)
}

// researchLitInvokeBehavior 使用 payload.Req 走本地 ARIS CLI（platform 行已在 payload.PlatformArgsLine）。
func (j *IdeaDiscoveryPhasesJob) researchLitInvokeBehavior(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	svc *papersvc.TopicDiscoveryRunService,
	paperRoot string,
	cfg topicdiscovery.ARISIdeaDiscoveryConfig,
	payload researchLitInvokePayload,
) error {
	_, _, err := svc.PhasedBotInvokeBehavior(ctx, step, payload.Req, payload.PlatformArgsLine)
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
