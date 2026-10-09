package paper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

// stageIdeasAndNovelty 一次 LLM：generate_ideas 落盘并 completed；novelty 预填 result/summary/extra，status 保持 pending。
func (s *TopicDiscoveryRunService) stageIdeasAndNovelty(
	ctx context.Context,
	userID uint,
	runVersion int,
	ideasStep *entity.PaperOutputTopicStep,
	input topicRunInput,
) error {
	noveltyStep, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "novelty")
	if err != nil || noveltyStep == nil {
		return fmt.Errorf("novelty step not found")
	}

	ctxBlock, _ := s.stageContextCompact(ctx, userID, runVersion, "retrieve")
	filesBlock := s.formatRetrieveCorpusFilesBlock(ctx, userID, runVersion)

	userMsg := fmt.Sprintf(`Research direction: %s
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
规则：ideas 最多 %d 条；每条 reference_keys 必须来自 Corpus/CorpusFiles；novelty.risks 须逐条对应 ideas 标题；文本字段用中文。`,
		input.Direction, input.Venue, input.MaxIdeas, ctxBlock, filesBlock, input.MaxIdeas)

	text, usage, err := s.callStageLLM(ctx, ideasStep, "generate_ideas", map[string]string{
		"direction": input.Direction,
		"max_ideas": fmt.Sprintf("%d", input.MaxIdeas),
	}, userMsg)
	if err != nil {
		return err
	}

	ideasResult, noveltyResult, summary := parseIdeasNoveltyLLMResponse(text)
	ideasStep.Result = mustJSON(ideasResult)
	if summary != "" {
		ideasStep.SummaryText = ptrString(summary)
	}
	appendUsageMeta(ideasStep, usage)
	meta := LoadStepExtra(ideasStep)
	meta["llm_combined_with"] = "novelty"
	ideasStep.Extra = mustJSON(meta)

	noveltyStep.Result = mustJSON(noveltyResult)
	noveltySummary := noveltySummaryFromResult(noveltyResult)
	if noveltySummary != "" {
		noveltyStep.SummaryText = ptrString(noveltySummary)
	} else if summary != "" {
		noveltyStep.SummaryText = ptrString(summary)
	}
	novMeta := LoadStepExtra(noveltyStep)
	novMeta["llm_combined_with"] = "generate_ideas"
	novMeta["prefilled"] = true
	noveltyStep.Extra = mustJSON(novMeta)
	// 保持 pending：不在此改 status / started_at / completed_at
	return s.steps.SaveStep(ctx, noveltyStep)
}

// stageNoveltyAck 用户确认检查点：novelty 已在 generate_ideas 预填，不再调模型。
func (s *TopicDiscoveryRunService) stageNoveltyAck(_ context.Context, step *entity.PaperOutputTopicStep) error {
	if len(step.Result) == 0 {
		return fmt.Errorf("novelty 无预填 result，无法确认")
	}
	return nil
}

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
