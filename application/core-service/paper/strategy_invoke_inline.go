package paper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

const (
	BehaviorTopicGenerate = "topic.stage.topic_generate"
	BehaviorTopicAudit    = "topic.stage.topic_audit"
)

func topicStageCodeFromBehaviorKey(behaviorKey string) string {
	switch strings.TrimSpace(behaviorKey) {
	case BehaviorTopicGenerate:
		return topicLLMStageGenerate
	case BehaviorTopicAudit:
		return topicLLMStageAudit
	default:
		return strings.TrimPrefix(strings.TrimSpace(behaviorKey), "topic.stage.")
	}
}

func strategyRouteMeta(route *entity.PaperStrategyRoute) map[string]any {
	if route == nil {
		return nil
	}
	m := map[string]any{
		"behavior_key":          route.BehaviorKey,
		"capability_id":         route.CapabilityID,
		"strategy_package_id":   route.StrategyPackageID,
		"default_model_config_id": route.DefaultModelConfigID,
	}
	if route.Package != nil {
		m["strategy_package_code"] = route.Package.Code
		m["provider_kind"] = route.Package.ProviderKind
	}
	return m
}

func (s *TopicDiscoveryRunService) requireTopicStrategyRoute(ctx context.Context, behaviorKey string) (*entity.PaperStrategyRoute, error) {
	route, err := s.resolveStrategyRoute(ctx, behaviorKey)
	if err != nil {
		return nil, err
	}
	if route == nil {
		return nil, errorx.ErrLLMNotConfigured.WithDetail(
			"缺少 paper_llm_strategy_route（behavior_key=" + behaviorKey + "）",
		)
	}
	pkg := route.Package
	if pkg == nil && s.strategyRepo != nil {
		pkg, _ = s.strategyRepo.FindPackageByID(ctx, route.StrategyPackageID)
		route.Package = pkg
	}
	if pkg == nil {
		return nil, errorx.ErrLLMNotConfigured.WithDetail(
			"缺少 paper_llm_strategy_package（id=" + fmt.Sprintf("%d", route.StrategyPackageID) + "）",
		)
	}
	if pkg.Status != "active" {
		return nil, errorx.ErrLLMNotConfigured.WithDetail("策略包 " + pkg.Code + " 未 active")
	}
	if pkg.ProviderKind == entity.StrategyProviderHTTP {
		return nil, errorx.ErrLLMNotConfigured.WithDetail(
			"策略包 " + pkg.Code + " 为 http 远程服务，尚未接入 Invoke（behavior_key=" + behaviorKey + "）",
		)
	}
	return route, nil
}

// invokeTopicDiscoveryBehavior 经 route → inline 策略包执行一次 capability（内置话术 + 模型），返回约定 JSON。
func (s *TopicDiscoveryRunService) invokeTopicDiscoveryBehavior(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	req StrategyInvokeRequest,
	vars map[string]string,
	userExtra string,
) (json.RawMessage, LLMUsage, error) {
	if _, err := s.requireTopicStrategyRoute(ctx, req.BehaviorKey); err != nil {
		return nil, LLMUsage{}, err
	}
	stageCode := topicStageCodeFromBehaviorKey(req.BehaviorKey)
	text, usage, err := s.callStageLLM(ctx, step, stageCode, vars, userExtra)
	if err != nil {
		return nil, usage, err
	}
	return stageResultJSON(stageCode, text), usage, nil
}

func stageResultJSON(stageCode, text string) json.RawMessage {
	switch normalizeBuiltinStage(stageCode) {
	case "topic_audit":
		return json.RawMessage(mustJSON(parseTopicAuditOnlyResponse(text)))
	default:
		return json.RawMessage(mustJSON(parseTopicGenerateBundleResponse(text)))
	}
}

func (s *TopicDiscoveryRunService) buildTopicStrategyInvokeContext(
	ctx context.Context,
	userID uint,
	runVersion int,
	step *entity.PaperOutputTopicStep,
	input topicRunInput,
	round, maxRounds int,
	prevBundle, prevAudit json.RawMessage,
) StrategyInvokeContext {
	var corpus json.RawMessage
	if ret, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "retrieve"); err == nil && ret != nil && len(ret.Result) > 0 {
		corpus = json.RawMessage(ret.Result)
	}
	return StrategyInvokeContext{
		ManuscriptID:    step.ManuscriptID,
		UserID:          uint64(userID),
		RunVersion:      runVersion,
		ContentLanguage: s.contentLocaleForStep(ctx, step),
		Direction:       input.Direction,
		Keywords:        input.Keywords,
		VenueCodes:      input.VenueCodes,
		VenueLabel:      input.Venue,
		Corpus:          corpus,
		PrevBundle:      prevBundle,
		PrevAudit:       prevAudit,
		Round:           round,
		MaxRounds:       maxRounds,
	}
}
