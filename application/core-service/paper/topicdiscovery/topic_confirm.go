package topicdiscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/internal/jsonx"
	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
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

	discStep, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "topic_discovery")
	if err != nil || discStep == nil {
		return nil, errorx.ErrTopicStepInvalid
	}
	if strings.TrimSpace(discStep.Status) != "completed" {
		return nil, errorx.ErrTopicStepInvalid.WithDetail("选题发现尚未完成，无法确认轮次")
	}

	confirmStep, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "topic_confirm")
	if err != nil || confirmStep == nil {
		return nil, errorx.ErrTopicStepInvalid.WithDetail("缺少确认步骤（请使用新版选题 run）")
	}
	if strings.TrimSpace(confirmStep.Status) == "completed" {
		return s.reloadView(ctx, userID, runVersion)
	}

	selectedRound := req.SelectedRound
	if selectedRound < 1 {
		return nil, errorx.ErrParamsError.WithDetail("请选择有效轮次")
	}

	var discResult map[string]any
	if err := json.Unmarshal(discStep.Result, &discResult); err != nil || len(discResult) == 0 {
		return nil, errorx.ErrTopicStepInvalid.WithDetail("选题发现结果为空")
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

	var litReviewID uint64
	if s.litReviews != nil && len(litPayload) > 0 {
		id, saveErr := s.persistLiteratureReviewFromLLM(ctx, discStep, input, userID, runVersion, litPayload)
		if saveErr != nil {
			return nil, saveErr
		}
		litReviewID = id
	}

	var planID uint64
	if s.expPlans != nil && planPayload != nil && len(planPayload) > 0 {
		planRow, planErr := s.persistTopicExperimentPlan(ctx, userID, discStep, input, litReviewID, planPayload)
		if planErr != nil {
			return nil, planErr
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
	sum := topicConfirmSummary(selectedRound, passed, auditBlock, litPayload)
	confirmStep.SummaryText = jsonx.PtrString(sum)
	meta := paperstep.LoadStepExtra(confirmStep)
	meta["selected_round"] = selectedRound
	if litReviewID > 0 {
		meta["literature_review_id"] = litReviewID
	}
	if planID > 0 {
		meta["experiment_plan_id"] = planID
	}
	confirmStep.Extra = jsonx.MustJSON(meta)

	if err := s.steps.SaveStep(ctx, confirmStep); err != nil {
		return nil, err
	}

	if discStep.ManuscriptID > 0 {
		if err := s.manuscripts.SetCurrentForUser(ctx, userID, uint(discStep.ManuscriptID)); err != nil {
			return nil, err
		}
	}

	return s.reloadView(ctx, userID, runVersion)
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

func topicConfirmSummary(round int, passed bool, audit, lit map[string]any) string {
	var parts []string
	if passed {
		parts = append(parts, fmt.Sprintf("已确认第 %d 轮（门禁通过）", round))
	} else {
		parts = append(parts, fmt.Sprintf("已确认第 %d 轮（该轮未完全通过门禁）", round))
	}
	if audit != nil {
		if sum := StrFromAny(audit["summary"]); sum != "" {
			parts = append(parts, sum)
		}
	}
	if lit != nil {
		if t := StrFromAny(lit["title"]); t != "" {
			parts = append(parts, "文献综述："+t)
		}
	}
	return strings.Join(parts, "\n")
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
