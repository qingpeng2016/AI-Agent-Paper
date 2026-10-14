package topicdiscovery

import (
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/lang"
)

// 内置 inline 策略包话术（原 paper_llm_description template_topic 段；prompt 真相在策略侧，非 DB）。

func builtinSystemPrompt(locale string) string {
	if lang.NormalizeContentLanguage(locale) == "en" {
		return strings.TrimSpace(`You are an academic research assistant for the Paper Agent platform. Answer only from literature and prior-step results in the user message; do not fabricate sources or experiments. {{output_language_rule}} When JSON is required, follow the schema in the user message.`)
	}
	return strings.TrimSpace(`你是 Paper Agent 平台的学术科研助手。只根据用户消息中给出的文献与前置步骤结果作答，不得编造未出现的来源或实验。{{output_language_rule}} 若要求 JSON，严格遵守用户消息中的 schema。`)
}

func builtinStageUserTemplate(stageCode, locale string) string {
	locale = lang.NormalizeContentLanguage(locale)
	switch normalizeBuiltinStage(stageCode) {
	case "topic_generate", "generate_ideas":
		if locale == "en" {
			return strings.TrimSpace(`Stage: brainstorming + novelty in one call. At most {{max_ideas}} ideas; use only Corpus and CorpusFiles from the user message. Output JSON with ideas and novelty (lines/risks/synthesis) per the user message schema. {{output_language_rule}}`)
		}
		return strings.TrimSpace(`环节：脑暴 + 新颖性（与 novelty 同一次模型调用完成）。最多 {{max_ideas}} 条 idea；仅依据用户消息 Corpus 与 CorpusFiles；输出 JSON 含 ideas 与 novelty（lines/risks/synthesis），以用户消息 schema 为准。{{output_language_rule}}`)
	case "topic_audit", "audit":
		if locale == "en" {
			return strings.TrimSpace(`Stage: audit conclusion + literature review. Direction: {{direction}}. Target venue: {{venue}}. Re-check generate_ideas (with novelty) for fit; reviewer-style issues/summary/off_topic; literature review from Corpus (structure/content_medium/citations). Output JSON per user schema (audit + literature_review). {{output_language_rule}}`)
		}
		return strings.TrimSpace(`环节：审查结论 + 生成文献综述。方向：{{direction}}。目标期刊：{{venue}}。须再次审查第二步 generate_ideas（含 novelty）是否贴题；以严格审稿人视角给出 issues/summary/off_topic；并基于 Corpus 生成文献综述（structure/content_medium/citations）。输出 JSON 以用户消息 schema 为准（含 audit 与 literature_review）。{{output_language_rule}}`)
	case "experiment_plan":
		if locale == "en" {
			return strings.TrimSpace(`Stage: experiment planning. Direction: {{direction}}. Target venue: {{venue}}. Use LiteratureReview and topic Corpus from the user message only; do not invent baselines or data. Output a review-ready plan (hypothesis, main experiment, baselines, metrics, ablations, resources/timeline) as experiment_plan JSON per user schema. {{output_language_rule}}`)
		}
		return strings.TrimSpace(`环节：生成实验方案（experiment planning）。方向：{{direction}}。目标期刊/会议：{{venue}}。须以用户消息中的 LiteratureReview（文献综述全文/摘要）与选题 Corpus 为依据，不得编造未给出的基线或数据。输出可审稿的实验计划：核心假设、主实验与对照基线、评价指标、消融设计、算力/数据资源与时间线（写入 paper_output_experiment_plan：title、summary、content_medium、meta）。严格 JSON，字段以用户消息 schema 为准（含 experiment_plan）。{{output_language_rule}}`)
	default:
		return ""
	}
}

func normalizeBuiltinStage(stageCode string) string {
	s := strings.TrimSpace(stageCode)
	if s == topicLLMStageGenerate {
		return "topic_generate"
	}
	if s == topicLLMStageAudit {
		return "topic_audit"
	}
	return s
}

func behaviorKeyForTopicStage(stageCode string) string {
	switch normalizeBuiltinStage(stageCode) {
	case "default":
		return "topic.stage.default"
	case "topic_generate":
		return "topic.stage.topic_generate"
	case "topic_audit":
		return "topic.stage.topic_audit"
	case "experiment_plan":
		return "topic.stage.experiment_plan"
	default:
		return "topic.stage." + normalizeBuiltinStage(stageCode)
	}
}
