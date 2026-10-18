package topicdiscovery

import (
	"context"
	"strings"
)

// topicDiscoveryPhasedStages：工作流 1 五段 ARIS skill，每段一条 paper_output_topic_step；最后 topic_confirm。
var topicDiscoveryPhasedStages = []string{
	"research_lit",
	"idea_creator",
	"novelty_check",
	"research_review",
	"research_refine",
	"topic_confirm",
}

// TopicDiscoveryPhasedStageCodes 供 bot 按序扫描 running step。
func TopicDiscoveryPhasedStageCodes() []string {
	out := make([]string, 0, len(topicDiscoveryPhasedStages))
	for _, s := range topicDiscoveryPhasedStages {
		if s == "topic_confirm" {
			continue
		}
		out = append(out, s)
	}
	return out
}

func isPhasedTopicStage(stageCode string) bool {
	for _, s := range topicDiscoveryPhasedStages {
		if s == stageCode {
			return stageCode != "topic_confirm"
		}
	}
	return false
}

func topicStageCodeToPhaseID(stageCode string) string {
	switch strings.TrimSpace(stageCode) {
	case "research_lit":
		return "research-lit"
	case "idea_creator":
		return "idea-creator"
	case "novelty_check":
		return "novelty-check"
	case "research_review":
		return "research-review"
	case "research_refine":
		return "research-refine-pipeline"
	default:
		return strings.ReplaceAll(stageCode, "_", "-")
	}
}

func topicStageBehaviorKey(stageCode string) string {
	return "topic.stage." + strings.TrimSpace(stageCode)
}

func nextPhasedTopicStageCode(stageCode string) string {
	for i, s := range topicDiscoveryPhasedStages {
		if s == stageCode && i+1 < len(topicDiscoveryPhasedStages) {
			return topicDiscoveryPhasedStages[i+1]
		}
	}
	return ""
}

func (s *TopicDiscoveryRunService) stagesForRun(ctx context.Context, userID uint, runVersion int) []string {
	step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "research_lit")
	if err == nil && step != nil {
		return topicDiscoveryPhasedStages
	}
	return topicDiscoveryStages
}

func (s *TopicDiscoveryRunService) phasedStagesPriorCompleted(ctx context.Context, userID uint, runVersion int, stageCode string) (bool, error) {
	for _, code := range topicDiscoveryPhasedStages {
		if code == stageCode {
			return true, nil
		}
		if code == "topic_confirm" {
			break
		}
		step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, code)
		if err != nil {
			return false, err
		}
		if step == nil || strings.TrimSpace(step.Status) != "completed" {
			return false, nil
		}
	}
	return true, nil
}
