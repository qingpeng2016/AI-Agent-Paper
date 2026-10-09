package paper

import (
	"context"
	"encoding/json"
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

func (s *TopicDiscoveryRunService) GetCurrentRun(ctx context.Context, userID uint, manuscriptID uint64) (*response.TopicDiscoveryRunView, error) {
	if err := s.assertManuscript(ctx, userID, manuscriptID); err != nil {
		return nil, err
	}
	runVersion, stepRows, err := s.steps.GetCurrentRun(ctx, manuscriptID)
	if err != nil {
		return nil, err
	}
	if runVersion == 0 || len(stepRows) == 0 {
		return nil, nil
	}
	return s.buildRunView(manuscriptID, runVersion, stepRows), nil
}

func (s *TopicDiscoveryRunService) Run(ctx context.Context, userID uint, req request.TopicDiscoveryRunRequest) (*response.TopicDiscoveryRunView, error) {
	msID, input, err := s.prepareRun(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if action == "" {
		action = "start"
	}

	var runVersion int
	switch action {
	case "start":
		paramsBytes, _ := json.Marshal(input)
		runVersion, err = s.steps.BeginRun(ctx, msID, uint64(userID), paramsBytes)
		if err != nil {
			return nil, err
		}
		if err := s.executeFromStage(ctx, msID, runVersion, input, 0); err != nil {
			return nil, err
		}
	case "continue":
		runVersion, _, err = s.steps.GetCurrentRun(ctx, msID)
		if err != nil || runVersion == 0 {
			return nil, errorx.ErrTopicRunNotFound
		}
		input, err = s.loadRunInput(ctx, msID, runVersion)
		if err != nil {
			return nil, err
		}
		idx := s.nextPendingIndex(ctx, msID, runVersion)
		if idx < 0 {
			return s.reloadView(ctx, msID, runVersion)
		}
		if err := s.executeFromStage(ctx, msID, runVersion, input, idx); err != nil {
			return nil, err
		}
	case "run_all":
		runVersion, stepRows, err := s.steps.GetCurrentRun(ctx, msID)
		if err != nil {
			return nil, err
		}
		if runVersion == 0 {
			paramsBytes, _ := json.Marshal(input)
			runVersion, err = s.steps.BeginRun(ctx, msID, uint64(userID), paramsBytes)
			if err != nil {
				return nil, err
			}
		} else {
			input, err = s.loadRunInput(ctx, msID, runVersion)
			if err != nil {
				return nil, err
			}
			_ = stepRows
		}
		for {
			idx := s.nextPendingIndex(ctx, msID, runVersion)
			if idx < 0 {
				break
			}
			if err := s.executeStage(ctx, msID, runVersion, input, topicDiscoveryStages[idx]); err != nil {
				return nil, err
			}
			if input.HumanCheckpoint {
				view, err := s.reloadView(ctx, msID, runVersion)
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

	return s.reloadView(ctx, msID, runVersion)
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

func (s *TopicDiscoveryRunService) prepareRun(ctx context.Context, userID uint, req request.TopicDiscoveryRunRequest) (uint64, topicRunInput, error) {
	direction := strings.TrimSpace(req.Direction)
	if direction == "" {
		return 0, topicRunInput{}, errorx.ErrParamsError
	}
	if len(req.SourceCodes) == 0 {
		return 0, topicRunInput{}, errorx.ErrParamsError
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
		return 0, topicRunInput{}, errorx.ErrParamsError
	}
	audit, err := s.catalog.FindAuditLevelByCode(ctx, auditCode)
	if err != nil || audit == nil {
		return 0, topicRunInput{}, errorx.ErrParamsError
	}

	msID, err := s.ensureManuscript(ctx, userID, req.ManuscriptID, req.ManuscriptTitle, direction)
	if err != nil {
		return 0, topicRunInput{}, err
	}

	input := topicRunInput{
		DisciplineCode:  strings.TrimSpace(req.DisciplineCode),
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
	return msID, input, nil
}

func (s *TopicDiscoveryRunService) ensureManuscript(ctx context.Context, userID uint, manuscriptID uint64, title, direction string) (uint64, error) {
	if manuscriptID > 0 {
		ms, err := s.manuscripts.GetByIDForUser(ctx, uint(manuscriptID), userID)
		if err != nil {
			return 0, err
		}
		if ms == nil {
			return 0, errorx.ErrManuscriptNotFound
		}
		return uint64(ms.ID), nil
	}
	t := strings.TrimSpace(title)
	if t == "" {
		t = direction
		if len([]rune(t)) > 120 {
			t = string([]rune(t)[:120])
		}
	}
	ms, err := s.manuscripts.Create(ctx, userID, t)
	if err != nil {
		return 0, err
	}
	return uint64(ms.ID), nil
}

func (s *TopicDiscoveryRunService) assertManuscript(ctx context.Context, userID uint, manuscriptID uint64) error {
	ms, err := s.manuscripts.GetByIDForUser(ctx, uint(manuscriptID), userID)
	if err != nil {
		return err
	}
	if ms == nil {
		return errorx.ErrManuscriptNotFound
	}
	return nil
}

func (s *TopicDiscoveryRunService) loadRunInput(ctx context.Context, manuscriptID uint64, runVersion int) (topicRunInput, error) {
	step, err := s.steps.GetStep(ctx, manuscriptID, runVersion, "retrieve")
	if err != nil || step == nil || len(step.InputParams) == 0 {
		return topicRunInput{}, errorx.ErrTopicRunNotFound
	}
	var input topicRunInput
	if err := json.Unmarshal(step.InputParams, &input); err != nil {
		return topicRunInput{}, errorx.ErrTopicRunNotFound
	}
	return input, nil
}

func (s *TopicDiscoveryRunService) nextPendingIndex(ctx context.Context, manuscriptID uint64, runVersion int) int {
	for i, stage := range topicDiscoveryStages {
		step, err := s.steps.GetStep(ctx, manuscriptID, runVersion, stage)
		if err != nil || step == nil {
			continue
		}
		if step.Status == "pending" || step.Status == "failed" {
			return i
		}
	}
	return -1
}

func (s *TopicDiscoveryRunService) executeFromStage(ctx context.Context, manuscriptID uint64, runVersion int, input topicRunInput, fromIndex int) error {
	for i := fromIndex; i < len(topicDiscoveryStages); i++ {
		if err := s.executeStage(ctx, manuscriptID, runVersion, input, topicDiscoveryStages[i]); err != nil {
			return err
		}
		if input.HumanCheckpoint {
			return nil
		}
	}
	return nil
}

func (s *TopicDiscoveryRunService) executeStage(ctx context.Context, manuscriptID uint64, runVersion int, input topicRunInput, stageCode string) error {
	step, err := s.steps.GetStep(ctx, manuscriptID, runVersion, stageCode)
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
		execErr = s.stageGenerateIdeas(ctx, manuscriptID, runVersion, step, input)
	case "novelty":
		execErr = s.stageNovelty(ctx, manuscriptID, runVersion, step, input)
	case "audit":
		execErr = s.stageAudit(ctx, manuscriptID, runVersion, step, input)
	default:
		execErr = errorx.ErrTopicStepInvalid
	}

	done := time.Now()
	step.CompletedAt = &done
	if execErr != nil {
		step.Status = "failed"
		meta := map[string]any{"error": execErr.Error()}
		step.Meta = mustJSON(meta)
		_ = s.steps.SaveStep(ctx, step)
		return execErr
	}
	step.Status = "completed"
	return s.steps.SaveStep(ctx, step)
}

func (s *TopicDiscoveryRunService) stageRetrieve(ctx context.Context, step *entity.PaperOutputTopicStep, input topicRunInput) error {
	query := ExtractLiteratureSearchQuery(input.Direction)
	if query == "" {
		query = strings.TrimSpace(input.Direction)
	}
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
	step.Result = mustJSON(result)
	step.Meta = mustJSON(map[string]any{
		"hit_count":        len(hits),
		"verified_count":   len(hits),
		"search_query":     query,
		"literature_links": links,
	})

	titles := make([]string, 0, min(12, len(hits)))
	for i, h := range hits {
		if i >= 12 {
			break
		}
		if h.URL != "" {
			titles = append(titles, fmt.Sprintf("- %s (%s) %s", h.Title, h.ExternalKey, h.URL))
		} else {
			titles = append(titles, fmt.Sprintf("- %s (%s)", h.Title, h.ExternalKey))
		}
	}
	contextBlock := strings.Join(titles, "\n")
	if contextBlock == "" {
		contextBlock = "(no literature hits from configured sources)"
	}

	summary, usage, err := s.callStageLLM(ctx, "retrieve", map[string]string{
		"direction": input.Direction,
		"venue":     input.Venue,
	}, contextBlock+"\n\nSummarize coverage and gaps in concise Chinese.")
	if err != nil {
		step.SummaryText = ptrString(fmt.Sprintf("检索完成，共 %d 篇文献（模型摘要未生成）", len(hits)))
		return nil
	}
	step.SummaryText = ptrString(summary)
	appendUsageMeta(step, usage)
	return nil
}

func (s *TopicDiscoveryRunService) stageGenerateIdeas(ctx context.Context, manuscriptID uint64, runVersion int, step *entity.PaperOutputTopicStep, input topicRunInput) error {
	ctxBlock, _ := s.stageContext(ctx, manuscriptID, runVersion)
	userMsg := fmt.Sprintf(`Research direction: %s
Target venue: %s
Max ideas: %d

Corpus:
%s

Return ONLY valid JSON: {"ideas":[{"title":"","problem":"","approach":"","contribution":""}]}
Use Chinese for text fields.`, input.Direction, input.Venue, input.MaxIdeas, ctxBlock)

	text, usage, err := s.callStageLLM(ctx, "generate_ideas", map[string]string{
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

func (s *TopicDiscoveryRunService) stageNovelty(ctx context.Context, manuscriptID uint64, runVersion int, step *entity.PaperOutputTopicStep, input topicRunInput) error {
	ctxBlock, _ := s.stageContext(ctx, manuscriptID, runVersion)
	userMsg := fmt.Sprintf(`Perform novelty analysis for direction: %s

Prior steps:
%s

Return ONLY valid JSON: {"lines":["..."],"risks":[{"idea_title":"","risk":"low|medium|high","note":""}]}
Use Chinese.`, input.Direction, ctxBlock)

	text, usage, err := s.callStageLLM(ctx, "novelty", map[string]string{"direction": input.Direction}, userMsg)
	if err != nil {
		return err
	}
	result, summary := parseJSONOrWrap(text, "novelty")
	step.Result = mustJSON(result)
	step.SummaryText = ptrString(summary)
	appendUsageMeta(step, usage)
	return nil
}

func (s *TopicDiscoveryRunService) stageAudit(ctx context.Context, manuscriptID uint64, runVersion int, step *entity.PaperOutputTopicStep, input topicRunInput) error {
	ctxBlock, _ := s.stageContext(ctx, manuscriptID, runVersion)
	rounds := input.AuditRounds
	if rounds < 1 {
		rounds = 1
	}
	var roundOutputs []map[string]any
	var lastText string
	var totalUsage LLMUsage
	for r := 1; r <= rounds; r++ {
		userMsg := fmt.Sprintf(`Audit round %d/%d for venue %s (audit level %s).

Context:
%s

Return ONLY valid JSON: {"issues":[{"severity":"blocker|major|minor","claim":"","fix":""}],"summary":""}
Use Chinese.`, r, rounds, input.Venue, input.AuditLevel, ctxBlock)
		text, usage, err := s.callStageLLM(ctx, "audit", map[string]string{"direction": input.Direction}, userMsg)
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

func (s *TopicDiscoveryRunService) stageContext(ctx context.Context, manuscriptID uint64, runVersion int) (string, error) {
	var parts []string
	for _, stage := range topicDiscoveryStages {
		if stage == "audit" {
			continue
		}
		st, err := s.steps.GetStep(ctx, manuscriptID, runVersion, stage)
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

func (s *TopicDiscoveryRunService) callStageLLM(ctx context.Context, stageCode string, vars map[string]string, userExtra string) (string, LLMUsage, error) {
	model, system, stagePrompt, err := s.resolveStageLLM(ctx, stageCode, vars)
	if err != nil {
		return "", LLMUsage{}, err
	}
	user := stagePrompt
	if strings.TrimSpace(userExtra) != "" {
		user = stagePrompt + "\n\n" + userExtra
	}
	return s.llm.Complete(ctx, model, system, user)
}

func (s *TopicDiscoveryRunService) resolveStageLLM(ctx context.Context, stageCode string, vars map[string]string) (*entity.PaperLLMModelConfig, string, string, error) {
	binding, err := s.llmRepo.GetActiveBindingByStage(ctx, stageCode)
	if err != nil {
		return nil, "", "", err
	}
	if binding == nil {
		binding, err = s.llmRepo.GetActiveBindingByStage(ctx, "default")
		if err != nil || binding == nil {
			return nil, "", "", errorx.ErrLLMNotConfigured
		}
	}
	model, err := s.llmRepo.GetActiveModelByID(ctx, binding.ModelConfigID)
	if err != nil || model == nil {
		return nil, "", "", errorx.ErrLLMNotConfigured
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
	return model, system, userTmpl, nil
}

func (s *TopicDiscoveryRunService) reloadView(ctx context.Context, manuscriptID uint64, runVersion int) (*response.TopicDiscoveryRunView, error) {
	rows, err := s.steps.ListByRun(ctx, manuscriptID, runVersion)
	if err != nil {
		return nil, err
	}
	return s.buildRunView(manuscriptID, runVersion, rows), nil
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
		if r.SummaryText != nil {
			sv.SummaryText = *r.SummaryText
		}
		if len(r.Result) > 0 {
			sv.Result = json.RawMessage(r.Result)
		}
		if len(r.Meta) > 0 {
			sv.Meta = json.RawMessage(r.Meta)
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
	if len(step.Meta) > 0 {
		_ = json.Unmarshal(step.Meta, &meta)
	}
	meta["tokens_prompt"] = usage.PromptTokens
	meta["tokens_completion"] = usage.CompletionTokens
	step.Meta = mustJSON(meta)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
