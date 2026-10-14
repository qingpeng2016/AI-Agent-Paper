package experimentplan

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/lang"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/listpage"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/llm"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/topicdiscovery"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
	"gorm.io/datatypes"
)

const (
	// 须 ≤ paper_output_literature_review.status VARCHAR(16)
	LitReviewStatusGeneratingExperimentPlan = "gen_exp_plan"
	LitReviewStatusExperimentPlanFailed     = "gen_exp_fail"
	ExperimentPlanStageCode                 = "experiment_plan"
)

type ExperimentPlanService struct {
	reviews     repository.PaperOutputLiteratureReviewRepo
	plans       repository.PaperOutputExperimentPlanRepo
	manuscripts repository.PaperManuscriptRepo
	llm         *llm.LLMChatService
	llmRepo     repository.PaperLLMRepo
	llmRunner   *topicdiscovery.TopicDiscoveryRunService
}

func NewExperimentPlanService(
	reviews repository.PaperOutputLiteratureReviewRepo,
	plans repository.PaperOutputExperimentPlanRepo,
	manuscripts repository.PaperManuscriptRepo,
	chat *llm.LLMChatService,
	llmRepo repository.PaperLLMRepo,
	llmRunner *topicdiscovery.TopicDiscoveryRunService,
) *ExperimentPlanService {
	return &ExperimentPlanService{
		reviews:     reviews,
		plans:       plans,
		manuscripts: manuscripts,
		llm:         chat,
		llmRepo:     llmRepo,
		llmRunner:   llmRunner,
	}
}

// EnqueueGenerate 实验方案已在选题发现生成包中产出；保留接口以提示前端。
func (s *ExperimentPlanService) EnqueueGenerate(
	ctx context.Context,
	userID uint,
	manuscriptID, reviewID uint64,
) error {
	_ = ctx
	_ = userID
	_ = manuscriptID
	_ = reviewID
	return errorx.ErrParamsError.WithDetail("实验方案已在选题发现步骤中生成；若缺失请重新运行选题发现")
}

// RunExperimentPlanLLM bot：对 generating_experiment_plan 的综述调用 LLM 并落库。
func (s *ExperimentPlanService) RunExperimentPlanLLM(ctx context.Context, review *entity.PaperOutputLiteratureReview) error {
	if review == nil || s.plans == nil || s.llmRunner == nil {
		return fmt.Errorf("experiment plan 服务未配置")
	}
	direction, venue := "", ""
	step := &entity.PaperOutputTopicStep{
		ManuscriptID: review.ManuscriptID,
		UserID:       uint64(review.UserID),
	}
	locale := s.llmRunner.ContentLocaleForStep(ctx, step)
	userExtra := buildExperimentPlanUserMessage(review, direction, venue, locale)
	vars := map[string]string{"direction": direction, "venue": venue}

	resolved, resolveErr := s.llmRunner.ResolveStageLLM(ctx, ExperimentPlanStageCode, vars, locale)
	if resolveErr != nil {
		return s.markExperimentPlanFailed(ctx, review, resolveErr.Error(), nil, nil)
	}
	systemPrompt := ""
	userPrefix := ""
	modelName := ""
	if resolved != nil {
		systemPrompt = strings.TrimSpace(resolved.System)
		userPrefix = strings.TrimSpace(resolved.UserPrefix)
		if resolved.Model != nil {
			modelName = resolved.Model.ModelName
		}
	}
	userPrompt := topicdiscovery.LLMUserMessage(userPrefix, userExtra)
	reqSnap, _ := json.Marshal(map[string]any{
		"stage_code":    ExperimentPlanStageCode,
		"vars":          vars,
		"system_prompt": systemPrompt,
		"user_prompt":   userPrompt,
		"model_name":    modelName,
	})
	_, _ = s.reviews.SaveExperimentPlanLLMRequest(
		ctx, review.ID, review.ManuscriptID, uint(review.UserID), reqSnap,
	)

	text, usage, err := s.llmRunner.CallStageLLM(ctx, step, ExperimentPlanStageCode, vars, userExtra)
	respBase := map[string]any{
		"raw_text":          text,
		"prompt_tokens":     usage.PromptTokens,
		"completion_tokens": usage.CompletionTokens,
	}
	if err != nil {
		respBase["error"] = err.Error()
		respJSON, _ := json.Marshal(respBase)
		return s.markExperimentPlanFailed(ctx, review, err.Error(), respJSON, reqSnap)
	}
	planPayload := parseExperimentPlanLLMResponse(text)
	if planPayload == nil {
		respBase["error"] = "模型未返回 experiment_plan JSON"
		respJSON, _ := json.Marshal(respBase)
		return s.markExperimentPlanFailed(ctx, review, "模型未返回 experiment_plan JSON", respJSON, reqSnap)
	}
	respBase["parsed"] = map[string]any{"experiment_plan": planPayload}
	respJSON, _ := json.Marshal(respBase)

	planRow, err := s.persistExperimentPlanFromLLM(ctx, review, planPayload, direction, venue)
	if err != nil {
		respBase["error"] = err.Error()
		respJSON, _ = json.Marshal(respBase)
		return s.markExperimentPlanFailed(ctx, review, err.Error(), respJSON, reqSnap)
	}
	ok, err := s.reviews.CompleteExperimentPlanLink(
		ctx, review.ID, review.ManuscriptID, uint(review.UserID), planRow.ID, respJSON,
	)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("文献综述状态已变更，无法关联实验方案")
	}
	return nil
}

func (s *ExperimentPlanService) markExperimentPlanFailed(
	ctx context.Context,
	review *entity.PaperOutputLiteratureReview,
	reason string,
	llmResponseJSON []byte,
	llmRequestJSON []byte,
) error {
	if len(llmRequestJSON) > 0 {
		_, _ = s.reviews.SaveExperimentPlanLLMRequest(
			ctx, review.ID, review.ManuscriptID, uint(review.UserID), llmRequestJSON,
		)
	}
	_, err := s.reviews.FailExperimentPlanGeneration(
		ctx, review.ID, review.ManuscriptID, uint(review.UserID), llmResponseJSON,
	)
	if err != nil {
		return err
	}
	return fmt.Errorf("%s", strings.TrimSpace(reason))
}

func buildExperimentPlanUserMessage(review *entity.PaperOutputLiteratureReview, direction, venue, locale string) string {
	locale = lang.NormalizeContentLanguage(locale)
	lit := map[string]any{
		"id":             review.ID,
		"version":        review.Version,
		"title":          ptrStr(review.Title),
		"content_medium": ptrStr(review.ContentMedium),
	}
	b, _ := json.Marshal(map[string]any{
		"manuscript_id":    review.ManuscriptID,
		"direction":        direction,
		"venue":            venue,
		"literature_review": lit,
	})
	if locale == "en" {
		return fmt.Sprintf(`Generate an experiment plan from the LiteratureReview below.

Return ONLY valid JSON:
{
  "experiment_plan": {
    "title": "string",
    "summary": "string",
    "content_medium": "string (Markdown body)",
    "meta": { "hypothesis": "", "baselines": [], "metrics": [], "ablations": [], "timeline": [] }
  }
}

%s

INPUT:
%s`, lang.JSONStringFieldsLocaleHint(locale), string(b))
	}
	return fmt.Sprintf(`请基于下列 LiteratureReview 生成实验方案。

输出 ONLY valid JSON：
{
  "experiment_plan": {
    "title": "string",
    "summary": "string",
    "content_medium": "string（Markdown 正文）",
    "meta": { "hypothesis": "", "baselines": [], "metrics": [], "ablations": [], "timeline": [] }
  }
}

%s

INPUT:
%s`, lang.JSONStringFieldsLocaleHint(locale), string(b))
}

func parseExperimentPlanLLMResponse(text string) map[string]any {
	root, _ := topicdiscovery.ParseJSONOrWrap(text, "raw")
	if root == nil {
		return nil
	}
	if ep, ok := root["experiment_plan"].(map[string]any); ok && len(ep) > 0 {
		return ep
	}
	return nil
}

func (s *ExperimentPlanService) persistExperimentPlanFromLLM(
	ctx context.Context,
	review *entity.PaperOutputLiteratureReview,
	plan map[string]any,
	direction, venue string,
) (*entity.PaperOutputExperimentPlan, error) {
	title := topicdiscovery.StrFromAny(plan["title"])
	summary := topicdiscovery.StrFromAny(plan["summary"])
	content := topicdiscovery.StrFromAny(plan["content_medium"])
	if content == "" {
		content = topicdiscovery.StrFromAny(plan["content"])
	}
	var meta datatypes.JSON
	if raw, ok := plan["meta"]; ok && raw != nil {
		b, _ := json.Marshal(raw)
		meta = b
	}
	inputSnap, _ := json.Marshal(map[string]any{
		"literature_review_id": review.ID,
		"direction":            direction,
		"venue":                venue,
		"stage_code":           ExperimentPlanStageCode,
	})
	row := &entity.PaperOutputExperimentPlan{
		ManuscriptID: review.ManuscriptID,
		UserID:       review.UserID,
		Status:       "completed",
		Format:       "md",
		InputParams:  inputSnap,
		Meta:         meta,
	}
	if title != "" {
		row.Title = &title
	}
	if summary != "" {
		row.Summary = &summary
	}
	if content != "" {
		row.ContentMedium = &content
	}
	if err := s.plans.Create(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *ExperimentPlanService) ListByManuscript(
	ctx context.Context,
	userID uint,
	manuscriptID uint64,
	page, pageSize int,
) (*response.PaperExperimentPlanListView, error) {
	if manuscriptID == 0 {
		return nil, errorx.ErrParamsError
	}
	ms, err := s.manuscripts.GetByIDForUser(ctx, uint(manuscriptID), userID)
	if err != nil {
		return nil, err
	}
	if ms == nil {
		return nil, errorx.ErrParamsError.WithDetail("manuscript 不存在或无权访问")
	}
	page, pageSize, offset := listpage.NormalizeListPage(page, pageSize)
	total, err := s.plans.CountByManuscript(ctx, manuscriptID)
	if err != nil {
		return nil, err
	}
	rows, err := s.plans.ListByManuscriptPaged(ctx, manuscriptID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	planIDs := make([]uint64, 0, len(rows))
	for _, row := range rows {
		planIDs = append(planIDs, row.ID)
	}
	lrByPlan, err := s.reviews.MapLiteratureReviewIDsByExperimentPlanIDs(ctx, manuscriptID, planIDs)
	if err != nil {
		return nil, err
	}
	out := &response.PaperExperimentPlanListView{
		ManuscriptID: manuscriptID,
		Items:        make([]response.PaperExperimentPlanItemView, 0, len(rows)),
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
	}
	for _, row := range rows {
		item := experimentPlanListItemView(row)
		item = enrichExperimentPlanLiteratureReviewID(item, lrByPlan[row.ID])
		out.Items = append(out.Items, item)
	}
	return out, nil
}

func enrichExperimentPlanLiteratureReviewID(
	item response.PaperExperimentPlanItemView,
	lrID uint64,
) response.PaperExperimentPlanItemView {
	if strings.TrimSpace(item.LiteratureReviewID) != "" || lrID == 0 {
		return item
	}
	item.LiteratureReviewID = strconv.FormatUint(lrID, 10)
	return item
}

func (s *ExperimentPlanService) enrichExperimentPlanLiteratureReviewID(
	ctx context.Context,
	manuscriptID, planID uint64,
	item response.PaperExperimentPlanItemView,
) response.PaperExperimentPlanItemView {
	if strings.TrimSpace(item.LiteratureReviewID) != "" {
		return item
	}
	lrID, err := s.reviews.GetIDByExperimentPlanID(ctx, planID, manuscriptID)
	if err != nil || lrID == 0 {
		return item
	}
	return enrichExperimentPlanLiteratureReviewID(item, lrID)
}

func (s *ExperimentPlanService) GetDetail(
	ctx context.Context,
	userID uint,
	manuscriptID, planID uint64,
) (*response.PaperExperimentPlanItemView, error) {
	if manuscriptID == 0 || planID == 0 {
		return nil, errorx.ErrParamsError
	}
	ms, err := s.manuscripts.GetByIDForUser(ctx, uint(manuscriptID), userID)
	if err != nil {
		return nil, err
	}
	if ms == nil {
		return nil, errorx.ErrParamsError.WithDetail("manuscript 不存在或无权访问")
	}
	row, err := s.plans.GetByIDForManuscript(ctx, planID, manuscriptID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errorx.ErrParamsError.WithDetail("实验方案不存在")
	}
	if row.Status == "deleted" {
		return nil, errorx.ErrParamsError.WithDetail("实验方案已删除")
	}
	item := experimentPlanItemView(*row)
	item = s.enrichExperimentPlanLiteratureReviewID(ctx, manuscriptID, planID, item)
	return &item, nil
}

func experimentPlanListItemView(row entity.PaperOutputExperimentPlan) response.PaperExperimentPlanItemView {
	item := experimentPlanItemView(row)
	item.ContentMedium = ""
	return item
}

func experimentPlanItemView(row entity.PaperOutputExperimentPlan) response.PaperExperimentPlanItemView {
	item := response.PaperExperimentPlanItemView{
		ID:        strconv.FormatUint(row.ID, 10),
		Version:   row.Version,
		Status:    row.Status,
		Format:    row.Format,
		CreatedAt: row.CreatedAt.UTC().Format(time.RFC3339),
	}
	if row.Title != nil {
		item.Title = *row.Title
	}
	if row.Summary != nil {
		item.Summary = *row.Summary
	}
	if row.ContentMedium != nil {
		item.ContentMedium = *row.ContentMedium
	}
	if row.ExperimentDataURI != nil && strings.TrimSpace(*row.ExperimentDataURI) != "" {
		item.ExperimentDataURI = strings.TrimSpace(*row.ExperimentDataURI)
	}
	if len(row.InputParams) > 0 {
		var snap map[string]any
		_ = json.Unmarshal(row.InputParams, &snap)
		if id := topicdiscovery.StrFromAny(snap["literature_review_id"]); id != "" {
			item.LiteratureReviewID = id
		}
	}
	return item
}

func ptrStr(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}
