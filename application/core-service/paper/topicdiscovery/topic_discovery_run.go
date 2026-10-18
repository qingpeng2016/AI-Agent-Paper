package topicdiscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/internal/jsonx"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/lang"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/literature"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/llm"
	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/userliterature"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/request"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
	"go.uber.org/zap"
)

type topicRunInput struct {
	DisciplineCode  string   `json:"discipline_code"`
	Keywords        []string `json:"keywords"`
	Description     string   `json:"description"`
	Direction       string   `json:"direction"`
	Venue           string   `json:"venue"`
	VenueCodes      []string `json:"venue_codes"`
	SourceCodes     []string `json:"source_codes"`
	Intensity       string   `json:"intensity"`
	AuditLevel      string   `json:"audit_level"`
	HumanCheckpoint  bool `json:"human_checkpoint"`
	RandomReferences  bool     `json:"random_references"`
	UserLiteratureIDs []uint64 `json:"user_literature_ids,omitempty"`
	MaxPapers         int      `json:"max_papers"`
	MaxIdeas        int      `json:"max_ideas"`
	ExecRounds      int      `json:"exec_rounds"`
	InvokeMode      string   `json:"invoke_mode,omitempty"`
}

type storedLiteratureHit = literature.StoredLiteratureHit

type storedLiteratureLink struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	ExternalKey string `json:"external_key,omitempty"`
	SourceCode  string `json:"source_code,omitempty"`
}

type TopicDiscoveryRunService struct {
	manuscripts repository.PaperManuscriptRepo
	steps       repository.PaperOutputTopicStepRepo
	litReviews  repository.PaperOutputLiteratureReviewRepo
	expPlans    repository.PaperOutputExperimentPlanRepo
	catalog     repository.PaperRefCatalogRepo
	litSource   repository.PaperLiteraturePlatformRepo
	litSearch    *literature.LiteratureSearchService
	llmRepo      repository.PaperLLMRepo
	strategyRepo repository.PaperStrategyRepo
	userLit      repository.PaperUserLiteratureRepo
}

func NewTopicDiscoveryRunService(
	manuscripts repository.PaperManuscriptRepo,
	steps repository.PaperOutputTopicStepRepo,
	litReviews repository.PaperOutputLiteratureReviewRepo,
	expPlans repository.PaperOutputExperimentPlanRepo,
	catalog repository.PaperRefCatalogRepo,
	litSource repository.PaperLiteraturePlatformRepo,
	litSearch *literature.LiteratureSearchService,
	llmRepo repository.PaperLLMRepo,
	strategyRepo repository.PaperStrategyRepo,
	userLit repository.PaperUserLiteratureRepo,
) *TopicDiscoveryRunService {
	return &TopicDiscoveryRunService{
		manuscripts:  manuscripts,
		steps:        steps,
		litReviews:   litReviews,
		expPlans:     expPlans,
		catalog:      catalog,
		litSource:    litSource,
		litSearch:    litSearch,
		llmRepo:      llmRepo,
		strategyRepo: strategyRepo,
		userLit:      userLit,
	}
}

func (s *TopicDiscoveryRunService) GetCurrentRun(ctx context.Context, userID uint, manuscriptID uint64) (*response.TopicDiscoveryRunView, error) {
	runVersion, stepRows, err := s.latestRunForManuscript(ctx, userID, manuscriptID)
	if err != nil {
		return nil, err
	}
	if runVersion == 0 || len(stepRows) == 0 || currentRunCancelled(stepRows) {
		return nil, nil
	}
	return s.buildRunView(manuscriptIDOf(stepRows), runVersion, stepRows), nil
}

func (s *TopicDiscoveryRunService) CancelCurrentRun(ctx context.Context, userID uint, manuscriptID uint64) error {
	runVersion, stepRows, err := s.latestRunForManuscript(ctx, userID, manuscriptID)
	if err != nil {
		return err
	}
	if runVersion == 0 || len(stepRows) == 0 || currentRunCancelled(stepRows) {
		return nil
	}
	return s.steps.CancelRun(ctx, uint64(userID), runVersion)
}

func (s *TopicDiscoveryRunService) latestRunForManuscript(ctx context.Context, userID uint, manuscriptID uint64) (int, []entity.PaperOutputTopicStep, error) {
	if manuscriptID > 0 {
		return s.steps.GetLatestRunByManuscript(ctx, uint64(userID), manuscriptID)
	}
	return s.steps.GetLatestRunByUser(ctx, uint64(userID))
}

func (s *TopicDiscoveryRunService) validateManuscriptForUser(ctx context.Context, userID uint, manuscriptID uint64) error {
	if manuscriptID == 0 {
		return errorx.ErrParamsError.WithDetail("缺少 manuscript_id，请先创建论文")
	}
	ms, err := s.manuscripts.GetByIDForUser(ctx, uint(manuscriptID), userID)
	if err != nil {
		return err
	}
	if ms == nil {
		return errorx.ErrParamsError.WithDetail("论文不存在或无权访问")
	}
	return nil
}

func manuscriptIDOf(rows []entity.PaperOutputTopicStep) uint64 {
	if len(rows) == 0 {
		return 0
	}
	return rows[0].ManuscriptID
}

func currentRunCancelled(rows []entity.PaperOutputTopicStep) bool {
	if len(rows) == 0 {
		return false
	}
	for _, r := range rows {
		if r.Status != "cancelled" {
			return false
		}
	}
	return true
}

func (s *TopicDiscoveryRunService) Run(ctx context.Context, userID uint, req request.TopicDiscoveryRunRequest) (*response.TopicDiscoveryRunView, error) {
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if action == "" {
		action = "start"
	}

	var runVersion int
	var input topicRunInput
	var err error

	switch action {
	case "start":
		if err := s.validateManuscriptForUser(ctx, userID, req.ManuscriptID); err != nil {
			return nil, err
		}
		input, err = s.prepareRun(ctx, userID, req)
		if err != nil {
			return nil, err
		}
		paramsBytes, _ := json.Marshal(input)
		if strings.EqualFold(strings.TrimSpace(input.InvokeMode), "phased") {
			runVersion, err = s.steps.BeginPhasedRun(ctx, uint64(userID), req.ManuscriptID, paramsBytes)
		} else {
			runVersion, err = s.steps.BeginRun(ctx, uint64(userID), req.ManuscriptID, paramsBytes)
		}
		if err != nil {
			return nil, err
		}
		if err := s.executeFromStage(ctx, userID, runVersion, input, 0); err != nil {
			return s.reloadViewOrErr(ctx, userID, runVersion, err)
		}
	case "continue":
		if err := s.validateManuscriptForUser(ctx, userID, req.ManuscriptID); err != nil {
			return nil, err
		}
		ver, rows, err := s.latestRunForManuscript(ctx, userID, req.ManuscriptID)
		if err != nil {
			return nil, err
		}
		if ver == 0 || currentRunCancelled(rows) {
			return nil, errorx.ErrTopicRunNotFound
		}
		runVersion = ver
		input, err = s.loadRunInput(ctx, userID, runVersion)
		if err != nil {
			return nil, err
		}
		if resumed, err := s.ResumeIdeaDiscoveryFromCheckpoint(ctx, userID, runVersion); err != nil {
			return nil, err
		} else if resumed {
			return s.reloadView(ctx, userID, runVersion)
		}
		idx := s.nextPendingIndex(ctx, userID, runVersion)
		if idx < 0 {
			return s.reloadView(ctx, userID, runVersion)
		}
		if st, _ := s.runningStageCode(ctx, userID, runVersion); st != "" {
			return s.reloadView(ctx, userID, runVersion)
		}
		if err := s.executeFromStage(ctx, userID, runVersion, input, idx); err != nil {
			return s.reloadViewOrErr(ctx, userID, runVersion, err)
		}
	case "resume":
		if err := s.validateManuscriptForUser(ctx, userID, req.ManuscriptID); err != nil {
			return nil, err
		}
		ver, rows, err := s.latestRunForManuscript(ctx, userID, req.ManuscriptID)
		if err != nil {
			return nil, err
		}
		if ver == 0 || currentRunCancelled(rows) {
			return nil, errorx.ErrTopicRunNotFound
		}
		runVersion = ver
		resumed, err := s.ResumeIdeaDiscoveryExecution(ctx, userID, runVersion)
		if err != nil {
			return nil, err
		}
		if !resumed {
			return nil, errorx.ErrParamsError.WithDetail("当前 run 无可续跑的 idea_discovery（已完成或失败）")
		}
		return s.reloadView(ctx, userID, runVersion)
	case "run_all":
		if err := s.validateManuscriptForUser(ctx, userID, req.ManuscriptID); err != nil {
			return nil, err
		}
		ver, rows, err := s.latestRunForManuscript(ctx, userID, req.ManuscriptID)
		if err != nil {
			return nil, err
		}
		if ver == 0 || currentRunCancelled(rows) {
			input, err = s.prepareRun(ctx, userID, req)
			if err != nil {
				return nil, err
			}
			paramsBytes, _ := json.Marshal(input)
			if strings.EqualFold(strings.TrimSpace(input.InvokeMode), "phased") {
				runVersion, err = s.steps.BeginPhasedRun(ctx, uint64(userID), req.ManuscriptID, paramsBytes)
			} else {
				runVersion, err = s.steps.BeginRun(ctx, uint64(userID), req.ManuscriptID, paramsBytes)
			}
			if err != nil {
				return nil, err
			}
		} else {
			runVersion = ver
			input, err = s.loadRunInput(ctx, userID, runVersion)
			if err != nil {
				return nil, err
			}
		}
		for {
			idx := s.nextPendingIndex(ctx, userID, runVersion)
			if idx < 0 {
				break
			}
			stages := s.stagesForRun(ctx, userID, runVersion)
			if idx >= len(stages) {
				break
			}
			if err := s.executeStage(ctx, userID, runVersion, input, stages[idx]); err != nil {
				return s.reloadViewOrErr(ctx, userID, runVersion, err)
			}
			if st, _ := s.runningStageCode(ctx, userID, runVersion); st != "" {
				return s.reloadView(ctx, userID, runVersion)
			}
			if input.HumanCheckpoint {
				view, err := s.reloadView(ctx, userID, runVersion)
				if err != nil {
					return nil, err
				}
				if view.PauseAfterStage != "" {
					return view, nil
				}
			}
		}
	default:
		return nil, errorx.ErrParamsError
	}

	return s.reloadView(ctx, userID, runVersion)
}

func (s *TopicDiscoveryRunService) defaultExecutionIntensityCode(ctx context.Context) string {
	rows, err := s.catalog.ListExecutionIntensities(ctx)
	if err == nil {
		for _, r := range rows {
			if r.IsDefault {
				return r.Code
			}
		}
		if len(rows) > 0 {
			return rows[0].Code
		}
	}
	return "balanced"
}

func (s *TopicDiscoveryRunService) defaultAuditLevelCode(ctx context.Context) string {
	rows, err := s.catalog.ListAuditLevels(ctx, entity.LLMSceneTopic)
	if err == nil {
		for _, r := range rows {
			if r.IsDefault {
				return r.Code
			}
		}
		if len(rows) > 0 {
			return rows[0].Code
		}
	}
	return "standard"
}

func (s *TopicDiscoveryRunService) prepareRun(ctx context.Context, userID uint, req request.TopicDiscoveryRunRequest) (topicRunInput, error) {
	keywords, description, direction, ok := normalizeTopicDiscoveryInput(req)
	if !ok {
		return topicRunInput{}, errorx.ErrParamsError
	}
	intensityCode := strings.TrimSpace(req.Intensity)
	if intensityCode == "" {
		intensityCode = s.defaultExecutionIntensityCode(ctx)
	}
	auditCode := strings.TrimSpace(req.AuditLevel)
	if auditCode == "" {
		auditCode = s.defaultAuditLevelCode(ctx)
	}
	intensity, err := s.catalog.FindExecutionIntensityByCode(ctx, intensityCode)
	if err != nil || intensity == nil {
		return topicRunInput{}, errorx.ErrParamsError
	}
	audit, err := s.catalog.FindAuditLevelByCode(ctx, entity.LLMSceneTopic, auditCode)
	if err != nil || audit == nil {
		return topicRunInput{}, errorx.ErrParamsError
	}

	venue := strings.TrimSpace(req.Venue)
	venueCodes := normalizeStringList(req.VenueCodes)
	if venue == "" && len(venueCodes) > 0 {
		venue = s.venueLabelsFromCodes(ctx, venueCodes)
	}
	if len(venueCodes) == 0 && venue == "" {
		return topicRunInput{}, errorx.ErrParamsError.WithDetail("请至少选择一个目标会议/期刊")
	}

	userLitIDs, err := userliterature.ParseUserLiteratureIDList(req.UserLiteratureIDs)
	if err != nil {
		return topicRunInput{}, errorx.ErrParamsError
	}
	sourceCodes, err := s.validateTopicLiteratureSourceSelection(ctx, req.SourceCodes, true)
	if err != nil {
		return topicRunInput{}, err
	}
	input := topicRunInput{
		DisciplineCode:    strings.TrimSpace(req.DisciplineCode),
		Keywords:          keywords,
		Description:       description,
		Direction:         direction,
		Venue:             venue,
		VenueCodes:        venueCodes,
		SourceCodes:       sourceCodes,
		Intensity:         intensityCode,
		AuditLevel:        auditCode,
		HumanCheckpoint:   !mustIdeaDiscoveryConfig().Platform.AutoProceed,
		RandomReferences:  req.RandomReferences,
		UserLiteratureIDs: userLitIDs,
		MaxPapers:         int(intensity.MaxPapers),
		MaxIdeas:          int(intensity.MaxIdeas),
		ExecRounds:        int(intensity.ExecRounds),
		InvokeMode:        strings.TrimSpace(req.InvokeMode),
	}
	if strings.TrimSpace(input.InvokeMode) == "" {
		input.InvokeMode = "phased"
	}
	return input, nil
}

func (s *TopicDiscoveryRunService) CommitManuscript(ctx context.Context, userID uint, title string) (*response.TopicDiscoveryRunView, error) {
	_ = title
	runVersion, rows, err := s.steps.GetLatestRunByUser(ctx, uint64(userID))
	if err != nil {
		return nil, err
	}
	if runVersion == 0 || len(rows) == 0 || currentRunCancelled(rows) {
		return nil, errorx.ErrTopicRunNotFound
	}
	for _, r := range rows {
		if strings.TrimSpace(r.Status) != "completed" {
			return nil, errorx.ErrTopicRunNotComplete
		}
	}
	id := manuscriptIDOf(rows)
	if id == 0 {
		return nil, errorx.ErrParamsError.WithDetail("run 未绑定论文，请在开始选题前创建论文")
	}
	if err := s.manuscripts.SetCurrentForUser(ctx, userID, uint(id)); err != nil {
		return nil, err
	}
	return s.buildRunView(id, runVersion, rows), nil
}

// LoadRunInputForUser 供 bot 等读取选题 run 表单快照（idea_discovery.input_params）。
func (s *TopicDiscoveryRunService) LoadRunInputForUser(ctx context.Context, userID uint, runVersion int) (topicRunInput, error) {
	return s.loadRunInput(ctx, userID, runVersion)
}

func (s *TopicDiscoveryRunService) loadRunInput(ctx context.Context, userID uint, runVersion int) (topicRunInput, error) {
	step, err := s.runInputStep(ctx, userID, runVersion)
	if err != nil || step == nil || len(step.InputParams) == 0 {
		return topicRunInput{}, errorx.ErrTopicRunNotFound
	}
	var input topicRunInput
	if err := json.Unmarshal(step.InputParams, &input); err != nil {
		return topicRunInput{}, errorx.ErrTopicRunNotFound
	}
	if input.ExecRounds < 1 && strings.TrimSpace(input.Intensity) != "" {
		if inten, err := s.catalog.FindExecutionIntensityByCode(ctx, strings.TrimSpace(input.Intensity)); err == nil && inten != nil {
			input.ExecRounds = int(inten.ExecRounds)
		}
	}
	if input.ExecRounds < 1 {
		input.ExecRounds = 1
	}
	return input, nil
}

func (s *TopicDiscoveryRunService) nextPendingIndex(ctx context.Context, userID uint, runVersion int) int {
	stages := s.stagesForRun(ctx, userID, runVersion)
	for i, stage := range stages {
		if stage == "topic_confirm" {
			continue
		}
		step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, stage)
		if err != nil || step == nil {
			continue
		}
		if step.Status == "pending" || step.Status == "failed" {
			return i
		}
	}
	return -1
}

func (s *TopicDiscoveryRunService) executeFromStage(ctx context.Context, userID uint, runVersion int, input topicRunInput, fromIndex int) error {
	stages := s.stagesForRun(ctx, userID, runVersion)
	for i := fromIndex; i < len(stages); i++ {
		stage := stages[i]
		if err := s.executeStage(ctx, userID, runVersion, input, stage); err != nil {
			return err
		}
		if st, _ := s.runningStageCode(ctx, userID, runVersion); st != "" {
			return nil
		}
	}
	return nil
}

func (s *TopicDiscoveryRunService) runningStageCode(ctx context.Context, userID uint, runVersion int) (string, error) {
	stages := s.stagesForRun(ctx, userID, runVersion)
	for _, stage := range stages {
		step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, stage)
		if err != nil || step == nil {
			continue
		}
		if strings.TrimSpace(step.Status) == "running" {
			return stage, nil
		}
	}
	return "", nil
}

func (s *TopicDiscoveryRunService) executeStage(ctx context.Context, userID uint, runVersion int, input topicRunInput, stageCode string) error {
	step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, stageCode)
	if err != nil || step == nil {
		return errorx.ErrTopicStepInvalid
	}
	if step.Status == "completed" {
		return nil
	}
	// 第三步仅由 POST confirm-round 完成，流水线不得将其标为 running/completed。
	if stageCode == "topic_confirm" {
		return nil
	}
	stageStarted := time.Now()
	step.Status = "running"
	step.StartedAt = &stageStarted
	if err := s.steps.SaveStep(ctx, step); err != nil {
		return err
	}

	var execErr error
	switch stageCode {
	case "idea_discovery":
		execErr = s.stageBeginTopicDiscoveryAsync(ctx, userID, runVersion, step)
	case "research_lit", "idea_creator", "novelty_check", "research_review", "research_refine":
		execErr = s.stageBeginPhasedTopicStageAsync(ctx, userID, runVersion, step)
	default:
		execErr = errorx.ErrTopicStepInvalid
	}

	if execErr != nil {
		if step.StartedAt == nil {
			step.StartedAt = &stageStarted
		}
		paperstep.FailTopicStep(step, execErr.Error())
		if saveErr := s.steps.SaveStep(ctx, step); saveErr != nil {
			return saveErr
		}
		return execErr
	}
	if (stageCode == "idea_discovery" || isPhasedTopicStage(stageCode)) && strings.TrimSpace(step.Status) == "running" {
		if step.StartedAt == nil {
			step.StartedAt = &stageStarted
		}
		step.CompletedAt = nil
		return s.steps.SaveStep(ctx, step)
	}
	done := time.Now()
	if step.StartedAt == nil {
		step.StartedAt = &stageStarted
	}
	step.CompletedAt = &done
	step.Status = "completed"
	return s.steps.SaveStep(ctx, step)
}

func (s *TopicDiscoveryRunService) orderSourceCodesBySort(ctx context.Context, selected []string) []string {
	active, err := s.litSource.ListActiveOrdered(ctx)
	if err != nil || len(active) == 0 {
		return selected
	}
	sel := map[string]struct{}{}
	for _, c := range selected {
		c = strings.TrimSpace(c)
		if c != "" {
			sel[c] = struct{}{}
		}
	}
	var ordered []string
	for _, row := range active {
		if _, ok := sel[row.Code]; ok {
			ordered = append(ordered, row.Code)
		}
	}
	seen := map[string]struct{}{}
	for _, c := range ordered {
		seen[c] = struct{}{}
	}
	for _, c := range selected {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if _, ok := seen[c]; !ok {
			ordered = append(ordered, c)
		}
	}
	return ordered
}

func (s *TopicDiscoveryRunService) stageRetrieveExecuteSearch(ctx context.Context, step *entity.PaperOutputTopicStep, input topicRunInput) error {
	query := literatureSearchQueryFromInput(input)
	scope := literature.RetrieveLogScope(step)
	logger.InfoZ(ctx, "topic-retrieve-start",
		zap.Uint64("step_id", scope.StepID),
		zap.Uint64("manuscript_id", scope.ManuscriptID),
		zap.Int("run_version", scope.RunVersion),
		zap.String("search_query", query),
		zap.Bool("random_references", input.RandomReferences),
		zap.Strings("source_codes", input.SourceCodes),
		zap.Int("max_papers", input.MaxPapers))

	resultMeta := paperstep.LoadStepResultMap(step)
	var hits []storedLiteratureHit
	if input.RandomReferences {
		hits, resultMeta = s.retrieveLiteratureRandom(ctx, query, input, resultMeta, scope)
	} else {
		hits, resultMeta = s.retrieveLiteratureSequential(ctx, query, input, resultMeta, scope)
	}
	logger.InfoZ(ctx, "topic-retrieve-finished",
		zap.Uint64("step_id", scope.StepID),
		zap.Int("run_version", scope.RunVersion),
		zap.Int("hit_count", len(hits)))

	links := make([]storedLiteratureLink, 0, len(hits))
	downloads := make([]literature.LiteratureDownloadItem, 0, len(hits))
	for _, h := range hits {
		if strings.TrimSpace(h.URL) == "" {
			continue
		}
		links = append(links, storedLiteratureLink{
			Title:       h.Title,
			URL:         h.URL,
			ExternalKey: h.ExternalKey,
			SourceCode:  h.SourceCode,
		})
		pdfURL := literature.ResolveLiteraturePDFURLWithPMC(h.SourceCode, h.ExternalKey, h.URL, h.DOI, literature.PmcidFromHitMeta(h.Meta))
		localRel := literature.LiteratureLocalRelPath(h.ExternalKey)
		downloads = append(downloads, literature.LiteratureDownloadItem{
			ExternalKey: h.ExternalKey,
			PdfURL:      pdfURL,
			LocalPath:   localRel,
		})
	}

	resultMeta["literature_hits"] = hits
	brief := fallbackLiteratureBrief(hits)
	mergeLiteratureBriefIntoResult(resultMeta, brief)
	resultMeta["hit_count"] = len(hits)
	resultMeta["verified_count"] = len(hits)
	resultMeta["search_query"] = query
	resultMeta["literature_links"] = links
	resultMeta["literature_downloads"] = downloads
	if b, e := json.Marshal(brief); e == nil {
		resultMeta["literature_brief"] = json.RawMessage(b)
	}
	step.Result = jsonx.MustJSON(resultMeta)
	paperstep.SyncStepFilesAppend(step, downloads)
	if len(hits) == 0 {
		return fmt.Errorf("未检索到文献")
	}
	if literature.LiteratureAllDownloadsLackPDFURL(downloads) {
		return fmt.Errorf("无可用 PDF 下载链接")
	}
	plannedPDF := literature.CountPlannedPDFDownloads(downloads)
	onDisk := literature.CountLiteraturePDFsOnDisk(downloads)
	if literature.LiteratureRetrieveCorpusReady(downloads) {
		step.Status = "completed"
		step.ExecProgressPct = 100
	} else {
		step.Status = "running"
		step.ExecProgressPct = literature.ExecProgressPct(onDisk, plannedPDF, onDisk == 0 && plannedPDF > 0)
	}
	return nil
}

func (s *TopicDiscoveryRunService) stageContext(ctx context.Context, userID uint, runVersion int) (string, error) {
	var parts []string
	for _, stage := range topicDiscoveryStages {
		if stage == "topic_discovery" {
			continue
		}
		st, err := s.steps.GetStep(ctx, uint64(userID), runVersion, stage)
		if err != nil || st == nil || len(st.Result) == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("[%s]\n%s", stage, string(st.Result)))
	}
	return strings.Join(parts, "\n\n"), nil
}

func (s *TopicDiscoveryRunService) contentLocaleForStep(ctx context.Context, step *entity.PaperOutputTopicStep) string {
	if s == nil || step == nil || step.ManuscriptID == 0 || s.manuscripts == nil {
		return "en"
	}
	ms, err := s.manuscripts.GetByIDForUser(ctx, uint(step.ManuscriptID), uint(step.UserID))
	if err != nil || ms == nil {
		return "en"
	}
	return lang.NormalizeContentLanguage(ms.Language)
}

func llmUserMessage(userPrefix, userExtra string) string {
	user := strings.TrimSpace(userPrefix)
	extra := strings.TrimSpace(userExtra)
	if user == "" {
		return extra
	}
	if extra == "" {
		return user
	}
	return user + "\n\n" + extra
}

func (s *TopicDiscoveryRunService) resolveStrategyRoute(ctx context.Context, behaviorKey string) (*entity.PaperStrategyRoute, error) {
	if s == nil || s.strategyRepo == nil {
		return nil, nil
	}
	return s.strategyRepo.FindActiveRouteByBehaviorKey(ctx, behaviorKey)
}

func (s *TopicDiscoveryRunService) reloadView(ctx context.Context, userID uint, runVersion int) (*response.TopicDiscoveryRunView, error) {
	rows, err := s.steps.ListByUserRun(ctx, uint64(userID), runVersion)
	if err != nil {
		return nil, err
	}
	return s.buildRunView(manuscriptIDOf(rows), runVersion, rows), nil
}

func (s *TopicDiscoveryRunService) reloadViewOrErr(ctx context.Context, userID uint, runVersion int, stageErr error) (*response.TopicDiscoveryRunView, error) {
	view, err := s.reloadView(ctx, userID, runVersion)
	if err != nil {
		return nil, stageErr
	}
	return view, nil
}

func (s *TopicDiscoveryRunService) buildRunView(manuscriptID uint64, runVersion int, rows []entity.PaperOutputTopicStep) *response.TopicDiscoveryRunView {
	rows = SortTopicSteps(rows)
	phasedRun := isPhasedPerStepRun(rows)
	view := &response.TopicDiscoveryRunView{
		ManuscriptID: manuscriptID,
		RunVersion:   runVersion,
		Steps:        make([]response.TopicDiscoveryStepView, 0, len(rows)),
	}
	view.HumanCheckpoint = !mustIdeaDiscoveryConfig().Platform.AutoProceed
	for _, g := range BuildIdeaDiscoveryUIGroups(mustIdeaDiscoveryConfig()) {
		view.IdeaDiscoveryUIGroups = append(view.IdeaDiscoveryUIGroups, response.IdeaDiscoveryUIGroupView{
			Index:    g.Index,
			PhaseIDs: append([]string(nil), g.PhaseIDs...),
			LabelZh:  g.LabelZh,
		})
	}

	runStatus := "completed"
	for _, r := range rows {
		st := strings.TrimSpace(r.Status)
		sv := response.TopicDiscoveryStepView{
			StageCode:       r.StageCode,
			Sort:            r.Sort,
			Status:          st,
			ExecProgressPct: r.ExecProgressPct,
		}
		if len(r.InputParams) > 0 {
			sv.InputParams = json.RawMessage(r.InputParams)
		}
		if r.ErrorMsg != nil {
			sv.ErrorMsg = strings.TrimSpace(*r.ErrorMsg)
		}
		if r.ExecCommand != nil {
			raw := strings.TrimSpace(*r.ExecCommand)
			if raw != "" {
				if json.Valid([]byte(raw)) {
					sv.ExecCommand = json.RawMessage(raw)
				} else {
					sv.ExecCommand = json.RawMessage(jsonx.MustJSON(map[string]any{"_legacy_plaintext": raw}))
				}
			}
		}
		if phasedRun && isPhasedTopicStage(r.StageCode) {
			if pub := paperstep.StepResultForAPI(&r); len(pub) > 0 {
				sv.Result = json.RawMessage(jsonx.MustJSON(pub))
			}
		} else if r.StageCode == "idea_discovery" || r.StageCode == "topic_discovery" {
			merged := paperstep.LoadStepResultMap(&r)
			var rt map[string]any
			if len(r.RuntimeState) > 0 {
				_ = json.Unmarshal(r.RuntimeState, &rt)
			}
			for k, v := range rt {
				merged[k] = v
			}
			if len(merged) > 0 {
				sv.Result = json.RawMessage(jsonx.MustJSON(merged))
			}
		} else {
			merged := paperstep.LoadStepExtra(&r)
			if len(merged) > 0 {
				sv.Result = json.RawMessage(jsonx.MustJSON(merged))
			} else if len(r.Result) > 0 {
				sv.Result = json.RawMessage(r.Result)
			}
		}
		if len(r.Files) > 0 {
			sv.Files = json.RawMessage(r.Files)
		}
		if r.StartedAt != nil {
			t := r.StartedAt.Format(time.RFC3339)
			sv.StartedAt = &t
		}
		if r.CompletedAt != nil {
			t := r.CompletedAt.Format(time.RFC3339)
			sv.CompletedAt = &t
		}
		view.Steps = append(view.Steps, sv)

		switch st {
		case "failed":
			runStatus = "failed"
		case "running":
			runStatus = "running"
		case "pending":
			if runStatus == "completed" {
				runStatus = "running"
			}
		}
	}
	if runStatus != "failed" {
		if phasedRun {
			if pause := PhasedPauseAfterStage(rows, view.HumanCheckpoint); pause != "" {
				view.PauseAfterStage = pause
				runStatus = "checkpoint"
			}
		} else {
			for _, r := range rows {
				if r.StageCode == "idea_discovery" && strings.TrimSpace(r.Status) == "running" {
					meta := paperstep.LoadStepExtra(&r)
					if IdeaDiscoveryRunAwaitingUserConfirm(meta) {
						view.PauseAfterStage = "idea_discovery"
						runStatus = "checkpoint"
						break
					}
				}
			}
		}
	}
	view.RunStatus = runStatus
	return view
}

func parseJSONOrWrap(text, key string) (map[string]any, string) {
	text = strings.TrimSpace(text)
	if idx := strings.Index(text, "{"); idx >= 0 {
		if end := strings.LastIndex(text, "}"); end > idx {
			text = text[idx : end+1]
		}
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(text), &m); err == nil {
		if sum, ok := m["summary"].(string); ok {
			return m, sum
		}
		return m, text
	}
	return map[string]any{key: text}, text
}

func appendUsageMeta(step *entity.PaperOutputTopicStep, usage llm.LLMUsage) {
	meta := paperstep.LoadStepExtra(step)
	meta["tokens_prompt"] = usage.PromptTokens
	meta["tokens_completion"] = usage.CompletionTokens
	step.Result = jsonx.MustJSON(meta)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func normalizeStringList(raw []string) []string {
	out := make([]string, 0, len(raw))
	seen := map[string]struct{}{}
	for _, s := range raw {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func (s *TopicDiscoveryRunService) venueLabelsFromCodes(ctx context.Context, codes []string) string {
	labels := make([]string, 0, len(codes))
	for _, code := range codes {
		v, err := s.catalog.FindVenueByCode(ctx, code)
		if err != nil || v == nil {
			labels = append(labels, code)
			continue
		}
		labels = append(labels, v.Name)
	}
	return strings.Join(labels, " / ")
}
