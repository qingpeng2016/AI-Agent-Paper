package paper

import (
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/domain/rest/request"
)

func normalizeTopicDiscoveryInput(req request.TopicDiscoveryRunRequest) (keywords []string, description, direction string, ok bool) {
	keywords = normalizeKeywordList(req.Keywords)
	description = strings.TrimSpace(req.Description)
	legacy := strings.TrimSpace(req.Direction)

	if len(keywords) == 0 && legacy != "" {
		if q := ExtractLiteratureSearchQuery(legacy); q != "" {
			keywords = strings.Fields(q)
		}
		if description == "" {
			description = legacy
		}
	}

	if len(keywords) == 0 || description == "" {
		return nil, "", "", false
	}
	direction = composeTopicDirection(description, keywords)
	return keywords, description, direction, true
}

func normalizeKeywordList(raw []string) []string {
	if len(raw) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	for _, item := range raw {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		for _, part := range splitKeywordTokens(item) {
			key := strings.ToLower(part)
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, part)
		}
	}
	return out
}

func splitKeywordTokens(s string) []string {
	s = strings.NewReplacer("，", ",", "；", ";", "、", ",").Replace(s)
	var parts []string
	for _, chunk := range strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n'
	}) {
		if t := strings.TrimSpace(chunk); t != "" {
			parts = append(parts, t)
		}
	}
	return parts
}

func composeTopicDirection(description string, keywords []string) string {
	description = strings.TrimSpace(description)
	if len(keywords) == 0 {
		return description
	}
	kwLine := strings.Join(keywords, ", ")
	if description == "" {
		return "关键词：" + kwLine
	}
	return "研究内容：" + description + "\n关键词：" + kwLine
}

func literatureSearchQueryFromInput(input topicRunInput) string {
	if len(input.Keywords) > 0 {
		return truncateLiteratureQuery(strings.Join(input.Keywords, " "), 400)
	}
	if q := ExtractLiteratureSearchQuery(input.Direction); q != "" {
		return q
	}
	return truncateLiteratureQuery(strings.TrimSpace(input.Direction), 400)
}
