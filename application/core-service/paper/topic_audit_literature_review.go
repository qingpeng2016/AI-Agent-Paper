package paper

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"gorm.io/datatypes"
)

// stageBeginAuditAsync 仅标记 running，由 bot 异步调用 LLM（审查 + 文献综述）。
func (s *TopicDiscoveryRunService) stageBeginAuditAsync(
	ctx context.Context,
	userID uint,
	runVersion int,
	step *entity.PaperOutputTopicStep,
) error {
	ideas, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "generate_ideas")
	if err != nil || ideas == nil || strings.TrimSpace(ideas.Status) != "completed" {
		return fmt.Errorf("generate_ideas 未完成，无法开始审查与文献综述")
	}
	if len(step.Result) > 0 && strings.TrimSpace(step.Status) == "completed" {
		return nil
	}
	sum := "审查结论与文献综述生成中…"
	step.SummaryText = &sum
	step.Status = "running"
	step.CompletedAt = nil
	meta := LoadStepExtra(step)
	meta["async_llm"] = true
	step.Extra = mustJSON(meta)
	return s.steps.SaveStep(ctx, step)
}

// RunAuditAndLiteratureReviewLLM bot：审计 + 写入 paper_output_literature_review。
func (s *TopicDiscoveryRunService) RunAuditAndLiteratureReviewLLM(
	ctx context.Context,
	userID uint,
	runVersion int,
	auditStep *entity.PaperOutputTopicStep,
	input topicRunInput,
) error {
	ideas, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "generate_ideas")
	if err != nil || ideas == nil || strings.TrimSpace(ideas.Status) != "completed" {
		return fmt.Errorf("generate_ideas 未完成")
	}

	rounds := input.AuditRounds
	if rounds < 1 {
		rounds = 1
	}
	var roundOutputs []map[string]any
	var lastText string
	var totalUsage LLMUsage
	var litReviewPayload map[string]any

	for r := 1; r <= rounds; r++ {
		locale := s.contentLocaleForStep(ctx, auditStep)
		userMsg := s.buildAuditUserExtra(ctx, userID, runVersion, input, r, rounds, locale)
		text, usage, err := s.callStageLLM(ctx, auditStep, "audit", map[string]string{
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
		auditRound, litPart := parseAuditLiteratureReviewLLMResponse(text)
		if len(litPart) > 0 {
			litReviewPayload = litPart
		}
		roundOutputs = append(roundOutputs, auditRound)
	}

	result := map[string]any{"rounds": roundOutputs}
	if litReviewPayload != nil {
		result["literature_review"] = litReviewPayload
	}

	var litReviewID uint64
	if s.litReviews != nil && litReviewPayload != nil {
		id, saveErr := s.persistLiteratureReviewFromLLM(ctx, auditStep, input, userID, runVersion, litReviewPayload)
		if saveErr != nil {
			return saveErr
		}
		litReviewID = id
		if litReviewID > 0 {
			result["literature_review_id"] = litReviewID
		}
	}

	auditStep.Result = mustJSON(result)
	if summary := auditSummaryFromRounds(roundOutputs, litReviewPayload); summary != "" {
		auditStep.SummaryText = ptrString(summary)
	} else {
		auditStep.SummaryText = ptrString(lastText)
	}
	appendUsageMeta(auditStep, totalUsage)
	meta := LoadStepExtra(auditStep)
	meta["includes_literature_review"] = litReviewPayload != nil
	if litReviewID > 0 {
		meta["literature_review_id"] = litReviewID
	}
	delete(meta, "async_llm")
	meta["async_llm_done"] = true
	auditStep.Extra = mustJSON(meta)

	if auditStep.ManuscriptID > 0 {
		if err := s.manuscripts.SetCurrentForUser(ctx, userID, uint(auditStep.ManuscriptID)); err != nil {
			return err
		}
	}

	done := time.Now()
	auditStep.Status = "completed"
	auditStep.CompletedAt = &done
	return s.steps.SaveStep(ctx, auditStep)
}

func parseAuditLiteratureReviewLLMResponse(text string) (auditRound map[string]any, literatureReview map[string]any) {
	text = strings.TrimSpace(text)
	if idx := strings.Index(text, "{"); idx >= 0 {
		if end := strings.LastIndex(text, "}"); end > idx {
			text = text[idx : end+1]
		}
	}
	var root map[string]any
	if err := json.Unmarshal([]byte(text), &root); err != nil {
		return map[string]any{"raw": text}, nil
	}
	if lr, ok := root["literature_review"].(map[string]any); ok {
		literatureReview = lr
	}
	if aud, ok := root["audit"].(map[string]any); ok {
		auditRound = aud
		return auditRound, literatureReview
	}
	// 兼容旧 schema：issues/summary/off_topic 在根上
	auditRound = map[string]any{}
	for _, k := range []string{"issues", "summary", "off_topic"} {
		if v, ok := root[k]; ok {
			auditRound[k] = v
		}
	}
	if len(auditRound) == 0 {
		auditRound = root
	}
	return auditRound, literatureReview
}

func auditSummaryFromRounds(rounds []map[string]any, lit map[string]any) string {
	if len(rounds) > 0 {
		last := rounds[len(rounds)-1]
		if sum, ok := last["summary"].(string); ok && strings.TrimSpace(sum) != "" {
			s := strings.TrimSpace(sum)
			if lit != nil {
				if t, ok := lit["title"].(string); ok && strings.TrimSpace(t) != "" {
					return s + "\n文献综述：" + strings.TrimSpace(t)
				}
			}
			return s
		}
	}
	if lit != nil {
		if sum, ok := lit["summary"].(string); ok && strings.TrimSpace(sum) != "" {
			return strings.TrimSpace(sum)
		}
		if t, ok := lit["title"].(string); ok && strings.TrimSpace(t) != "" {
			return "文献综述：" + strings.TrimSpace(t)
		}
	}
	return ""
}

func (s *TopicDiscoveryRunService) persistLiteratureReviewFromLLM(
	ctx context.Context,
	auditStep *entity.PaperOutputTopicStep,
	input topicRunInput,
	userID uint,
	runVersion int,
	lit map[string]any,
) (uint64, error) {
	if s.litReviews == nil {
		return 0, nil
	}
	msID, err := s.manuscriptIDForRun(ctx, userID, runVersion, auditStep)
	if err != nil {
		return 0, err
	}
	if msID == 0 {
		return 0, fmt.Errorf("无法确定 manuscript_id，文献综述未写入")
	}
	auditStep.ManuscriptID = msID
	title := strFromAny(lit["title"])
	summary := strFromAny(lit["summary"])
	content := strFromAny(lit["content_medium"])
	if content == "" {
		content = strFromAny(lit["content"])
	}
	structure := strFromAny(lit["structure"])
	if structure == "" {
		structure = "thematic"
	}
	var citations datatypes.JSON
	if raw, ok := lit["citations"]; ok && raw != nil {
		b, _ := json.Marshal(raw)
		citations = b
	}
	inputSnap, _ := json.Marshal(map[string]any{
		"run_version":  runVersion,
		"direction":    input.Direction,
		"venue":        input.Venue,
		"audit_level":  input.AuditLevel,
		"source_codes": input.SourceCodes,
	})
	metaSnap, _ := json.Marshal(map[string]any{
		"topic_step_id": auditStep.ID,
		"stage_code":    "audit",
	})

	row := &entity.PaperOutputLiteratureReview{
		ManuscriptID: msID,
		UserID:       uint64(userID),
		Status:       "completed",
		Format:       "md",
		Citations:    citations,
		InputParams:  inputSnap,
		Meta:         metaSnap,
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
	if structure != "" {
		row.Structure = &structure
	}
	if err := s.litReviews.Create(ctx, row); err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (s *TopicDiscoveryRunService) manuscriptIDForRun(
	ctx context.Context,
	userID uint,
	runVersion int,
	auditStep *entity.PaperOutputTopicStep,
) (uint64, error) {
	if auditStep.ManuscriptID > 0 {
		return auditStep.ManuscriptID, nil
	}
	rows, err := s.steps.ListByUserRun(ctx, uint64(userID), runVersion)
	if err == nil {
		if id := manuscriptIDOf(rows); id > 0 {
			return id, nil
		}
	}
	return 0, fmt.Errorf("manuscript_id 未绑定，请在开始选题前创建论文")
}

func strFromAny(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case float32:
		return strconv.FormatInt(int64(t), 10)
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case int32:
		return strconv.FormatInt(int64(t), 10)
	case uint64:
		return strconv.FormatUint(t, 10)
	case uint:
		return strconv.FormatUint(uint64(t), 10)
	case json.Number:
		return strings.TrimSpace(t.String())
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}
