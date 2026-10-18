package topicdiscovery

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

const topicDiscoveryInputStageCode = "idea_discovery"

// runInputStep 选题 run 表单快照：新 run 在 idea_discovery；旧 run 可能在 retrieve。
func (s *TopicDiscoveryRunService) runInputStep(
	ctx context.Context,
	userID uint,
	runVersion int,
) (*entity.PaperOutputTopicStep, error) {
	if s == nil || s.steps == nil {
		return nil, nil
	}
	for _, code := range []string{"research_lit", topicDiscoveryInputStageCode, "retrieve"} {
		step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, code)
		if err != nil {
			return nil, err
		}
		if step != nil && len(step.InputParams) > 0 {
			return step, nil
		}
	}
	return nil, nil
}

// literatureCorpusStep 文献 hits / downloads / brief：优先 idea_discovery.result，兼容 retrieve。
func (s *TopicDiscoveryRunService) literatureCorpusStep(
	ctx context.Context,
	userID uint,
	runVersion int,
) (*entity.PaperOutputTopicStep, error) {
	if s == nil || s.steps == nil {
		return nil, nil
	}
	for _, code := range []string{"research_lit", topicDiscoveryInputStageCode, "retrieve"} {
		step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, code)
		if err != nil {
			return nil, err
		}
		if step != nil && len(step.Result) > 0 {
			return step, nil
		}
	}
	return nil, nil
}
