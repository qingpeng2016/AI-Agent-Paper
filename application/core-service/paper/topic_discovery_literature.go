package paper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func compactDirectionForPrompt(input topicRunInput) string {
	if strings.TrimSpace(input.Description) != "" {
		return strings.TrimSpace(input.Description)
	}
	if len(input.Keywords) > 0 {
		return strings.Join(input.Keywords, ", ")
	}
	return strings.TrimSpace(input.Direction)
}

func keywordsLine(input topicRunInput) string {
	if len(input.Keywords) > 0 {
		return strings.Join(input.Keywords, ", ")
	}
	return ExtractLiteratureSearchQuery(input.Direction)
}

func buildRetrievePaperCards(hits []storedLiteratureHit) string {
	var b strings.Builder
	for _, h := range hits {
		abstract := ""
		if m, ok := h.Meta.(map[string]any); ok {
			if a, ok := m["abstract"].(string); ok {
				abstract = strings.TrimSpace(a)
			}
		}
		if len([]rune(abstract)) > 400 {
			abstract = string([]rune(abstract)[:400]) + "…"
		}
		fmt.Fprintf(&b, "- ref_id: %s\n  title: %s\n  year: %d\n  doi: %s\n  abstract: %s\n\n",
			h.ExternalKey, h.Title, h.PublishedYear, h.DOI, emptyFallback(abstract, "(无摘要)"))
	}
	return strings.TrimSpace(b.String())
}

func emptyFallback(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

func fallbackLiteratureBrief(hits []storedLiteratureHit) map[string]any {
	papers := make([]map[string]any, 0, len(hits))
	for _, h := range hits {
		papers = append(papers, map[string]any{
			"ref_id":        h.ExternalKey,
			"title":         h.Title,
			"year":          h.PublishedYear,
			"takeaway":      "（规则摘要：请结合标题理解；模型简报未生成）",
			"methods_note":  "",
		})
	}
	synthesis := fmt.Sprintf("共 %d 篇已验真文献；模型未生成 synthesis，请基于标题与摘要字段人工判断。", len(hits))
	return map[string]any{
		"papers":         papers,
		"synthesis":      synthesis,
		"gaps":           []string{},
		"coverage_note":  "自动生成（无 LLM）",
	}
}

func parseLiteratureBriefFromLLM(text string) (map[string]any, string) {
	text = strings.TrimSpace(text)
	if idx := strings.Index(text, "{"); idx >= 0 {
		if end := strings.LastIndex(text, "}"); end > idx {
			text = text[idx : end+1]
		}
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		return nil, text
	}
	if _, ok := m["papers"]; !ok {
		return nil, text
	}
	sum, _ := m["synthesis"].(string)
	if sum == "" {
		sum = text
	}
	return m, sum
}

func mergeLiteratureBriefIntoResult(result map[string]any, brief map[string]any) {
	if brief == nil {
		return
	}
	result["literature_brief"] = brief
}

func allowedRefIDsFromBrief(brief map[string]any) []string {
	raw, _ := brief["papers"].([]any)
	var out []string
	for _, item := range raw {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id, _ := row["ref_id"].(string)
		id = strings.TrimSpace(id)
		if id != "" {
			out = append(out, id)
		}
	}
	return out
}

// formatRetrieveCorpusFilesBlock 列出 retrieve 步已落盘 PDF，供 generate_ideas 与 ref_id 对齐。
func (s *TopicDiscoveryRunService) formatRetrieveCorpusFilesBlock(ctx context.Context, userID uint, runVersion int) string {
	st, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "retrieve")
	if err != nil || st == nil {
		return "(无 CorpusFiles：retrieve 未完成或未落盘 PDF)"
	}
	seen := map[string]struct{}{}
	var lines []string
	for _, f := range LoadStepFiles(st) {
		if f.Kind != stepFileKindLiteraturePDF {
			continue
		}
		key := strings.TrimSpace(f.Path) + "|" + strings.TrimSpace(f.ExternalKey)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		lines = append(lines, fmt.Sprintf("- external_key: %s\n  local_path: %s",
			strings.TrimSpace(f.ExternalKey), strings.TrimSpace(f.Path)))
	}
	if len(lines) == 0 {
		items := ParseLiteratureDownloads(st.Extra)
		for _, it := range items {
			rel := strings.TrimSpace(it.LocalPath)
			if rel == "" {
				continue
			}
			key := rel + "|" + it.ExternalKey
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			lines = append(lines, fmt.Sprintf("- external_key: %s\n  local_path: %s",
				strings.TrimSpace(it.ExternalKey), rel))
		}
	}
	if len(lines) == 0 {
		return "(无 CorpusFiles：尚无 local_path；仅可依据 Corpus 中的标题/摘要)"
	}
	return strings.Join(lines, "\n")
}

func formatLiteratureBriefForPrompt(brief map[string]any) string {
	if brief == nil {
		return "(无文献简报)"
	}
	b, err := json.MarshalIndent(brief, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", brief)
	}
	return string(b)
}

func (s *TopicDiscoveryRunService) literatureBriefForRun(ctx context.Context, userID uint, runVersion int) (map[string]any, []string, error) {
	st, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "retrieve")
	if err != nil || st == nil || len(st.Result) == 0 {
		return nil, nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(st.Result, &result); err != nil {
		return nil, nil, err
	}
	if brief, ok := result["literature_brief"].(map[string]any); ok && brief != nil {
		return brief, allowedRefIDsFromBrief(brief), nil
	}
	// 旧 run：从 hits 即时生成规则简报
	var hits []storedLiteratureHit
	if raw, ok := result["literature_hits"]; ok {
		b, _ := json.Marshal(raw)
		_ = json.Unmarshal(b, &hits)
	}
	if len(hits) == 0 {
		return nil, nil, nil
	}
	brief := fallbackLiteratureBrief(hits)
	return brief, allowedRefIDsFromBrief(brief), nil
}

func (s *TopicDiscoveryRunService) stageContextCompact(ctx context.Context, userID uint, runVersion int, includeStages ...string) (string, error) {
	allow := map[string]struct{}{}
	for _, st := range includeStages {
		allow[st] = struct{}{}
	}
	var parts []string
	for _, stage := range topicDiscoveryStages {
		if len(allow) > 0 {
			if _, ok := allow[stage]; !ok {
				continue
			}
		}
		st, err := s.steps.GetStep(ctx, uint64(userID), runVersion, stage)
		if err != nil || st == nil {
			continue
		}
		if stage == "retrieve" {
			brief, _, _ := s.literatureBriefForRun(ctx, userID, runVersion)
			if brief != nil {
				parts = append(parts, "[literature_brief]\n"+formatLiteratureBriefForPrompt(brief))
				continue
			}
		}
		if len(st.Result) > 0 {
			parts = append(parts, fmt.Sprintf("[%s]\n%s", stage, string(st.Result)))
		}
		if st.SummaryText != nil && strings.TrimSpace(*st.SummaryText) != "" {
			parts = append(parts, *st.SummaryText)
		}
	}
	return strings.Join(parts, "\n\n"), nil
}

func (s *TopicDiscoveryRunService) stepResultText(ctx context.Context, userID uint, runVersion int, stageCode string) string {
	st, err := s.steps.GetStep(ctx, uint64(userID), runVersion, stageCode)
	if err != nil || st == nil || len(st.Result) == 0 {
		return "（本步无 result）"
	}
	return string(st.Result)
}

func (s *TopicDiscoveryRunService) stepSummaryText(ctx context.Context, userID uint, runVersion int, stageCode string) string {
	st, err := s.steps.GetStep(ctx, uint64(userID), runVersion, stageCode)
	if err != nil || st == nil || st.SummaryText == nil {
		return ""
	}
	return strings.TrimSpace(*st.SummaryText)
}

// buildAuditUserExtra 审计步：显式带上关键词/描述、文献、idea、新颖性，并要求评估是否偏题。
func (s *TopicDiscoveryRunService) buildAuditUserExtra(
	ctx context.Context,
	userID uint,
	runVersion int,
	input topicRunInput,
	round, rounds int,
) string {
	brief, refIDs, _ := s.literatureBriefForRun(ctx, userID, runVersion)
	briefText := formatLiteratureBriefForPrompt(brief)
	refsLine := strings.Join(refIDs, ", ")
	if refsLine == "" {
		refsLine = "（无）"
	}
	ideasJSON := s.stepResultText(ctx, userID, runVersion, "generate_ideas")
	noveltyJSON, noveltySummary := s.noveltyBlockForAudit(ctx, userID, runVersion)

	return fmt.Sprintf(`审计轮次 %d/%d
目标期刊：%s
审计档位：%s

【用户关键词】
%s

【研究描述】
%s

【文献简报】（仅可引用下列 ref_id：%s）
%s

【候选 idea · generate_ideas 完整产出】
%s

【新颖性分析 · novelty 完整产出】
%s
%s

【审计要求】
1. 候选 idea 是否与关键词、研究描述一致；是否明显偏题（off-topic）。
2. 新颖性结论是否支撑 idea，有无与文献/方向矛盾或过度声称。
3. 以严格审稿人视角检查：论断是否有文献支持、是否缺少基线/对照、贡献是否清晰；并列出 unsupported claims、缺失基线/对照、贡献含糊等问题。

输出 ONLY valid JSON：
{"issues":[{"severity":"blocker|major|minor","claim":"","fix":""}],"summary":"","off_topic":{"ideas_aligned":true,"novelty_aligned":true,"notes":""}}
off_topic 与 issues 字段使用中文。`,
		round, rounds,
		input.Venue, input.AuditLevel,
		keywordsLine(input),
		compactDirectionForPrompt(input),
		refsLine,
		briefText,
		ideasJSON,
		noveltyJSON,
		func() string {
			if noveltySummary == "" {
				return ""
			}
			return "\n【新颖性摘要 summary_text】\n" + noveltySummary
		}(),
	)
}
