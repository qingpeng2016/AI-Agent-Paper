package paper

import (
	"fmt"
	"strings"
)

// NormalizeContentLanguage 归一化为 zh 或 en；默认 en。
func NormalizeContentLanguage(s string) string {
	if strings.EqualFold(strings.TrimSpace(s), "zh") {
		return "zh"
	}
	return "en"
}

func jsonStringFieldsLocaleHint(locale string) string {
	if NormalizeContentLanguage(locale) == "en" {
		return "Use English for all string fields in JSON."
	}
	return "JSON 内字符串字段使用中文。"
}

// PromptLocaleVars 供 paper_template_llm_prompt 占位符（与论文 content_language 一致）。
func PromptLocaleVars(locale string) map[string]string {
	locale = NormalizeContentLanguage(locale)
	if locale == "en" {
		return map[string]string{
			"content_language": "en",
			"output_language_rule": "Manuscript working language is English: all natural-language output, Markdown bodies, " +
				"and JSON string fields must be written in English (proper nouns and standard acronyms excepted).",
		}
	}
	return map[string]string{
		"content_language": "zh",
		"output_language_rule": "本论文工作语言为中文：所有自然语言输出、Markdown 正文与 JSON 字符串字段须使用中文（专有名词、标准缩写除外）。",
	}
}

func mergePromptVars(vars map[string]string, locale string) map[string]string {
	out := make(map[string]string, len(vars)+4)
	for k, v := range vars {
		out[k] = v
	}
	for k, v := range PromptLocaleVars(locale) {
		out[k] = v
	}
	return out
}

func buildIdeasNoveltyUserMessage(input topicRunInput, ctxBlock, filesBlock, locale string) string {
	locale = NormalizeContentLanguage(locale)
	if locale == "en" {
		return fmt.Sprintf(`Research direction: %s
Target venue: %s
Max ideas: %d

Corpus:
%s

CorpusFiles (local PDFs bound to external_key; Corpus titles/abstracts are supplementary):
%s

Return ONLY valid JSON (ideas + novelty in one response):
{
  "ideas": [
    {"title":"","problem":"","approach":"","contribution":"","reference_keys":[]}
  ],
  "novelty": {
    "lines": ["…"],
    "risks": [{"idea_title":"","risk":"low|medium|high","note":"","overlap_refs":[]}],
    "synthesis": ""
  }
}
Rules: at most %d ideas; each reference_keys must come from Corpus/CorpusFiles; novelty.risks must align with idea titles; %s`,
			input.Direction, input.Venue, input.MaxIdeas, ctxBlock, filesBlock, input.MaxIdeas, jsonStringFieldsLocaleHint(locale))
	}
	return fmt.Sprintf(`Research direction: %s
Target venue: %s
Max ideas: %d

Corpus:
%s

CorpusFiles（已与 external_key 绑定的本地 PDF；Corpus 内摘要/标题仅作补充）:
%s

Return ONLY valid JSON（一次输出脑暴 + 新颖性，不要分两次）:
{
  "ideas": [
    {"title":"","problem":"","approach":"","contribution":"","reference_keys":[]}
  ],
  "novelty": {
    "lines": ["…"],
    "risks": [{"idea_title":"","risk":"low|medium|high","note":"","overlap_refs":[]}],
    "synthesis": ""
  }
}
规则：ideas 最多 %d 条；每条 reference_keys 必须来自 Corpus/CorpusFiles；novelty.risks 须逐条对应 ideas 标题；%s`,
		input.Direction, input.Venue, input.MaxIdeas, ctxBlock, filesBlock, input.MaxIdeas, jsonStringFieldsLocaleHint(locale))
}
