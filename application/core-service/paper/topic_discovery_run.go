package paper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/request"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
	"gorm.io/datatypes"
)

var topicDiscoveryStages = []string{"retrieve", "generate_ideas", "novelty", "audit"}

type topicRunInput struct {
	DisciplineCode  string   `json:"discipline_code"`
	Keywords        []string `json:"keywords"`
	Description     string   `json:"description"`
	Direction       string   `json:"direction"`
	Venue           string   `json:"venue"`
	SourceCodes     []string `json:"source_codes"`
	Intensity       string   `json:"intensity"`
	AuditLevel      string   `json:"audit_level"`
	HumanCheckpoint bool     `json:"human_checkpoint"`
	MaxPapers       int      `json:"max_papers"`
	MaxIdeas        int      `json:"max_ideas"`
	AuditRounds     int      `json:"audit_rounds"`
}

type storedLiteratureHit struct {
	SourceID        uint     `json:"source_id"`
	SourceCode      string   `json:"source_code"`
	ExternalKey     string   `json:"external_key"`
	Title           string   `json:"title"`
	Authors         []string `json:"authors,omitempty"`
	URL             string   `json:"url,omitempty"`
	DOI             string   `json:"doi,omitempty"`
	PublishedYear   int      `json:"published_year,omitempty"`
	RelevanceScore  float64  `json:"relevance_score,omitempty"`
	QueryText       string   `json:"query_text,omitempty"`
	Meta            any      `json:"meta,omitempty"`
}

type storedLiteratureLink struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	ExternalKey string `json:"external_key,omitempty"`
	SourceCode  string `json:"source_code,omitempty"`
}

type TopicDiscoveryRunService struct {
	manuscripts repository.PaperManuscriptRepo
	steps       repository.PaperOutputTopicStepRepo
	catalog     repository.PaperRefCatalogRepo
	litSource   repository.PaperRefLiteratureSourceRepo
	litSearch   *LiteratureSearchService
	llmRepo     repository.PaperLLMRepo
	llm         *LLMChatService
}

func NewTopicDiscoveryRunService(
	manuscripts repository.PaperManuscriptRepo,
	steps repository.PaperOutputTopicStepRepo,
	catalog repository.PaperRefCatalogRepo,
	litSource repository.PaperRefLiteratureSourceRepo,
	litSearch *LiteratureSearchService,
	llmRepo repository.PaperLLMRepo,
	llm *LLMChatService,
) *TopicDiscoveryRunService {
	return &TopicDiscoveryRunService{
		manuscripts: manuscripts,
		steps:       steps,
		catalog:     catalog,
		litSource:   litSource,
		litSearch:   litSearch,
		llmRepo:     llmRepo,
		llm:         llm,
	}
}

func (s *TopicDiscoveryRunService) GetCurrentRun(ctx context.Context, userID uint) (*response.TopicDiscoveryRunView, error) {
	runVersion, stepRows, err := s.steps.GetLatestRunByUser(ctx, uint64(userID))
	if err != nil {
		return nil, err
	}
	if runVersion == 0 || len(stepRows) == 0 || currentRunCancelled(stepRows) {
		return nil, nil
	}
	return s.buildRunView(manuscriptIDOf(stepRows), runVersion, stepRows), nil
}

func (s *TopicDiscoveryRunService) CancelCurrentRun(ctx context.Context, userID uint) error {
	runVersion, stepRows, err := s.steps.GetLatestRunByUser(ctx, uint64(userID))
	if err != nil {
		return err
	}
	if runVersion == 0 || len(stepRows) == 0 || currentRunCancelled(stepRows) {
		return nil
	}
	return s.steps.CancelRun(ctx, uint64(userID), runVersion)
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
		input, err = s.prepareRun(ctx, userID, req)
		if err != nil {
			return nil, err
		}
		paramsBytes, _ := json.Marshal(input)
		runVersion, err = s.steps.BeginRun(ctx, uint64(userID), paramsBytes)
		if err != nil {
			return nil, err
		}
		if err := s.executeFromStage(ctx, userID, runVersion, input, 0); err != nil {
			return nil, err
		}
	case "continue":
		ver, rows, err := s.steps.GetLatestRunByUser(ctx, uint64(userID))
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
		if err := s.executeFromStage(ctx, userID, runVersion, input, idx); err != nil {
			return nil, err
		}
	case "run_all":
		ver, rows, err := s.steps.GetLatestRunByUser(ctx, uint64(userID))
		if err != nil {
			return nil, err
		}
		if ver == 0 || currentRunCancelled(rows) {
			input, err = s.prepareRun(ctx, userID, req)
			if err != nil {
				return nil, err
			}
			paramsBytes, _ := json.Marshal(input)
			runVersion, err = s.steps.BeginRun(ctx, uint64(userID), paramsBytes)
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
				return nil, err
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
	rows, err := s.catalog.ListAuditLevels(ctx)
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
	return "polished"
}

func (s *TopicDiscoveryRunService) prepareRun(ctx context.Context, userID uint, req request.TopicDiscoveryRunRequest) (topicRunInput, error) {
	keywords, description, direction, ok := normalizeTopicDiscoveryInput(req)
	if !ok {
		return topicRunInput{}, errorx.ErrParamsError
	}
	if len(req.SourceCodes) == 0 {
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
	audit, err := s.catalog.FindAuditLevelByCode(ctx, auditCode)
	if err != nil || audit == nil {
		return topicRunInput{}, errorx.ErrParamsError
	}

	input := topicRunInput{
		DisciplineCode:  strings.TrimSpace(req.DisciplineCode),
		Keywords:        keywords,
		Description:     description,
		Direction:       direction,
		Venue:           strings.TrimSpace(req.Venue),
		SourceCodes:     req.SourceCodes,
		Intensity:       intensityCode,
		AuditLevel:      auditCode,
		HumanCheckpoint: req.HumanCheckpoint,
		MaxPapers:       int(intensity.MaxPapers),
		MaxIdeas:        int(intensity.MaxIdeas),
		AuditRounds:     int(audit.AuditRounds),
	}
	return input, nil
}

func (s *TopicDiscoveryRunService) CommitManuscript(ctx context.Context, userID uint, title string) (*response.TopicDiscoveryRunView, error) {
	runVersion, rows, err := s.steps.GetLatestRunByUser(ctx, uint64(userID))
	if err != nil {
		return nil, err
	}
	if runVersion == 0 || len(rows) == 0 || currentRunCancelled(rows) {
		return nil, errorx.ErrTopicRunNotFound
	}
	for _, r := range rows {
		if r.Status != "completed" {
			return nil, errorx.ErrTopicRunNotComplete
		}
	}
	if id := manuscriptIDOf(rows); id > 0 {
		return s.buildRunView(id, runVersion, rows), nil
	}
	t := strings.TrimSpace(title)
	if t == "" {
		if in, e := s.loadRunInput(ctx, userID, runVersion); e == nil {
			t = strings.TrimSpace(in.Description)
			if t == "" {
				t = strings.TrimSpace(in.Direction)
			}
		}
	}
	if t == "" {
		t = "未命名论文"
	}
	if len([]rune(t)) > 120 {
		t = string([]rune(t)[:120])
	}
	ms, err := s.manuscripts.Create(ctx, userID, t)
	if err != nil {
		return nil, err
	}
	if err := s.steps.BindManuscript(ctx, uint64(userID), runVersion, uint64(ms.ID)); err != nil {
		return nil, err
	}
	return s.reloadView(ctx, userID, runVersion)
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
	return input, nil
}

func (s *TopicDiscoveryRunService) nextPendingIndex(ctx context.Context, userID uint, runVersion int) int {
	for i, stage := range topicDiscoveryStages {
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
		if input.HumanCheckpoint {
			return nil
		}
	}
	return nil
}

func (s *TopicDiscoveryRunService) executeStage(ctx context.Context, userID uint, runVersion int, input topicRunInput, stageCode string) error {
	step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, stageCode)
	if err != nil || step == nil {
		return errorx.ErrTopicStepInvalid
	}
	if step.Status == "completed" {
		return nil
	}
	now := time.Now()
	step.Status = "running"
	step.StartedAt = &now
	if err := s.steps.SaveStep(ctx, step); err != nil {
		return err
	}

	var execErr error
	switch stageCode {
	case "retrieve":
		execErr = s.stageRetrieve(ctx, step, input)
	case "generate_ideas":
		execErr = s.stageGenerateIdeas(ctx, userID, runVersion, step, input)
	case "novelty":
		execErr = s.stageNovelty(ctx, userID, runVersion, step, input)
	case "audit":
		execErr = s.stageAudit(ctx, userID, runVersion, step, input)
	default:
		execErr = errorx.ErrTopicStepInvalid
	}

	done := time.Now()
	step.CompletedAt = &done
	if execErr != nil {
		step.Status = "failed"
		stepMeta := map[string]any{}
		if len(step.Extra) > 0 {
			_ = json.Unmarshal(step.Extra, &stepMeta)
		}
		stepMeta["error"] = execErr.Error()
		step.Extra = mustJSON(stepMeta)
		_ = s.steps.SaveStep(ctx, step)
		return execErr
	}
	step.Status = "completed"
	return s.steps.SaveStep(ctx, step)
}

func (s *TopicDiscoveryRunService) stageRetrieve(ctx context.Context, step *entity.PaperOutputTopicStep, input topicRunInput) error {
	query := literatureSearchQueryFromInput(input)
	perSource := input.MaxPapers / len(input.SourceCodes)
	if perSource < 5 {
		perSource = 5
	}
	if perSource > 50 {
		perSource = 50
	}

	seen := map[string]struct{}{}
	hits := make([]storedLiteratureHit, 0, input.MaxPapers)
	for _, code := range input.SourceCodes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		srcRow, _ := s.litSource.FindActiveByCode(ctx, code)
		var searchResult *response.PaperLiteratureSearchResult
		var searchErr error
		switch code {
		case sourceArxiv:
			searchResult, searchErr = s.litSearch.SearchArxiv(ctx, query, perSource)
		case sourceOpenAlex:
			searchResult, searchErr = s.litSearch.SearchOpenAlex(ctx, query, perSource)
		case sourceSemanticScholar:
			searchResult, searchErr = s.litSearch.SearchSemanticScholar(ctx, query, perSource)
		default:
			continue
		}
		if searchErr != nil || searchResult == nil {
			continue
		}
		var sourceID uint
		if srcRow != nil {
			sourceID = srcRow.ID
		}
		for _, h := range searchResult.Hits {
			if len(hits) >= input.MaxPapers {
				break
			}
			key := h.ExternalKey
			if key == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			linkURL := resolveLiteratureHitURL(code, key, h.URL, h.DOI)
			hits = append(hits, storedLiteratureHit{
				SourceID:      sourceID,
				SourceCode:    code,
				ExternalKey:   key,
				Title:         h.Title,
				Authors:       h.Authors,
				URL:           linkURL,
				DOI:           h.DOI,
				PublishedYear: h.PublishedYear,
				QueryText:     query,
				Meta: map[string]any{
					"abstract":       h.Abstract,
					"doi":            h.DOI,
					"url":            linkURL,
					"published_year": h.PublishedYear,
				},
			})
		}
	}

	links := make([]storedLiteratureLink, 0, len(hits))
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
	}

	result := map[string]any{"literature_hits": hits}
	brief := fallbackLiteratureBrief(hits)
	mergeLiteratureBriefIntoResult(result, brief)
	step.Result = mustJSON(result)
	extra := map[string]any{
		"hit_count":        len(hits),
		"verified_count":   len(hits),
		"search_query":     query,
		"literature_links": links,
	}
	if b, e := json.Marshal(brief); e == nil {
		extra["literature_brief"] = json.RawMessage(b)
	}
	step.Extra = mustJSON(extra)
	step.SummaryText = ptrString(fmt.Sprintf("检索完成，共 %d 篇文献", len(hits)))
	return nil
}

func (s *TopicDiscoveryRunService) stageGenerateIdeas(ctx context.Context, userID uint, runVersion int, step *entity.PaperOutputTopicStep, input topicRunInput) error {
	ctxBlock, _ := s.stageContext(ctx, userID, runVersion)
	userMsg := fmt.Sprintf(`Research direction: %s
Target venue: %s
Max ideas: %d

Corpus:
%s

Return ONLY valid JSON: {"ideas":[{"title":"","problem":"","approach":"","contribution":""}]}
Use Chinese for text fields.`, input.Direction, input.Venue, input.MaxIdeas, ctxBlock)

	text, usage, err := s.callStageLLM(ctx, step, "generate_ideas", map[string]string{
		"direction": input.Direction,
		"max_ideas": fmt.Sprintf("%d", input.MaxIdeas),
	}, userMsg)
	if err != nil {
		return err
	}
	result, summary := parseJSONOrWrap(text, "ideas")
	step.Result = mustJSON(result)
	step.SummaryText = ptrString(summary)
	appendUsageMeta(step, usage)
	return nil
}

func (s *TopicDiscoveryRunService) stageNovelty(ctx context.Context, userID uint, runVersion int, step *entity.PaperOutputTopicStep, input topicRunInput) error {
	ctxBlock, _ := s.stageContext(ctx, userID, runVersion)
	userMsg := fmt.Sprintf(`Perform novelty analysis for direction: %s

Prior steps:
%s

Return ONLY valid JSON: {"lines":["..."],"risks":[{"idea_title":"","risk":"low|medium|high","note":""}]}
Use Chinese.`, input.Direction, ctxBlock)

	text, usage, err := s.callStageLLM(ctx, step, "novelty", map[string]string{"direction": input.Direction}, userMsg)
	if err != nil {
		return err
	}
	result, summary := parseJSONOrWrap(text, "novelty")
	step.Result = mustJSON(result)
	step.SummaryText = ptrString(summary)
	appendUsageMeta(step, usage)
	return nil
}

func (s *TopicDiscoveryRunService) stageAudit(ctx context.Context, userID uint, runVersion int, step *entity.PaperOutputTopicStep, input topicRunInput) error {
	rounds := input.AuditRounds
	if rounds < 1 {
		rounds = 1
	}
	var roundOutputs []map[string]any
	var lastText string
	var totalUsage LLMUsage
	for r := 1; r <= rounds; r++ {
		userMsg := s.buildAuditUserExtra(ctx, userID, runVersion, input, r, rounds)
		text, usage, err := s.callStageLLM(ctx, step, "audit", map[string]string{
			"direction": compactDirectionForPrompt(input),
			"keywords":  keywordsLine(input),
			"venue":     input.Venue,
		}, userMsg)
		if err != nil {
			return err
		}
		lastText = text
		totalUsage.PromptTokens += usage.PromptTokens
		totalUsage.CompletionTokens += usage.CompletionTokens
		parsed, _ := parseJSONOrWrap(text, "audit")
		roundOutputs = append(roundOutputs, parsed)
	}
	result := map[string]any{"rounds": roundOutputs}
	step.Result = mustJSON(result)
	if summary, ok := roundOutputs[len(roundOutputs)-1]["summary"].(string); ok && summary != "" {
		step.SummaryText = ptrString(summary)
	} else {
		step.SummaryText = ptrString(lastText)
	}
	appendUsageMeta(step, totalUsage)
	return nil
}

func (s *TopicDiscoveryRunService) stageContext(ctx context.Context, userID uint, runVersion int) (string, error) {
	var parts []string
	for _, stage := range topicDiscoveryStages {
		if stage == "audit" {
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

type stageLLMResolve struct {
	model          *entity.PaperLLMModelConfig
	binding        *entity.PaperLLMWorkflowBinding
	defPrompt      *entity.PaperLLMPromptTemplate
	stagePrompt    *entity.PaperLLMPromptTemplate
	system         string
	userPrefix     string
}

func (s *TopicDiscoveryRunService) callStageLLM(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	stageCode string,
	vars map[string]string,
	userExtra string,
) (string, LLMUsage, error) {
	resolved, err := s.resolveStageLLM(ctx, stageCode, vars)
	if err != nil {
		persistStepLLMInput(step, stageCode, "", llmUserMessage("", userExtra), "")
		return "", LLMUsage{}, err
	}
	user := llmUserMessage(resolved.userPrefix, userExtra)
	modelName := resolved.model.ModelName
	persistStepLLMInput(step, stageCode, resolved.system, user, modelName)
	start := time.Now()
	text, usage, err := s.llm.Complete(ctx, resolved.model, resolved.system, user)
	s.persistLLMCallLog(ctx, step.ManuscriptID, stageCode, resolved, usage, int(time.Since(start).Milliseconds()), err)
	if err != nil {
		persistStepLLMInput(step, stageCode, resolved.system, user, modelName)
	}
	return text, usage, err
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

// persistStepLLMInput 将发给模型的 system/user 写入 step.input_params（主存储），并在 extra 保留 llm_* 副本便于接口展示。
func persistStepLLMInput(step *entity.PaperOutputTopicStep, stageCode, system, user, modelName string) {
	if step == nil {
		return
	}
	llmInput := map[string]any{
		"stage_code":    stageCode,
		"system_prompt": system,
		"user_prompt":   user,
	}
	if strings.TrimSpace(modelName) != "" {
		llmInput["model_name"] = modelName
	}

	if stageCode == "retrieve" && len(step.InputParams) > 0 {
		existing := map[string]any{}
		_ = json.Unmarshal(step.InputParams, &existing)
		if existing == nil {
			existing = map[string]any{}
		}
		existing["llm_request"] = llmInput
		step.InputParams = mustJSON(existing)
	} else {
		step.InputParams = mustJSON(llmInput)
	}

	meta := map[string]any{}
	if len(step.Extra) > 0 {
		_ = json.Unmarshal(step.Extra, &meta)
	}
	meta["llm_stage_code"] = stageCode
	if strings.TrimSpace(system) != "" {
		meta["llm_system"] = system
	}
	if strings.TrimSpace(user) != "" {
		meta["llm_user"] = user
	}
	if strings.TrimSpace(modelName) != "" {
		meta["llm_model_name"] = modelName
	}
	step.Extra = mustJSON(meta)
}

func (s *TopicDiscoveryRunService) persistLLMCallLog(
	ctx context.Context,
	manuscriptID uint64,
	stageCode string,
	resolved *stageLLMResolve,
	usage LLMUsage,
	latencyMs int,
	callErr error,
) {
	if resolved == nil || resolved.model == nil {
		return
	}
	status := int16(200)
	if callErr != nil {
		status = int16(500)
		var re *errorx.RespErr
		if errors.As(callErr, &re) && re != nil && re.Code > 0 && re.Code <= 32767 {
			status = int16(re.Code)
		}
	}
	stage := stageCode
	var modelCfgID *uint
	if resolved.model.ID > 0 {
		id := resolved.model.ID
		modelCfgID = &id
	}
	var bindingID *uint
	if resolved.binding != nil && resolved.binding.ID > 0 {
		id := resolved.binding.ID
		bindingID = &id
	}
	var promptTplID *uint
	if resolved.stagePrompt != nil && resolved.stagePrompt.ID > 0 {
		id := resolved.stagePrompt.ID
		promptTplID = &id
	} else if resolved.defPrompt != nil && resolved.defPrompt.ID > 0 {
		id := resolved.defPrompt.ID
		promptTplID = &id
	}
	var promptTok, completionTok *int
	if usage.PromptTokens > 0 {
		promptTok = &usage.PromptTokens
	}
	if usage.CompletionTokens > 0 {
		completionTok = &usage.CompletionTokens
	}
	row := &entity.PaperLLMCallLog{
		ManuscriptID:      manuscriptID,
		StageCode:         &stage,
		ModelConfigID:     modelCfgID,
		WorkflowBindingID: bindingID,
		PromptTemplateID:  promptTplID,
		ModelName:         resolved.model.ModelName,
		PromptTokens:      promptTok,
		CompletionTokens:  completionTok,
		LatencyMs:         latencyMs,
		Status:            status,
		CreatedAt:         time.Now(),
	}
	_ = s.llmRepo.InsertCallLog(ctx, row)
}

func (s *TopicDiscoveryRunService) resolveStageLLM(ctx context.Context, stageCode string, vars map[string]string) (*stageLLMResolve, error) {
	binding, err := s.llmRepo.GetActiveBindingByStage(ctx, stageCode)
	if err != nil {
		return nil, err
	}
	if binding == nil {
		binding, err = s.llmRepo.GetActiveBindingByStage(ctx, "default")
		if err != nil || binding == nil {
			return nil, errorx.ErrLLMNotConfigured.WithDetail(
				"缺少 paper_llm_workflow_binding（stage=" + stageCode + " 或 default）",
			)
		}
	}
	model, err := s.llmRepo.GetActiveModelByID(ctx, binding.ModelConfigID)
	if err != nil || model == nil {
		return nil, errorx.ErrLLMNotConfigured.WithDetail(
			"paper_llm_model_config 不存在或未 active（binding model_config_id=" +
				fmt.Sprintf("%d", binding.ModelConfigID) + "）",
		)
	}
	defPrompt, _ := s.llmRepo.GetActivePromptByStage(ctx, "default")
	stagePrompt, _ := s.llmRepo.GetActivePromptByStage(ctx, stageCode)
	system := ""
	if defPrompt != nil {
		system = renderPromptTemplate(defPrompt.TemplateBody, vars)
	}
	userTmpl := ""
	if stagePrompt != nil {
		userTmpl = renderPromptTemplate(stagePrompt.TemplateBody, vars)
	}
	if userTmpl == "" {
		userTmpl = stageCode
	}
	return &stageLLMResolve{
		model:       model,
		binding:     binding,
		defPrompt:   defPrompt,
		stagePrompt: stagePrompt,
		system:      system,
		userPrefix:  userTmpl,
	}, nil
}

func (s *TopicDiscoveryRunService) reloadView(ctx context.Context, userID uint, runVersion int) (*response.TopicDiscoveryRunView, error) {
	rows, err := s.steps.ListByUserRun(ctx, uint64(userID), runVersion)
	if err != nil {
		return nil, err
	}
	return s.buildRunView(manuscriptIDOf(rows), runVersion, rows), nil
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
		sv := response.TopicDiscoveryStepView{
			StageCode: r.StageCode,
			Status:    r.Status,
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
		if r.StartedAt != nil {
			t := r.StartedAt.Format(time.RFC3339)
			sv.StartedAt = &t
		}
		if r.CompletedAt != nil {
			t := r.CompletedAt.Format(time.RFC3339)
			sv.CompletedAt = &t
		}
		view.Steps = append(view.Steps, sv)

		switch r.Status {
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
			if r.Status == "completed" && i+1 < len(rows) && rows[i+1].Status == "pending" {
				view.PauseAfterStage = r.StageCode
				runStatus = "checkpoint"
				break
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

func mustJSON(v any) datatypes.JSON {
	b, _ := json.Marshal(v)
	return datatypes.JSON(b)
}

func ptrString(s string) *string { return &s }

func appendUsageMeta(step *entity.PaperOutputTopicStep, usage LLMUsage) {
	meta := map[string]any{}
	if len(step.Extra) > 0 {
		_ = json.Unmarshal(step.Extra, &meta)
	}
	meta["tokens_prompt"] = usage.PromptTokens
	meta["tokens_completion"] = usage.CompletionTokens
	step.Extra = mustJSON(meta)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
