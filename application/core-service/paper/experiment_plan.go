package paper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/datatypes"
)

const (
	LitReviewStatusGeneratingExperimentPlan = "generating_experiment_plan"
	ExperimentPlanStageCode                 = "experiment_plan"
)

type ExperimentPlanService struct {
	reviews     repository.PaperOutputLiteratureReviewRepo
	plans       repository.PaperOutputExperimentPlanRepo
	manuscripts repository.PaperManuscriptRepo
	llm         *LLMChatService
	llmRepo     repository.PaperLLMRepo
	llmRunner   *TopicDiscoveryRunService
}

func NewExperimentPlanService(
	reviews repository.PaperOutputLiteratureReviewRepo,
	plans repository.PaperOutputExperimentPlanRepo,
	manuscripts repository.PaperManuscriptRepo,
	llm *LLMChatService,
	llmRepo repository.PaperLLMRepo,
) *ExperimentPlanService {
	runner := &TopicDiscoveryRunService{llm: llm, llmRepo: llmRepo}
	return &ExperimentPlanService{
		reviews:     reviews,
		plans:       plans,
		manuscripts: manuscripts,
		llm:         llm,
		llmRepo:     llmRepo,
		llmRunner:   runner,
	}
}

// EnqueueGenerate 立即将文献综述标为 generating_experiment_plan，由 bot 异步调 LLM。
func (s *ExperimentPlanService) EnqueueGenerate(
	ctx context.Context,
	userID uint,
	manuscriptID, reviewID uint64,
) error {
	if manuscriptID == 0 || reviewID == 0 {
		return errorx.ErrParamsError
	}
	ms, err := s.manuscripts.GetByIDForUser(ctx, uint(manuscriptID), userID)
	if err != nil {
		return err
	}
	if ms == nil {
		return errorx.ErrParamsError.WithDetail("manuscript 不存在或无权访问")
	}
	row, err := s.reviews.GetByIDForUser(ctx, reviewID, manuscriptID, userID)
	if err != nil {
		return err
	}
	if row == nil {
		return errorx.ErrParamsError.WithDetail("文献综述不存在或无权操作")
	}
	st := strings.TrimSpace(row.Status)
	if st == "deleted" {
		return errorx.ErrParamsError.WithDetail("文献综述已删除")
	}
	if st == LitReviewStatusGeneratingExperimentPlan {
		return errorx.ErrParamsError.WithDetail("实验方案生成中，请稍候")
	}
	if st != "completed" {
		return errorx.ErrParamsError.WithDetail("仅已完成的文献综述可生成实验方案")
	}
	ok, err := s.reviews.TryBeginExperimentPlanGeneration(ctx, reviewID, manuscriptID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return errorx.ErrParamsError.WithDetail("无法开始生成实验方案")
	}
	return nil
}

// RunExperimentPlanLLM bot：对 generating_experiment_plan 的综述调用 LLM 并落库。
func (s *ExperimentPlanService) RunExperimentPlanLLM(ctx context.Context, review *entity.PaperOutputLiteratureReview) error {
	if review == nil || s.plans == nil || s.llmRunner == nil {
		return fmt.Errorf("experiment plan 服务未配置")
	}
	direction, venue := directionVenueFromReview(review)
	userExtra := buildExperimentPlanUserMessage(review, direction, venue)
	step := &entity.PaperOutputTopicStep{ManuscriptID: review.ManuscriptID}
	text, _, err := s.llmRunner.callStageLLM(ctx, step, ExperimentPlanStageCode, map[string]string{
		"direction": direction,
		"venue":     venue,
	}, userExtra)
	if err != nil {
		return s.markExperimentPlanFailed(ctx, review, err.Error())
	}
	planPayload := parseExperimentPlanLLMResponse(text)
	if planPayload == nil {
		return s.markExperimentPlanFailed(ctx, review, "模型未返回 experiment_plan JSON")
	}
	planRow, err := s.persistExperimentPlanFromLLM(ctx, review, planPayload, direction, venue)
	if err != nil {
		return s.markExperimentPlanFailed(ctx, review, err.Error())
	}
	meta := mergeReviewMeta(review, map[string]any{
		"experiment_plan_id":      planRow.ID,
		"experiment_plan_error":   nil,
		"experiment_plan_done_at": planRow.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
	metaJSON, _ := json.Marshal(meta)
	ok, err := s.reviews.CompleteExperimentPlanLink(
		ctx, review.ID, review.ManuscriptID, uint(review.UserID), planRow.ID, metaJSON,
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
) error {
	meta := mergeReviewMeta(review, map[string]any{
		"experiment_plan_error": strings.TrimSpace(reason),
	})
	metaJSON, _ := json.Marshal(meta)
	_, err := s.reviews.FailExperimentPlanGeneration(
		ctx, review.ID, review.ManuscriptID, uint(review.UserID), metaJSON,
	)
	if err != nil {
		return err
	}
	return fmt.Errorf("%s", strings.TrimSpace(reason))
}

func directionVenueFromReview(review *entity.PaperOutputLiteratureReview) (direction, venue string) {
	if review == nil || len(review.InputParams) == 0 {
		return "", ""
	}
	var snap map[string]any
	_ = json.Unmarshal(review.InputParams, &snap)
	if snap == nil {
		return "", ""
	}
	direction = strFromAny(snap["direction"])
	venue = strFromAny(snap["venue"])
	return direction, venue
}

func buildExperimentPlanUserMessage(review *entity.PaperOutputLiteratureReview, direction, venue string) string {
	lit := map[string]any{
		"id":             review.ID,
		"version":        review.Version,
		"title":          ptrStr(review.Title),
		"summary":        ptrStr(review.Summary),
		"content_medium": ptrStr(review.ContentMedium),
		"structure":      ptrStr(review.Structure),
	}
	if len(review.Citations) > 0 {
		var cites any
		_ = json.Unmarshal(review.Citations, &cites)
		lit["citations"] = cites
	}
	b, _ := json.Marshal(map[string]any{
		"manuscript_id":    review.ManuscriptID,
		"direction":        direction,
		"venue":            venue,
		"literature_review": lit,
	})
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

experiment_plan 内文本使用中文。

INPUT:
%s`, string(b))
}

func parseExperimentPlanLLMResponse(text string) map[string]any {
	root, _ := parseJSONOrWrap(text, "raw")
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
	title := strFromAny(plan["title"])
	summary := strFromAny(plan["summary"])
	content := strFromAny(plan["content_medium"])
	if content == "" {
		content = strFromAny(plan["content"])
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

func mergeReviewMeta(review *entity.PaperOutputLiteratureReview, patch map[string]any) map[string]any {
	out := map[string]any{}
	if review != nil && len(review.Meta) > 0 {
		_ = json.Unmarshal(review.Meta, &out)
	}
	if out == nil {
		out = map[string]any{}
	}
	for k, v := range patch {
		out[k] = v
	}
	return out
}

func ptrStr(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}
