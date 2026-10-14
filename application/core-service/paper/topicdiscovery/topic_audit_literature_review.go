package topicdiscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

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
	title := StrFromAny(lit["title"])
	content := StrFromAny(lit["content_medium"])
	if content == "" {
		content = StrFromAny(lit["content"])
	}

	row := &entity.PaperOutputLiteratureReview{
		ManuscriptID: msID,
		UserID:       uint64(userID),
		Status:       "completed",
		From:         "generated",
		Format:       "md",
	}
	if title != "" {
		row.Title = &title
	}
	if content != "" {
		row.ContentMedium = &content
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

func StrFromAny(v any) string {
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
