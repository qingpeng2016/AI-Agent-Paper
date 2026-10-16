package arxiv

import (
	"net/url"
	"strconv"
	"strings"
)

// buildArxivSearchQueryParam 构造 search_query 原文（交给 url.Values 编码）。
// 勿使用 all:<整句带空格>：API 会把空格拆成 OR，命中百万级无关结果。
func buildArxivSearchQueryParam(userQuery string) string {
	tokens := strings.Fields(strings.TrimSpace(userQuery))
	if len(tokens) == 0 {
		return "all:*"
	}
	if len(tokens) == 1 {
		return "all:" + arxivTerm(tokens[0])
	}
	parts := make([]string, len(tokens))
	for i, t := range tokens {
		parts[i] = "all:" + arxivTerm(t)
	}
	return strings.Join(parts, " AND ")
}

func arxivTerm(t string) string {
	t = strings.TrimSpace(t)
	if t == "" {
		return "*"
	}
	if strings.Contains(t, " ") {
		return `"` + strings.ReplaceAll(t, `"`, "") + `"`
	}
	return t
}

func arxivAPIURL(baseURL, userQuery string, start, maxResults int) string {
	v := url.Values{}
	v.Set("search_query", buildArxivSearchQueryParam(userQuery))
	v.Set("start", strconv.Itoa(start))
	v.Set("max_results", strconv.Itoa(maxResults))
	return strings.TrimRight(baseURL, "/") + "?" + v.Encode()
}
