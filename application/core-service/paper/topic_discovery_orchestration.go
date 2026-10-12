package paper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"gorm.io/datatypes"
)

const (
	topicLLMStageGenerate = "topic_generate"
	topicLLMStageAudit    = "topic_audit"
)

func (s *TopicDiscoveryRunService) stageBeginTopicDiscoveryAsync(
	ctx context.Context,
	userID uint,
	runVersion int,
	step *entity.PaperOutputTopicStep,
) error {
	ret, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "retrieve")
	if err != nil || ret == nil || strings.TrimSpace(ret.Status) != "completed" {
		return fmt.Errorf("retrieve 未完成，无法开始选题发现")
	}
	if len(step.Result) > 0 && strings.TrimSpace(step.Status) == "completed" {
		return nil
	}
	sum := "选题发现：生成与门禁审计迭代中…"
	step.SummaryText = &sum
	step.Status = "running"
	step.CompletedAt = nil
	meta := LoadStepExtra(step)
	meta["async_llm"] = true
	step.Extra = mustJSON(meta)
	return s.steps.SaveStep(ctx, step)
}

// RunTopicDiscoveryLLM bot：生成包 ↔ 审计包迭代，直至通过或达到 exec_rounds。
func (s *TopicDiscoveryRunService) RunTopicDiscoveryLLM(
	ctx context.Context,
	userID uint,
	runVersion int,
	step *entity.PaperOutputTopicStep,
	input topicRunInput,
) error {
	ret, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "retrieve")
	if err != nil || ret == nil || strings.TrimSpace(ret.Status) != "completed" {
		return fmt.Errorf("retrieve 未完成")
	}

	maxRounds := input.ExecRounds
	if maxRounds < 1 {
		maxRounds = 1
	}

	ctxBlock, _ := s.stageContextCompact(ctx, userID, runVersion, "retrieve")
	filesBlock := s.formatRetrieveCorpusFilesBlock(ctx, userID, runVersion)
	locale := s.contentLocaleForStep(ctx, step)

	var (
		prevBundle     map[string]any
		prevAudit      map[string]any
		iterations     []map[string]any
		totalUsage     LLMUsage
		finalBundle    map[string]any
		finalAudit     map[string]any
		passed         bool
	)

	for round := 1; round <= maxRounds; round++ {
		genUser := buildTopicGenerateBundleUserMessage(input, ctxBlock, filesBlock, locale, round, maxRounds, prevBundle, prevAudit)
		genText, genUsage, err := s.callStageLLM(ctx, step, topicLLMStageGenerate, map[string]string{
			"direction": input.Direction,
			"venue":     input.Venue,
			"max_ideas": fmt.Sprintf("%d", input.MaxIdeas),
		}, genUser)
		if err != nil {
			return err
		}
		totalUsage.PromptTokens += genUsage.PromptTokens
		totalUsage.CompletionTokens += genUsage.CompletionTokens

		bundle := parseTopicGenerateBundleResponse(genText)
		bundleJSON, _ := json.Marshal(bundle)

		auditUser := s.buildTopicAuditOnlyUserExtra(ctx, userID, runVersion, input, round, maxRounds, locale, string(bundleJSON))
		auditText, auditUsage, err := s.callStageLLM(ctx, step, topicLLMStageAudit, map[string]string{
			"direction": compactDirectionForPrompt(input),
			"venue":     input.Venue,
		}, auditUser)
		if err != nil {
			return err
		}
		totalUsage.PromptTokens += auditUsage.PromptTokens
		totalUsage.CompletionTokens += auditUsage.CompletionTokens

		auditRoot := parseTopicAuditOnlyResponse(auditText)
		auditBlock, _ := auditRoot["audit"].(map[string]any)
		if auditBlock == nil {
			auditBlock = auditRoot
		}
		roundPassed := topicAuditPassed(auditBlock)

		iterations = append(iterations, map[string]any{
			"round":    round,
			"generate": bundle,
			"audit":    auditBlock,
			"passed":   roundPassed,
		})

		finalBundle = bundle
		finalAudit = auditBlock
		passed = roundPassed
		prevBundle = bundle
		prevAudit = auditBlock

		if passed {
			break
		}
	}

	litPayload, _ := finalBundle["literature_review"].(map[string]any)
	if litPayload == nil {
		litPayload = map[string]any{}
	}
	planPayload, _ := finalBundle["experiment_plan"].(map[string]any)

	var litReviewID uint64
	if s.litReviews != nil && len(litPayload) > 0 {
		id, saveErr := s.persistLiteratureReviewFromLLM(ctx, step, input, userID, runVersion, litPayload)
		if saveErr != nil {
			return saveErr
		}
		litReviewID = id
	}

	var planID uint64
	if s.expPlans != nil && len(planPayload) > 0 {
		planRow, planErr := s.persistTopicExperimentPlan(ctx, userID, step, input, litReviewID, planPayload)
		if planErr != nil {
			return planErr
		}
		if planRow != nil {
			planID = planRow.ID
		}
	}
	if litReviewID > 0 && planID > 0 && s.litReviews != nil {
		_ = s.litReviews.LinkExperimentPlan(ctx, litReviewID, planID, uint64(userID), step.ManuscriptID)
	}

	result := map[string]any{
		"passed":              passed,
		"iterations":          iterations,
		"ideas":               finalBundle["ideas"],
		"novelty":             finalBundle["novelty"],
		"literature_review":   litPayload,
		"experiment_plan":     planPayload,
		"final_audit":         finalAudit,
	}
	if litReviewID > 0 {
		result["literature_review_id"] = litReviewID
	}
	if planID > 0 {
		result["experiment_plan_id"] = planID
	}

	step.Result = mustJSON(result)
	summary := topicDiscoverySummaryFromResult(passed, finalAudit, litPayload, len(iterations))
	step.SummaryText = ptrString(summary)
	appendUsageMeta(step, totalUsage)
	meta := LoadStepExtra(step)
	meta["revision_rounds"] = len(iterations)
	meta["gate_passed"] = passed
	if litReviewID > 0 {
		meta["literature_review_id"] = litReviewID
	}
	if planID > 0 {
		meta["experiment_plan_id"] = planID
	}
	delete(meta, "async_llm")
	meta["async_llm_done"] = true
	step.Extra = mustJSON(meta)

	if step.ManuscriptID > 0 {
		if err := s.manuscripts.SetCurrentForUser(ctx, userID, uint(step.ManuscriptID)); err != nil {
			return err
		}
	}

	done := time.Now()
	step.Status = "completed"
	step.CompletedAt = &done
	return s.steps.SaveStep(ctx, step)
}

func topicAuditPassed(audit map[string]any) bool {
	if audit == nil {
		return false
	}
	if p, ok := audit["passed"].(bool); ok {
		return p
	}
	issues, _ := audit["issues"].([]any)
	for _, iss := range issues {
		m, ok := iss.(map[string]any)
		if !ok {
			continue
		}
		sev := strings.ToLower(strings.TrimSpace(strFromAny(m["gate"])))
		if sev == "" {
			sev = strings.ToLower(strings.TrimSpace(strFromAny(m["severity"])))
		}
		if sev == "blocker" {
			return false
		}
	}
	return true
}

func parseTopicGenerateBundleResponse(text string) map[string]any {
	text = strings.TrimSpace(text)
	if idx := strings.Index(text, "{"); idx >= 0 {
		if end := strings.LastIndex(text, "}"); end > idx {
			text = text[idx : end+1]
		}
	}
	var root map[string]any
	if err := json.Unmarshal([]byte(text), &root); err != nil {
		return map[string]any{"raw": text}
	}
	out := map[string]any{}
	for _, k := range []string{"ideas", "novelty", "literature_review", "experiment_plan"} {
		if v, ok := root[k]; ok && v != nil {
			out[k] = v
		}
	}
	if len(out) == 0 {
		ideas, novelty, _ := parseIdeasNoveltyLLMResponse(text)
		for k, v := range ideas {
			out[k] = v
		}
		if len(novelty) > 0 {
			out["novelty"] = novelty
		}
	}
	return out
}

func parseTopicAuditOnlyResponse(text string) map[string]any {
	text = strings.TrimSpace(text)
	if idx := strings.Index(text, "{"); idx >= 0 {
		if end := strings.LastIndex(text, "}"); end > idx {
			text = text[idx : end+1]
		}
	}
	var root map[string]any
	if err := json.Unmarshal([]byte(text), &root); err != nil {
		return map[string]any{"audit": map[string]any{"raw": text, "passed": false}}
	}
	if _, ok := root["audit"]; ok {
		return root
	}
	return map[string]any{"audit": root}
}

func topicDiscoverySummaryFromResult(passed bool, audit, lit map[string]any, rounds int) string {
	var parts []string
	if passed {
		parts = append(parts, fmt.Sprintf("选题发现通过（%d 轮）", rounds))
	} else {
		parts = append(parts, fmt.Sprintf("选题发现未完全通过门禁（%d 轮后保留最优稿）", rounds))
	}
	if sum := strFromAny(audit["summary"]); sum != "" {
		parts = append(parts, sum)
	}
	if t := strFromAny(lit["title"]); t != "" {
		parts = append(parts, "文献综述："+t)
	}
	return strings.Join(parts, "\n")
}

func (s *TopicDiscoveryRunService) persistTopicExperimentPlan(
	ctx context.Context,
	userID uint,
	step *entity.PaperOutputTopicStep,
	input topicRunInput,
	litReviewID uint64,
	plan map[string]any,
) (*entity.PaperOutputExperimentPlan, error) {
	if s.expPlans == nil || len(plan) == 0 {
		return nil, nil
	}
	msID := step.ManuscriptID
	if msID == 0 {
		return nil, fmt.Errorf("manuscript_id 未绑定")
	}
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
		"literature_review_id": litReviewID,
		"direction":            input.Direction,
		"venue":                input.Venue,
		"stage_code":           "topic_discovery",
		"topic_step_id":        step.ID,
	})
	row := &entity.PaperOutputExperimentPlan{
		ManuscriptID: msID,
		UserID:       uint64(userID),
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
	if err := s.expPlans.Create(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func buildTopicGenerateBundleUserMessage(
	input topicRunInput,
	ctxBlock, filesBlock, locale string,
	round, maxRounds int,
	prevBundle, prevAudit map[string]any,
) string {
	locale = NormalizeContentLanguage(locale)
	reviseBlock := ""
	if round > 1 && (prevBundle != nil || prevAudit != nil) {
		prevB, _ := json.Marshal(prevBundle)
		prevA, _ := json.Marshal(prevAudit)
		if locale == "en" {
			reviseBlock = fmt.Sprintf(`
[Revision round %d/%d]
Revise the previous bundle minimally to address the audit_report. Do not invent new ref_id outside Corpus.
[Previous generate bundle]
%s
[Previous audit_report]
%s`, round, maxRounds, string(prevB), string(prevA))
		} else {
			reviseBlock = fmt.Sprintf(`
【修订轮次 %d/%d】
在保持研究方向前提下，根据 audit_report 最小改动上一版生成包；禁止编造 Corpus 外的新 ref_id。
【上一版生成包】
%s
【上一版 audit_report】
%s`, round, maxRounds, string(prevB), string(prevA))
		}
	}
	base := buildIdeasNoveltyUserMessage(input, ctxBlock, filesBlock, locale)
	schemaExtra := `
{
  "ideas": [{"title":"","problem":"","approach":"","contribution":"","reference_keys":[]}],
  "novelty": {"lines":[],"risks":[],"synthesis":""},
  "literature_review": {
    "title":"","summary":"","structure":"thematic","content_medium":"","citations":[{"ref_id":"","note":""}]
  },
  "experiment_plan": null
}`
	if locale == "en" {
		return base + reviseBlock + `

Also output literature_review and experiment_plan in the SAME JSON. experiment_plan may be null if the venue/direction needs no experiments.
Full schema:` + schemaExtra
	}
	return base + reviseBlock + `

同一次 JSON 还须输出 literature_review；experiment_plan 可为 null（无需实验的论文类型须显式 null）。
完整 schema：` + schemaExtra
}
