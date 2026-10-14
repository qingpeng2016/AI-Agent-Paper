package topicdiscovery

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

// paperOutputTitleMaxRunes 与 paper_output_literature_review / experiment_plan 的 title VARCHAR(256) 一致（按字符截断）。
const paperOutputTitleMaxRunes = 256

func isContentFormatHint(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "text", "md", "markdown", "plain", "html", "json":
		return true
	default:
		return false
	}
}

// literatureReviewBodyFromPayload 选题 generate 包里的 literature_review → 入库正文。
func literatureReviewBodyFromPayload(lit map[string]any) string {
	if len(lit) == 0 {
		return ""
	}
	if s := strings.TrimSpace(StrFromAny(lit["summary"])); s != "" {
		return s
	}
	if c := strings.TrimSpace(StrFromAny(lit["content"])); c != "" {
		return c
	}
	cm := strings.TrimSpace(StrFromAny(lit["content_medium"]))
	if cm != "" && !isContentFormatHint(cm) {
		return cm
	}
	return ""
}

func literatureReviewReferencesAppend(lit map[string]any, body string) string {
	raw, ok := lit["citations"]
	if !ok || raw == nil {
		return body
	}
	arr, ok := raw.([]any)
	if !ok || len(arr) == 0 {
		return body
	}
	var b strings.Builder
	b.WriteString(strings.TrimSpace(body))
	b.WriteString("\n\n## References\n\n")
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		ref := strings.TrimSpace(StrFromAny(m["ref_id"]))
		note := strings.TrimSpace(StrFromAny(m["note"]))
		if ref == "" && note == "" {
			continue
		}
		if ref != "" && note != "" {
			b.WriteString("- **")
			b.WriteString(ref)
			b.WriteString("**: ")
			b.WriteString(note)
		} else if note != "" {
			b.WriteString("- ")
			b.WriteString(note)
		} else {
			b.WriteString("- ")
			b.WriteString(ref)
		}
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

func experimentPlanTitleFromPayload(plan map[string]any, fallback string) string {
	if t := strings.TrimSpace(StrFromAny(plan["title"])); t != "" {
		return t
	}
	if fallback != "" {
		return fallback + " — 实验方案"
	}
	return "实验方案"
}

func experimentPlanSummaryFromPayload(plan map[string]any) string {
	if s := strings.TrimSpace(StrFromAny(plan["summary"])); s != "" {
		return s
	}
	if sc := strings.TrimSpace(StrFromAny(plan["success_criteria"])); sc != "" {
		return sc
	}
	if f := strings.TrimSpace(StrFromAny(plan["features"])); f != "" {
		if len(f) > 512 {
			return f[:512] + "…"
		}
		return f
	}
	return ""
}

func experimentPlanBodyFromPayload(plan map[string]any) string {
	if c := strings.TrimSpace(StrFromAny(plan["content"])); c != "" {
		return c
	}
	cm := strings.TrimSpace(StrFromAny(plan["content_medium"]))
	if cm != "" && !isContentFormatHint(cm) {
		return cm
	}
	return formatTopicExperimentPlanMarkdown(plan)
}

func formatTopicExperimentPlanMarkdown(plan map[string]any) string {
	var b strings.Builder
	appendPlanSection := func(heading, key string) {
		v, ok := plan[key]
		if !ok || v == nil {
			return
		}
		b.WriteString("## ")
		b.WriteString(heading)
		b.WriteString("\n\n")
		switch t := v.(type) {
		case string:
			b.WriteString(strings.TrimSpace(t))
			b.WriteString("\n\n")
		case []any:
			for _, item := range t {
				line := strings.TrimSpace(fmt.Sprint(item))
				if line == "" {
					continue
				}
				b.WriteString("- ")
				b.WriteString(line)
				b.WriteString("\n")
			}
			b.WriteString("\n")
		default:
			raw, _ := json.MarshalIndent(v, "", "  ")
			b.Write(raw)
			b.WriteString("\n\n")
		}
	}
	appendPlanSection("数据集", "datasets")
	appendPlanSection("特征", "features")
	appendPlanSection("流水线", "pipelines")
	appendPlanSection("预处理", "preprocessing")
	appendPlanSection("评价", "evaluation")
	appendPlanSection("基线", "baselines")
	appendPlanSection("消融", "ablations")
	appendPlanSection("成功标准", "success_criteria")
	out := strings.TrimSpace(b.String())
	if out != "" {
		return out
	}
	raw, _ := json.MarshalIndent(plan, "", "  ")
	return string(raw)
}

func truncateForDBTitle(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if utf8.RuneCountInString(s) <= paperOutputTitleMaxRunes {
		return s
	}
	rs := []rune(s)
	if len(rs) <= paperOutputTitleMaxRunes {
		return s
	}
	trim := paperOutputTitleMaxRunes - 1
	if trim < 1 {
		trim = 1
	}
	return string(rs[:trim]) + "…"
}
