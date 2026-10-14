package topicdiscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/internal/jsonx"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/lang"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/llm"
	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
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
	sum := "生成文献综述、实验方案中…"
	step.SummaryText = &sum
	step.Status = "running"
	step.CompletedAt = nil
	step.ExecProgressPct = 0
	meta := paperstep.LoadStepExtra(step)
	meta["async_llm"] = true
	delete(meta, "exec_progress_completed_calls")
	delete(meta, "exec_progress_planned_calls")
	step.Extra = jsonx.MustJSON(meta)
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

	genRoute, err := s.requireTopicStrategyRoute(ctx, BehaviorTopicGenerate)
	if err != nil {
		return err
	}
	auditRoute, err := s.requireTopicStrategyRoute(ctx, BehaviorTopicAudit)
	if err != nil {
		return err
	}

	ctxBlock, _ := s.stageContextCompact(ctx, userID, runVersion, "retrieve")
	filesBlock := s.formatRetrieveCorpusFilesBlock(ctx, userID, runVersion)
	locale := s.contentLocaleForStep(ctx, step)

	var (
		prevBundle     map[string]any
		prevAudit      map[string]any
		prevBundleRaw  json.RawMessage
		prevAuditRaw   json.RawMessage
		iterations     []map[string]any
		totalUsage     llm.LLMUsage
		finalBundle    map[string]any
		finalAudit     map[string]any
		passed         bool
	)
	plannedCalls := topicDiscoveryPlannedCalls(maxRounds)
	completedCalls := 0

	for round := 1; round <= maxRounds; round++ {
		invokeCtx := s.buildTopicStrategyInvokeContext(
			ctx, userID, runVersion, step, input, round, maxRounds, prevBundleRaw, prevAuditRaw,
		)
		genUser := buildTopicGenerateBundleUserMessage(input, ctxBlock, filesBlock, locale, round, maxRounds, prevBundle, prevAudit)
		genReq := BuildTopicDiscoveryInvokeRequest(genRoute.BehaviorKey, genRoute.CapabilityID, input, invokeCtx)
		if err := s.saveTopicDiscoveryExecProgress(ctx, step, completedCalls, plannedCalls, completedCalls == 0); err != nil {
			return err
		}
		bundleRaw, genUsage, err := s.invokeTopicDiscoveryBehavior(ctx, step, genReq, genUser)
		if err != nil {
			return err
		}
		completedCalls++
		if err := s.saveTopicDiscoveryExecProgress(ctx, step, completedCalls, plannedCalls, false); err != nil {
			return err
		}
		totalUsage.PromptTokens += genUsage.PromptTokens
		totalUsage.CompletionTokens += genUsage.CompletionTokens

		bundle := map[string]any{}
		if err := json.Unmarshal(bundleRaw, &bundle); err != nil || len(bundle) == 0 {
			bundle = parseTopicGenerateBundleResponse(string(bundleRaw))
		}
		bundleJSON := bundleRaw

		invokeCtx.PrevBundle = bundleJSON
		auditUser := s.buildTopicAuditOnlyUserExtra(ctx, userID, runVersion, input, round, maxRounds, locale, string(bundleJSON))
		auditReq := BuildTopicDiscoveryInvokeRequest(auditRoute.BehaviorKey, auditRoute.CapabilityID, input, invokeCtx)
		if err := s.saveTopicDiscoveryExecProgress(ctx, step, completedCalls, plannedCalls, false); err != nil {
			return err
		}
		auditRaw, auditUsage, err := s.invokeTopicDiscoveryBehavior(ctx, step, auditReq, auditUser)
		if err != nil {
			return err
		}
		completedCalls++
		if err := s.saveTopicDiscoveryExecProgress(ctx, step, completedCalls, plannedCalls, false); err != nil {
			return err
		}
		totalUsage.PromptTokens += auditUsage.PromptTokens
		totalUsage.CompletionTokens += auditUsage.CompletionTokens

		auditRoot := map[string]any{}
		if err := json.Unmarshal(auditRaw, &auditRoot); err != nil || len(auditRoot) == 0 {
			auditRoot = parseTopicAuditOnlyResponse(string(auditRaw))
		}
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
		prevBundleRaw = bundleJSON
		if auditBlock != nil {
			prevAuditRaw = json.RawMessage(jsonx.MustJSON(auditBlock))
		}

		if passed {
			break
		}
	}

	litPayload, _ := finalBundle["literature_review"].(map[string]any)
	if litPayload == nil {
		litPayload = map[string]any{}
	}
	planPayload, _ := finalBundle["experiment_plan"].(map[string]any)

	result := map[string]any{
		"passed":            passed,
		"iterations":        iterations,
		"ideas":             finalBundle["ideas"],
		"novelty":           finalBundle["novelty"],
		"literature_review": litPayload,
		"experiment_plan":   planPayload,
		"final_audit":       finalAudit,
		"exec_rounds":       maxRounds,
	}

	step.Result = jsonx.MustJSON(result)
	summary := topicDiscoverySummaryFromResult(passed, finalAudit, litPayload, len(iterations))
	step.SummaryText = jsonx.PtrString(summary)
	appendUsageMeta(step, totalUsage)
	meta := paperstep.LoadStepExtra(step)
	meta["strategy_routes"] = map[string]any{
		"topic_generate": strategyRouteMeta(genRoute),
		"topic_audit":    strategyRouteMeta(auditRoute),
	}
	meta["revision_rounds"] = len(iterations)
	meta["gate_passed"] = passed
	meta["awaiting_user_confirm"] = true
	delete(meta, "async_llm")
	meta["async_llm_done"] = true
	meta["exec_progress_completed_calls"] = completedCalls
	meta["exec_progress_planned_calls"] = plannedCalls
	step.ExecProgressPct = 100
	step.Extra = jsonx.MustJSON(meta)

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
		sev := strings.ToLower(strings.TrimSpace(StrFromAny(m["gate"])))
		if sev == "" {
			sev = strings.ToLower(strings.TrimSpace(StrFromAny(m["severity"])))
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
	if sum := StrFromAny(audit["summary"]); sum != "" {
		parts = append(parts, sum)
	}
	if t := StrFromAny(lit["title"]); t != "" {
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
	title := StrFromAny(plan["title"])
	summary := StrFromAny(plan["summary"])
	content := StrFromAny(plan["content_medium"])
	if content == "" {
		content = StrFromAny(plan["content"])
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
	locale = lang.NormalizeContentLanguage(locale)
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
	base := lang.BuildIdeasNoveltyUserMessage(input.Direction, input.Venue, input.MaxIdeas, ctxBlock, filesBlock, locale)
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
