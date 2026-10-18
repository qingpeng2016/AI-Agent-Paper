package topicdiscovery

import (
	"context"
	"encoding/json"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/llm"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

// TopicRunInput 供 bot/scripts/idea_discovery_phases 使用。
type TopicRunInput = topicRunInput

// 以下导出仅供 application/bot/scripts/idea_discovery_phases 单段 ARIS tick 使用。

func (s *TopicDiscoveryRunService) PhasedBotRequireRoute(ctx context.Context, behaviorKey string) (*entity.PaperStrategyRoute, error) {
	return s.requireTopicStrategyRoute(ctx, behaviorKey)
}

func (s *TopicDiscoveryRunService) PhasedBotBuildInvokeContext(
	ctx context.Context,
	userID uint,
	runVersion int,
	step *entity.PaperOutputTopicStep,
	input topicRunInput,
) StrategyInvokeContext {
	return s.buildTopicStrategyInvokeContext(ctx, userID, runVersion, step, input, 1, 1, nil, nil)
}

func PhasedBotResolveRunID(paperRoot string, step *entity.PaperOutputTopicStep, input topicRunInput) string {
	return resolveIdeaDiscoveryRunID(paperRoot, step, input)
}

func (s *TopicDiscoveryRunService) PhasedBotInvokeBehavior(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	req StrategyInvokeRequest,
	userExtra string,
) (json.RawMessage, llm.LLMUsage, error) {
	return s.invokeTopicDiscoveryBehavior(ctx, step, req, userExtra)
}

func (s *TopicDiscoveryRunService) PhasedBotHandleInvokeFailure(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	paperRoot string,
	cfg ARISIdeaDiscoveryConfig,
	err error,
) error {
	return s.handleIdeaDiscoveryInvokeFailure(ctx, step, paperRoot, cfg, err)
}

func (s *TopicDiscoveryRunService) PhasedBotIngestLiterature(
	ctx context.Context,
	userID uint,
	runVersion int,
	manuscriptID uint64,
	paperRoot string,
	cfg ARISIdeaDiscoveryConfig,
) error {
	_, err := s.ingestARISWorkspaceLiterature(ctx, userID, runVersion, manuscriptID, paperRoot, cfg)
	return err
}

func (s *TopicDiscoveryRunService) PhasedBotPersistStageFailed(ctx context.Context, step *entity.PaperOutputTopicStep, reason string) error {
	return s.persistIdeaDiscoveryPhasedStepFailed(ctx, step, reason)
}

func (s *TopicDiscoveryRunService) PhasedBotFinishAfterGate(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	userID uint,
	runVersion int,
	route *entity.PaperStrategyRoute,
	paperRoot, runID string,
	cfg ARISIdeaDiscoveryConfig,
) error {
	return s.finishIdeaDiscoveryAfterGate(ctx, step, userID, runVersion, route, paperRoot, runID, cfg)
}

func (s *TopicDiscoveryRunService) PhasedBotBeginNextStageAsync(ctx context.Context, userID uint, runVersion int, stageCode string) error {
	return s.beginPhasedTopicStageAsync(ctx, userID, runVersion, stageCode)
}

func IsPhasedTopicStageCode(stageCode string) bool { return isPhasedTopicStage(stageCode) }

func TopicStageBehaviorKey(stageCode string) string { return topicStageBehaviorKey(stageCode) }

func NextPhasedTopicStageCode(stageCode string) string { return nextPhasedTopicStageCode(stageCode) }

func MustIdeaDiscoveryConfig() ARISIdeaDiscoveryConfig { return mustIdeaDiscoveryConfig() }

func IdeaDiscoveryInvokeCount(meta map[string]any) int { return ideaDiscoveryInvokeCount(meta) }

func IdeaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, phase string, cfg ARISIdeaDiscoveryConfig) bool {
	return ideaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, phase, cfg)
}

func WriteNativePlatformArguments(paperRoot, argsLine string) error {
	return writeNativePlatformArguments(paperRoot, argsLine)
}

func TruncateUTF8(s string, maxRunes int) string { return truncateUTF8(s, maxRunes) }
