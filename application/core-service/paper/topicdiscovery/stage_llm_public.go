package topicdiscovery

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/llm"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

// StageLLMResolved 供 experimentplan 等子包使用的阶段 LLM 解析结果。
type StageLLMResolved struct {
	System     string
	UserPrefix string
	Model      *entity.PaperLLMModelConfig
}

func (s *TopicDiscoveryRunService) ContentLocaleForStep(ctx context.Context, step *entity.PaperOutputTopicStep) string {
	return s.contentLocaleForStep(ctx, step)
}

func (s *TopicDiscoveryRunService) ResolveStageLLM(
	ctx context.Context,
	stageCode string,
	vars map[string]string,
	locale string,
) (*StageLLMResolved, error) {
	r, err := s.resolveStageLLM(ctx, stageCode, vars, locale)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return &StageLLMResolved{
		System:     r.system,
		UserPrefix: r.userPrefix,
		Model:      r.model,
	}, nil
}

func (s *TopicDiscoveryRunService) CallStageLLM(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	stageCode string,
	vars map[string]string,
	userExtra string,
) (string, llm.LLMUsage, error) {
	return s.callStageLLM(ctx, step, stageCode, vars, userExtra)
}

func LLMUserMessage(userPrefix, userExtra string) string {
	return llmUserMessage(userPrefix, userExtra)
}

func ParseJSONOrWrap(text, key string) (map[string]any, string) {
	return parseJSONOrWrap(text, key)
}
