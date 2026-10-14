package topicdiscovery

import (
	"context"
	"encoding/json"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/llm"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

func (s *TopicDiscoveryRunService) ContentLocaleForStep(ctx context.Context, step *entity.PaperOutputTopicStep) string {
	return s.contentLocaleForStep(ctx, step)
}

func LLMUserMessage(userPrefix, userExtra string) string {
	return llmUserMessage(userPrefix, userExtra)
}

func ParseJSONOrWrap(text, key string) (map[string]any, string) {
	return parseJSONOrWrap(text, key)
}

// InvokeExperimentPlan 实验方案 bot：本地 ARIS CLI。
func (s *TopicDiscoveryRunService) InvokeExperimentPlan(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	invokeCtx StrategyInvokeContext,
	userExtra string,
) (json.RawMessage, llm.LLMUsage, error) {
	return s.InvokeBehavior(ctx, step, BehaviorExperimentPlan, topicRunInput{}, invokeCtx, userExtra)
}

// InvokeBehavior 经 paper_llm_strategy_route 走本地 ARIS CLI（stdin JSON）。
func (s *TopicDiscoveryRunService) InvokeBehavior(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	behaviorKey string,
	input topicRunInput,
	invokeCtx StrategyInvokeContext,
	userExtra string,
) (json.RawMessage, llm.LLMUsage, error) {
	route, err := s.requireTopicStrategyRoute(ctx, behaviorKey)
	if err != nil {
		return nil, llm.LLMUsage{}, err
	}
	req := BuildTopicDiscoveryInvokeRequest(behaviorKey, route.CapabilityID, input, invokeCtx)
	return s.invokeTopicDiscoveryBehavior(ctx, step, req, userExtra)
}
