package topicdiscovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/internal/jsonx"
	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/request"
	"github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
)

// ConfirmTopicRound 第三步：用户选定迭代轮次后，写入文献综述与实验方案表并完成 run。
func (s *TopicDiscoveryRunService) ConfirmTopicRound(
	ctx context.Context,
	userID uint,
	req request.TopicDiscoveryConfirmRoundRequest,
) (*response.TopicDiscoveryRunView, error) {
	if req.ManuscriptID == 0 {
		return nil, errorx.ErrParamsError.WithDetail("缺少 manuscript_id")
	}
	if err := s.validateManuscriptForUser(ctx, userID, req.ManuscriptID); err != nil {
		return nil, err
	}
	runVersion := req.RunVersion
	if runVersion <= 0 {
		ver, rows, err := s.latestRunForManuscript(ctx, userID, req.ManuscriptID)
		if err != nil {
			return nil, err
		}
		if ver == 0 || len(rows) == 0 || currentRunCancelled(rows) {
			return nil, errorx.ErrTopicRunNotFound
		}
		runVersion = ver
	}

	discStep, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "idea_discovery")
	if err != nil || discStep == nil {
		discStep, err = s.steps.GetStep(ctx, uint64(userID), runVersion, "topic_discovery")
	}
	if err != nil || discStep == nil {
		discStep, err = s.steps.GetStep(ctx, uint64(userID), runVersion, "research_refine")
	}
	if err != nil || discStep == nil {
		return nil, errorx.ErrTopicStepInvalid
	}
	if strings.TrimSpace(discStep.Status) != "completed" {
		return nil, errorx.ErrTopicStepInvalid.WithDetail("选题流水线尚未完成，无法确认")
	}

	confirmStep, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "topic_confirm")
	if err != nil || confirmStep == nil {
		return nil, errorx.ErrTopicStepInvalid.WithDetail("缺少确认步骤（请使用新版选题 run）")
	}
	selectedRound := req.SelectedRound
	if selectedRound < 1 {
		return nil, errorx.ErrParamsError.WithDetail("请选择有效轮次")
	}

	var discResult map[string]any
	if agg, aggErr := s.BuildIdeaDiscoveryAggregateResult(ctx, userID, runVersion); aggErr == nil && len(agg) > 0 {
		discResult = agg
	} else if err := json.Unmarshal(discStep.Result, &discResult); err != nil || len(discResult) == 0 {
		return nil, errorx.ErrTopicStepInvalid.WithDetail("选题发现结果为空")
	}
	if selectedRound < 1 {
		selectedRound = 1
	}
	iter, generate, auditBlock, passed, err := topicIterationByRound(discResult, selectedRound)
	if err != nil {
		return nil, err
	}
	_ = iter

	input, err := s.loadRunInput(ctx, userID, runVersion)
	if err != nil {
		return nil, err
	}

	litPayload, _ := generate["literature_review"].(map[string]any)
	if litPayload == nil {
		litPayload = map[string]any{}
	}
	planPayload, _ := generate["experiment_plan"].(map[string]any)
	needsPlan := topicExperimentPlanPayloadNonEmpty(planPayload)

	if strings.TrimSpace(confirmStep.Status) == "completed" {
		if s.topicConfirmArtifactsReady(ctx, userID, discStep.ManuscriptID, confirmStep, selectedRound, needsPlan) {
			return s.reloadView(ctx, userID, runVersion)
		}
		confirmStep.Status = "pending"
		confirmStep.CompletedAt = nil
		if err := s.steps.SaveStep(ctx, confirmStep); err != nil {
			return nil, topicConfirmWrapErr(err)
		}
	} else if done, err := s.tryFinishConfirmStepIfPersisted(
		ctx, userID, discStep, confirmStep, selectedRound, needsPlan,
	); err != nil {
		return nil, err
	} else if done {
		return s.reloadView(ctx, userID, runVersion)
	}

	var litReviewID uint64
	if s.litReviews != nil && len(litPayload) > 0 {
		id, saveErr := s.persistLiteratureReviewFromLLM(ctx, discStep, input, userID, runVersion, litPayload)
		if saveErr != nil {
			return nil, topicConfirmWrapErr(saveErr)
		}
		litReviewID = id
	}

	var planID uint64
	if s.expPlans != nil && planPayload != nil && len(planPayload) > 0 {
		planRow, planErr := s.persistTopicExperimentPlan(ctx, userID, discStep, input, litReviewID, planPayload)
		if planErr != nil {
			return nil, topicConfirmWrapErr(planErr)
		}
		if planRow != nil {
			planID = planRow.ID
		}
	}
	if litReviewID > 0 && planID > 0 && s.litReviews != nil {
		_ = s.litReviews.LinkExperimentPlan(ctx, litReviewID, planID, uint64(userID), discStep.ManuscriptID)
	}

	confirmResult := map[string]any{
		"selected_round":    selectedRound,
		"passed":            passed,
		"ideas":             generate["ideas"],
		"novelty":           generate["novelty"],
		"literature_review": litPayload,
		"experiment_plan":   planPayload,
		"audit":             auditBlock,
	}
	if litReviewID > 0 {
		confirmResult["literature_review_id"] = litReviewID
	}
	if planID > 0 {
		confirmResult["experiment_plan_id"] = planID
	}

	now := time.Now()
	if confirmStep.StartedAt == nil {
		confirmStep.StartedAt = &now
	}
	confirmStep.CompletedAt = &now
	confirmStep.Status = "completed"
	confirmStep.Result = jsonx.MustJSON(confirmResult)

	if err := s.steps.SaveStep(ctx, confirmStep); err != nil {
		return nil, topicConfirmWrapErr(err)
	}

	if discStep.ManuscriptID > 0 {
		if err := s.manuscripts.SetCurrentForUser(ctx, userID, uint(discStep.ManuscriptID)); err != nil {
			return nil, topicConfirmWrapErr(err)
		}
	}

	return s.reloadView(ctx, userID, runVersion)
}

// tryFinishConfirmStepIfPersisted 仅当 DB 中综述/方案已完整落库时补全 topic_confirm（不信任 step 里孤立的 id）。
func (s *TopicDiscoveryRunService) tryFinishConfirmStepIfPersisted(
	ctx context.Context,
	userID uint,
	discStep *entity.PaperOutputTopicStep,
	confirmStep *entity.PaperOutputTopicStep,
	selectedRound int,
	needsPlan bool,
) (bool, error) {
	if !s.topicConfirmArtifactsReady(ctx, userID, discStep.ManuscriptID, confirmStep, selectedRound, needsPlan) {
		return false, nil
	}
	now := time.Now()
	if confirmStep.StartedAt == nil {
		confirmStep.StartedAt = &now
	}
	if confirmStep.CompletedAt == nil {
		confirmStep.CompletedAt = &now
	}
	confirmStep.Status = "completed"
	if err := s.steps.SaveStep(ctx, confirmStep); err != nil {
		return false, topicConfirmWrapErr(err)
	}
	if discStep.ManuscriptID > 0 {
		if err := s.manuscripts.SetCurrentForUser(ctx, userID, uint(discStep.ManuscriptID)); err != nil {
			return false, topicConfirmWrapErr(err)
		}
	}
	return true, nil
}

func (s *TopicDiscoveryRunService) topicConfirmArtifactsReady(
	ctx context.Context,
	userID uint,
	manuscriptID uint64,
	confirmStep *entity.PaperOutputTopicStep,
	selectedRound int,
	needsPlan bool,
) bool {
	if confirmStep == nil || manuscriptID == 0 {
		return false
	}
	meta := paperstep.LoadStepExtra(confirmStep)
	if intFromAny(meta["selected_round"]) != selectedRound {
		return false
	}
	litID := topicConfirmStoredLiteratureReviewID(confirmStep)
	if !s.topicConfirmLiteratureReady(ctx, userID, manuscriptID, litID) {
		return false
	}
	if !needsPlan {
		return true
	}
	planID := topicConfirmStoredPlanID(confirmStep)
	return s.topicConfirmExperimentPlanReady(ctx, userID, manuscriptID, planID)
}

func (s *TopicDiscoveryRunService) topicConfirmLiteratureReady(
	ctx context.Context,
	userID uint,
	manuscriptID, litID uint64,
) bool {
	if s.litReviews == nil || litID == 0 {
		return false
	}
	row, err := s.litReviews.GetByIDForUser(ctx, litID, manuscriptID, userID)
	if err != nil || row == nil {
		return false
	}
	if strings.TrimSpace(row.Status) != "completed" {
		return false
	}
	if strings.TrimSpace(row.From) != "generated" {
		return false
	}
	body := ""
	if row.ContentMedium != nil {
		body = strings.TrimSpace(*row.ContentMedium)
	}
	if body == "" {
		return false
	}
	return topicConfirmExportURIReady(row.PdfURI) && topicConfirmExportURIReady(row.WordURI)
}

func (s *TopicDiscoveryRunService) topicConfirmExperimentPlanReady(
	ctx context.Context,
	userID uint,
	manuscriptID, planID uint64,
) bool {
	if s.expPlans == nil || planID == 0 {
		return false
	}
	row, err := s.expPlans.GetByIDForUser(ctx, planID, manuscriptID, userID)
	if err != nil || row == nil {
		return false
	}
	if strings.TrimSpace(row.Status) != "completed" {
		return false
	}
	body := ""
	if row.ContentMedium != nil {
		body = strings.TrimSpace(*row.ContentMedium)
	}
	if body == "" {
		return false
	}
	return topicConfirmExportURIReady(row.PdfURI) && topicConfirmExportURIReady(row.WordURI)
}

func topicConfirmExportURIReady(uri *string) bool {
	return uri != nil && strings.TrimSpace(*uri) != ""
}

func topicExperimentPlanPayloadNonEmpty(plan map[string]any) bool {
	if len(plan) == 0 {
		return false
	}
	return strings.TrimSpace(experimentPlanBodyFromPayload(plan)) != ""
}

func topicConfirmStoredPlanID(step *entity.PaperOutputTopicStep) uint64 {
	if step == nil {
		return 0
	}
	meta := paperstep.LoadStepExtra(step)
	if id := uint64FromAny(meta["experiment_plan_id"]); id > 0 {
		return id
	}
	if len(step.Result) == 0 {
		return 0
	}
	var result map[string]any
	if json.Unmarshal(step.Result, &result) != nil {
		return 0
	}
	return uint64FromAny(result["experiment_plan_id"])
}

func topicConfirmWrapErr(err error) error {
	if err == nil {
		return nil
	}
	var biz *errorx.RespErr
	if errors.As(err, &biz) {
		return err
	}
	return errorx.ErrParamsError.WithDetail(err.Error())
}

func topicConfirmStoredLiteratureReviewID(step *entity.PaperOutputTopicStep) uint64 {
	if step == nil {
		return 0
	}
	meta := paperstep.LoadStepExtra(step)
	if id := uint64FromAny(meta["literature_review_id"]); id > 0 {
		return id
	}
	if len(step.Result) == 0 {
		return 0
	}
	var result map[string]any
	if json.Unmarshal(step.Result, &result) != nil {
		return 0
	}
	return uint64FromAny(result["literature_review_id"])
}

func uint64FromAny(v any) uint64 {
	switch t := v.(type) {
	case float64:
		if t > 0 {
			return uint64(t)
		}
	case int:
		if t > 0 {
			return uint64(t)
		}
	case int64:
		if t > 0 {
			return uint64(t)
		}
	case uint64:
		return t
	case json.Number:
		if n, err := t.Int64(); err == nil && n > 0 {
			return uint64(n)
		}
	}
	return 0
}

func topicIterationByRound(
	discResult map[string]any,
	round int,
) (iter map[string]any, generate map[string]any, audit map[string]any, passed bool, err error) {
	rawIters, _ := discResult["iterations"].([]any)
	for _, item := range rawIters {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		r := intFromAny(m["round"])
		if r != round {
			continue
		}
		gen, _ := m["generate"].(map[string]any)
		if gen == nil {
			gen = map[string]any{}
		}
		aud, _ := m["audit"].(map[string]any)
		p, _ := m["passed"].(bool)
		return m, gen, aud, p, nil
	}
	return nil, nil, nil, false, errorx.ErrParamsError.WithDetail(fmt.Sprintf("未找到第 %d 轮结果", round))
}

func intFromAny(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	default:
		return 0
	}
}
