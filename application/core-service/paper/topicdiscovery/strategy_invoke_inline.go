package topicdiscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/internal/jsonx"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/llm"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

const (
	BehaviorTopicGenerate  = "topic.stage.topic_generate"
	BehaviorTopicAudit     = "topic.stage.topic_audit"
	BehaviorExperimentPlan = "topic.stage.experiment_plan"
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
	return route, nil
}

// invokeTopicDiscoveryBehavior 经 route 调用本地 ARIS CLI（Go 落盘 + claude/aris + skill）。
func (s *TopicDiscoveryRunService) invokeTopicDiscoveryBehavior(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	req StrategyInvokeRequest,
	userExtra string,
) (json.RawMessage, llm.LLMUsage, error) {
	route, err := s.requireTopicStrategyRoute(ctx, req.BehaviorKey)
	if err != nil {
		return nil, llm.LLMUsage{}, err
	}
	return s.invokeLocalARIS(ctx, step, req, route, userExtra)
}

func stageResultJSON(stageCode, text string) json.RawMessage {
	switch normalizeBuiltinStage(stageCode) {
	case "topic_audit":
		return json.RawMessage(jsonx.MustJSON(parseTopicAuditOnlyResponse(text)))
	default:
		return json.RawMessage(jsonx.MustJSON(parseTopicGenerateBundleResponse(text)))
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
