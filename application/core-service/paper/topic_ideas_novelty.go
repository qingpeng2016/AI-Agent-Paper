package paper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

// stageBeginGenerateIdeasAsync 仅标记 running，由 bot 异步调用 LLM（与 retrieve PDF 下载同理）。
func (s *TopicDiscoveryRunService) stageBeginGenerateIdeasAsync(
	ctx context.Context,
	userID uint,
	runVersion int,
	step *entity.PaperOutputTopicStep,
) error {
	ret, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "retrieve")
	if err != nil || ret == nil || strings.TrimSpace(ret.Status) != "completed" {
		return fmt.Errorf("retrieve 未完成，无法开始脑暴")
	}
	if len(step.Result) > 0 && strings.TrimSpace(step.Status) == "completed" {
		return nil
	}
	sum := "脑暴与新颖性分析中…"
	step.SummaryText = &sum
	step.Status = "running"
	step.CompletedAt = nil
	meta := LoadStepExtra(step)
	meta["async_llm"] = true
	step.Extra = mustJSON(meta)
	return s.steps.SaveStep(ctx, step)
}

// RunIdeasAndNoveltyLLM bot：一次 LLM，generate_ideas completed（novelty 写入同一步 result）。
func (s *TopicDiscoveryRunService) RunIdeasAndNoveltyLLM(
	ctx context.Context,
	userID uint,
	runVersion int,
	ideasStep *entity.PaperOutputTopicStep,
	input topicRunInput,
) error {
	ret, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "retrieve")
	if err != nil || ret == nil || strings.TrimSpace(ret.Status) != "completed" {
		return fmt.Errorf("retrieve 未完成")
	}
	ctxBlock, _ := s.stageContextCompact(ctx, userID, runVersion, "retrieve")
	filesBlock := s.formatRetrieveCorpusFilesBlock(ctx, userID, runVersion)
	locale := s.contentLocaleForStep(ctx, ideasStep)
	userMsg := buildIdeasNoveltyUserMessage(input, ctxBlock, filesBlock, locale)

	text, usage, err := s.callStageLLM(ctx, ideasStep, "generate_ideas", map[string]string{
		"direction": input.Direction,
		"max_ideas": fmt.Sprintf("%d", input.MaxIdeas),
	}, userMsg)
	if err != nil {
		return err
	}

	ideasResult, noveltyResult, summary := parseIdeasNoveltyLLMResponse(text)
	if len(noveltyResult) > 0 {
		ideasResult["novelty"] = noveltyResult
	}
	ideasStep.Result = mustJSON(ideasResult)
	if summary != "" {
		ideasStep.SummaryText = ptrString(summary)
	} else if syn := noveltySummaryFromResult(noveltyResult); syn != "" {
		ideasStep.SummaryText = ptrString(syn)
	}
	appendUsageMeta(ideasStep, usage)
	meta := LoadStepExtra(ideasStep)
	meta["includes_novelty"] = true
	delete(meta, "async_llm")
	meta["async_llm_done"] = true
	ideasStep.Extra = mustJSON(meta)

	done := time.Now()
	ideasStep.Status = "completed"
	ideasStep.CompletedAt = &done
	return s.steps.SaveStep(ctx, ideasStep)
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

// noveltyBlockForAudit 新颖性取自 generate_ideas.result.novelty；旧 run 回退 novelty 步。
func (s *TopicDiscoveryRunService) noveltyBlockForAudit(ctx context.Context, userID uint, runVersion int) (noveltyJSON, noveltySummary string) {
	ideasRaw := strings.TrimSpace(s.stepResultText(ctx, userID, runVersion, "generate_ideas"))
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
