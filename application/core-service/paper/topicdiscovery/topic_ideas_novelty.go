package topicdiscovery

import (
	"context"
	"encoding/json"
	"strings"
)

func parseIdeasNoveltyLLMResponse(text string) (ideas map[string]any, novelty map[string]any, summary string) {
	text = strings.TrimSpace(text)
	if idx := strings.Index(text, "{"); idx >= 0 {
		if end := strings.LastIndex(text, "}"); end > idx {
			text = text[idx : end+1]
		}
	}
	var root map[string]any
	if err := json.Unmarshal([]byte(text), &root); err != nil {
		return map[string]any{"raw": text}, map[string]any{"lines": []string{text}}, text
	}
	if sum, ok := root["summary"].(string); ok {
		summary = strings.TrimSpace(sum)
	}
	ideas = map[string]any{}
	if raw, ok := root["ideas"]; ok {
		ideas["ideas"] = raw
	} else {
		ideas["ideas"] = []any{}
	}
	novelty = map[string]any{}
	if nov, ok := root["novelty"].(map[string]any); ok {
		novelty = nov
	} else {
		if lines, ok := root["lines"]; ok {
			novelty["lines"] = lines
		}
		if risks, ok := root["risks"]; ok {
			novelty["risks"] = risks
		}
		if syn, ok := root["synthesis"].(string); ok {
			novelty["synthesis"] = syn
		}
	}
	if summary == "" {
		summary = noveltySummaryFromResult(novelty)
	}
	return ideas, novelty, summary
}

func noveltySummaryFromResult(novelty map[string]any) string {
	if novelty == nil {
		return ""
	}
	if syn, ok := novelty["synthesis"].(string); ok && strings.TrimSpace(syn) != "" {
		return strings.TrimSpace(syn)
	}
	if raw, ok := novelty["lines"].([]any); ok && len(raw) > 0 {
		var lines []string
		for _, item := range raw {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				lines = append(lines, strings.TrimSpace(s))
			}
		}
		if len(lines) > 0 {
			return strings.Join(lines, "\n")
		}
	}
	return ""
}

func (s *TopicDiscoveryRunService) noveltyBlockForAudit(ctx context.Context, userID uint, runVersion int) (noveltyJSON, noveltySummary string) {
	ideasRaw := strings.TrimSpace(s.stepResultText(ctx, userID, runVersion, "topic_discovery"))
	if ideasRaw == "" {
		ideasRaw = strings.TrimSpace(s.stepResultText(ctx, userID, runVersion, "generate_ideas"))
	}
	if ideasRaw != "" {
		var root map[string]any
		if err := json.Unmarshal([]byte(ideasRaw), &root); err == nil {
			if nov, ok := root["novelty"]; ok && nov != nil {
				if b, err := json.Marshal(nov); err == nil {
					noveltyJSON = string(b)
				}
				if mm, ok := nov.(map[string]any); ok {
					noveltySummary = noveltySummaryFromResult(mm)
				}
			}
		}
	}
	if strings.TrimSpace(noveltyJSON) == "" {
		noveltyJSON = s.stepResultText(ctx, userID, runVersion, "novelty")
		noveltySummary = s.stepSummaryText(ctx, userID, runVersion, "novelty")
	}
	return noveltyJSON, noveltySummary
}
