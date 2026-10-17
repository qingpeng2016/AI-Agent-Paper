package topicdiscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/internal/jsonx"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/llm"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"gorm.io/datatypes"
)

func (s *TopicDiscoveryRunService) invokeLocalARIS(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	req StrategyInvokeRequest,
	route *entity.PaperStrategyRoute,
	userExtra string,
) (json.RawMessage, llm.LLMUsage, error) {
	reqJSON, _ := json.Marshal(map[string]any{
		"behavior_key":           req.BehaviorKey,
		"capability_id":          req.CapabilityID,
		"context":                req.Context,
		"options":                req.Options,
		"platform_user_prompt":   strings.TrimSpace(userExtra),
	})

	start := time.Now()
	out, err := runLocalStrategyInvoke(ctx, req, userExtra)
	latencyMs := int(time.Since(start).Milliseconds())
	if err != nil {
		s.persistARISInvokeLog(ctx, step, route, req, string(reqJSON), "", latencyMs, llm.LLMUsage{}, nil, err)
		return nil, llm.LLMUsage{}, errorx.ErrLLMCallFailed.WithDetail(err.Error())
	}
	if out == nil || len(out.ResultJSON) == 0 {
		err = fmt.Errorf("ARIS 未返回 result_json")
		s.persistARISInvokeLog(ctx, step, route, req, string(reqJSON), "", latencyMs, llm.LLMUsage{}, nil, err)
		return nil, llm.LLMUsage{}, errorx.ErrLLMCallFailed.WithDetail(err.Error())
	}

	respJSON, _ := json.Marshal(map[string]any{
		"result_json":       json.RawMessage(out.ResultJSON),
		"external_trace_id": out.ExternalTraceID,
		"raw_text":          out.RawText,
	})
	stageCode := topicStageCodeFromBehaviorKey(req.BehaviorKey)
	result := stageResultJSONFromObject(stageCode, out.ResultJSON, out.RawText)
	s.persistARISInvokeLog(
		ctx, step, route, req,
		clipJSONForMySQL(string(reqJSON)),
		clipJSONForMySQL(string(respJSON)),
		latencyMs, out.Usage, out, nil,
	)
	return result, out.Usage, nil
}

func (s *TopicDiscoveryRunService) resolveModelConfigForRoute(
	ctx context.Context,
	route *entity.PaperStrategyRoute,
) (*entity.PaperLLMModelConfig, error) {
	if s == nil || s.llmRepo == nil || route == nil {
		return nil, nil
	}
	if route.DefaultModelConfigID > 0 {
		return s.llmRepo.GetActiveModelByID(ctx, uint(route.DefaultModelConfigID))
	}
	return s.llmRepo.ResolveActiveModel(ctx, 0)
}

func stageResultJSONFromObject(stageCode string, raw json.RawMessage, rawText string) json.RawMessage {
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil || len(obj) == 0 {
		if rawText != "" {
			return stageResultJSON(stageCode, rawText)
		}
		return raw
	}
	switch normalizeBuiltinStage(stageCode) {
	case "topic_audit":
		if _, ok := obj["audit"]; ok {
			return json.RawMessage(jsonx.MustJSON(obj))
		}
		return json.RawMessage(jsonx.MustJSON(map[string]any{"audit": obj}))
	default:
		return json.RawMessage(jsonx.MustJSON(obj))
	}
}

func parsePositiveInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func (s *TopicDiscoveryRunService) persistARISInvokeLog(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	route *entity.PaperStrategyRoute,
	req StrategyInvokeRequest,
	requestJSON, responseJSON string,
	latencyMs int,
	usage llm.LLMUsage,
	invokeOut *localARISInvokeResult,
	callErr error,
) {
	if s == nil || s.llmRepo == nil || route == nil {
		return
	}
	status := int16(200)
	if callErr != nil {
		status = 500
	}
	stage := topicStageCodeFromBehaviorKey(req.BehaviorKey)
	bk := req.BehaviorKey
	capID := req.CapabilityID
	if capID == "" {
		capID = route.CapabilityID
	}
	var pkgID *uint64
	if route.StrategyPackageID > 0 {
		id := route.StrategyPackageID
		pkgID = &id
	}

	modelName := "local-aris-cli"
	var modelConfigID *uint
	if cfg, _ := s.resolveModelConfigForRoute(ctx, route); cfg != nil {
		modelName = cfg.ModelName
		mid := cfg.ID
		modelConfigID = &mid
	} else if invokeOut != nil {
		if invokeOut.CLIModel != "" {
			modelName = invokeOut.CLIModel
		}
		if invokeOut.CLIExecutor != "" {
			modelName = invokeOut.CLIExecutor + ":" + modelName
		}
	}

	var promptTok, completionTok *int
	if usage.PromptTokens > 0 {
		p := usage.PromptTokens
		promptTok = &p
	}
	if usage.CompletionTokens > 0 {
		c := usage.CompletionTokens
		completionTok = &c
	}

	var traceID *string
	if invokeOut != nil && strings.TrimSpace(invokeOut.ExternalTraceID) != "" {
		t := invokeOut.ExternalTraceID
		traceID = &t
	}

	row := &entity.PaperLLMCallLog{
		ManuscriptID:      step.ManuscriptID,
		StageCode:         &stage,
		BehaviorKey:       &bk,
		StrategyPackageID: pkgID,
		CapabilityID:      &capID,
		ExternalTraceID:   traceID,
		ModelConfigID:      modelConfigID,
		ModelName:         modelName,
		RequestJSON:       datatypes.JSON(requestJSON),
		ResponseJSON:      datatypes.JSON(responseJSON),
		UseInTokens:  promptTok,
		UseOutTokens: completionTok,
		LatencyMs:         latencyMs,
		Status:            status,
		CreatedAt:         time.Now(),
	}
	if err := s.llmRepo.InsertCallLog(ctx, row); err != nil {
		return
	}
	if modelConfigID != nil && (usage.PromptTokens > 0 || usage.CompletionTokens > 0) {
		_ = s.llmRepo.AddTokensUsed(ctx, *modelConfigID, usage.PromptTokens, usage.CompletionTokens)
	}
}
