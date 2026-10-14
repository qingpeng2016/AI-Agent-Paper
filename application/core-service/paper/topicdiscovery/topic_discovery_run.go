package topicdiscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

var topicDiscoveryStages = []string{"retrieve", "topic_discovery", "topic_confirm"}

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
		runVersion, err = s.steps.BeginRun(ctx, uint64(userID), req.ManuscriptID, paramsBytes)
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
			runVersion, err = s.steps.BeginRun(ctx, uint64(userID), req.ManuscriptID, paramsBytes)
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
			if err := s.executeStage(ctx, userID, runVersion, input, topicDiscoveryStages[idx]); err != nil {
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
	input := topicRunInput{
		DisciplineCode:    strings.TrimSpace(req.DisciplineCode),
		Keywords:          keywords,
		Description:       description,
		Direction:         direction,
		Venue:             venue,
		VenueCodes:        venueCodes,
		SourceCodes:       req.SourceCodes,
		Intensity:         intensityCode,
		AuditLevel:        auditCode,
		HumanCheckpoint:   req.HumanCheckpoint,
		RandomReferences:  req.RandomReferences,
		UserLiteratureIDs: userLitIDs,
		MaxPapers:         int(intensity.MaxPapers),
		MaxIdeas:          int(intensity.MaxIdeas),
		ExecRounds:        int(intensity.ExecRounds),
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

// LoadRunInputForUser 供 bot 等读取 retrieve 步表单快照。
func (s *TopicDiscoveryRunService) LoadRunInputForUser(ctx context.Context, userID uint, runVersion int) (topicRunInput, error) {
	return s.loadRunInput(ctx, userID, runVersion)
}

func (s *TopicDiscoveryRunService) loadRunInput(ctx context.Context, userID uint, runVersion int) (topicRunInput, error) {
	step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "retrieve")
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
	for i, stage := range topicDiscoveryStages {
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
	for i := fromIndex; i < len(topicDiscoveryStages); i++ {
		if err := s.executeStage(ctx, userID, runVersion, input, topicDiscoveryStages[i]); err != nil {
			return err
		}
		if st, _ := s.runningStageCode(ctx, userID, runVersion); st != "" {
			return nil
		}
		if input.HumanCheckpoint {
			return nil
		}
	}
	return nil
}

func (s *TopicDiscoveryRunService) runningStageCode(ctx context.Context, userID uint, runVersion int) (string, error) {
	for _, stage := range topicDiscoveryStages {
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
	// retrieve 须在 stageRetrieve 写入 result/extra 后再落库 running，否则 PDF bot 会扫到「running 且无 result」误杀。
	if stageCode != "retrieve" {
		step.Status = "running"
		step.StartedAt = &stageStarted
		if err := s.steps.SaveStep(ctx, step); err != nil {
			return err
		}
	}

	var execErr error
	switch stageCode {
	case "retrieve":
		execErr = s.stageRetrieve(ctx, step, input)
	case "topic_discovery":
		execErr = s.stageBeginTopicDiscoveryAsync(ctx, userID, runVersion, step)
	default:
		execErr = errorx.ErrTopicStepInvalid
	}

	if execErr != nil {
		done := time.Now()
		if step.StartedAt == nil {
			step.StartedAt = &stageStarted
		}
		step.CompletedAt = &done
		step.Status = "failed"
		stepMeta := paperstep.LoadStepExtra(step)
		stepMeta["error"] = execErr.Error()
		step.Extra = jsonx.MustJSON(stepMeta)
		if saveErr := s.steps.SaveStep(ctx, step); saveErr != nil {
			return saveErr
		}
		return execErr
	}
	if (stageCode == "retrieve" || stageCode == "topic_discovery") && strings.TrimSpace(step.Status) == "running" {
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

func (s *TopicDiscoveryRunService) stageRetrieve(ctx context.Context, step *entity.PaperOutputTopicStep, input topicRunInput) error {
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

	extraMeta := paperstep.LoadStepExtra(step)
	var hits []storedLiteratureHit
	if input.RandomReferences {
		hits, extraMeta = s.retrieveLiteratureRandom(ctx, query, input, extraMeta, scope)
	} else {
		hits, extraMeta = s.retrieveLiteratureSequential(ctx, query, input, extraMeta, scope)
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
		downloads = append(downloads, literature.LiteratureDownloadItem{
			ExternalKey: h.ExternalKey,
			PdfURL:      pdfURL,
			LocalPath:   relIfExists(pdfURL, h.ExternalKey),
		})
	}

	result := map[string]any{"literature_hits": hits}
	brief := fallbackLiteratureBrief(hits)
	mergeLiteratureBriefIntoResult(result, brief)
	step.Result = jsonx.MustJSON(result)
	extraPatch := map[string]any{
		"hit_count":            len(hits),
		"verified_count":       len(hits),
		"search_query":         query,
		"literature_links":     links,
		"literature_downloads": downloads,
	}
	if b, e := json.Marshal(brief); e == nil {
		extraPatch["literature_brief"] = json.RawMessage(b)
	}
	step.Extra = jsonx.MustJSON(paperstep.MergeStepExtra(extraMeta, extraPatch))
	paperstep.SyncStepFilesAppend(step, downloads)
	if len(hits) == 0 {
		step.SummaryText = jsonx.PtrString("检索未命中文献，未落盘 PDF")
		return fmt.Errorf("未检索到文献")
	}
	if literature.LiteratureAllDownloadsLackPDFURL(downloads) {
		step.SummaryText = jsonx.PtrString(fmt.Sprintf("检索到 %d 篇，但来源无 PDF 下载链接，未落盘", len(hits)))
		return fmt.Errorf("无可用 PDF 下载链接")
	}
	step.SummaryText = jsonx.PtrString(fmt.Sprintf("检索完成，共 %d 篇文献，PDF 下载中…", len(hits)))
	if literature.LiteratureRetrieveCorpusReady(downloads) {
		n := literature.CountLiteraturePDFsOnDisk(downloads)
		step.SummaryText = jsonx.PtrString(literature.FormatRetrievePDFSummary(len(hits), n))
		step.Status = "completed"
	} else {
		step.Status = "running"
	}
	return nil
}

func fileExistsNonEmpty(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.Size() > 0
}

func relIfExists(pdfURL, externalKey string) string {
	if strings.TrimSpace(pdfURL) == "" {
		return ""
	}
	rel := literature.LiteratureLocalRelPath(externalKey)
	if fileExistsNonEmpty(literature.LiteratureLocalAbsPath(rel)) {
		return rel
	}
	return ""
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
		if st.SummaryText != nil {
			parts = append(parts, *st.SummaryText)
		}
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
	return lang.NormalizeContentLanguage(ms.ContentLanguage)
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
	view := &response.TopicDiscoveryRunView{
		ManuscriptID: manuscriptID,
		RunVersion:   runVersion,
		Steps:        make([]response.TopicDiscoveryStepView, 0, len(rows)),
	}
	var human bool
	for _, r := range rows {
		if len(r.InputParams) > 0 {
			var in topicRunInput
			if json.Unmarshal(r.InputParams, &in) == nil {
				human = in.HumanCheckpoint
			}
		}
	}
	view.HumanCheckpoint = human

	runStatus := "completed"
	for _, r := range rows {
		st := strings.TrimSpace(r.Status)
		sv := response.TopicDiscoveryStepView{
			StageCode:       r.StageCode,
			Status:          st,
			ExecProgressPct: r.ExecProgressPct,
		}
		if len(r.InputParams) > 0 {
			sv.InputParams = json.RawMessage(r.InputParams)
		}
		if r.SummaryText != nil {
			sv.SummaryText = *r.SummaryText
		}
		if len(r.Result) > 0 {
			sv.Result = json.RawMessage(r.Result)
		}
		if len(r.Extra) > 0 {
			sv.Extra = json.RawMessage(r.Extra)
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
	if runStatus != "failed" && human {
		for i, r := range rows {
			cur := strings.TrimSpace(r.Status)
			next := ""
			if i+1 < len(rows) {
				next = strings.TrimSpace(rows[i+1].Status)
			}
			if cur == "completed" && next == "pending" {
				view.PauseAfterStage = r.StageCode
				runStatus = "checkpoint"
				break
			}
		}
	}
	if runStatus != "failed" {
		discDone, confirmPending := false, false
		for _, r := range rows {
			switch r.StageCode {
			case "topic_discovery":
				discDone = strings.TrimSpace(r.Status) == "completed"
			case "topic_confirm":
				confirmPending = strings.TrimSpace(r.Status) == "pending"
			}
		}
		if discDone && confirmPending {
			view.PauseAfterStage = "topic_confirm"
			runStatus = "checkpoint"
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
	step.Extra = jsonx.MustJSON(meta)
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
