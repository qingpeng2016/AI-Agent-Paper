package paper

import (
	"context"
	"fmt"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
)

// LLMDescForLocale 按论文语言取 description 正文。
func LLMDescForLocale(locale string, llmDescZh, llmDescEn string) string {
	if NormalizeContentLanguage(locale) == "en" {
		return strings.TrimSpace(llmDescEn)
	}
	return strings.TrimSpace(llmDescZh)
}

// auditGateInstructionsBlock 按审计档位拼接三门禁 LLM 指令（id=0 的 gate 跳过）。
// llm_description_*_check_code 均为 paper_llm_description.description_code。
func auditGateInstructionsBlock(ctx context.Context, catalog repository.PaperRefCatalogRepo, scene, auditLevelCode, locale string) string {
	level, err := catalog.FindAuditLevelByCode(ctx, strings.TrimSpace(scene), strings.TrimSpace(auditLevelCode))
	if err != nil || level == nil {
		return ""
	}
	gates := []struct {
		gate   string
		descCode uint64
	}{
		{"citation", level.LlmDescriptionCitationCheckCode},
		{"claim", level.LlmDescriptionClaimCheckCode},
		{"kill", level.LlmDescriptionKillCheckCode},
	}
	var b strings.Builder
	for _, g := range gates {
		if g.descCode == 0 {
			continue
		}
		row, err := catalog.FindLLMDescriptionByCode(ctx, g.descCode)
		if err != nil || row == nil {
			continue
		}
		desc := LLMDescForLocale(locale, row.LlmDescZh, row.LlmDescEn)
		if desc == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "- [%s] %s", g.gate, desc)
	}
	if b.Len() == 0 {
		return ""
	}
	if NormalizeContentLanguage(locale) == "en" {
		return "\n[Audit gates — apply all listed gates; tag each issue with gate=citation|claim|kill]\n" + b.String()
	}
	return "\n【审计三门禁 — 须全部适用；每条 issue 标注 gate=citation|claim|kill】\n" + b.String()
}
