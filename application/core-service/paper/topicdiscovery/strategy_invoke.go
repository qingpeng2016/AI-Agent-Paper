package topicdiscovery

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

// StrategyInvokeRequest 平台 → 策略包（ARIS）单次调用。
type StrategyInvokeRequest struct {
	BehaviorKey  string                 `json:"behavior_key"`
	CapabilityID string                 `json:"capability_id"`
	Context      StrategyInvokeContext  `json:"context"`
	Options      StrategyInvokeOptions  `json:"options"`
}

type StrategyInvokeContext struct {
	ManuscriptID       uint64          `json:"manuscript_id"`
	UserID             uint64          `json:"user_id"`
	RunVersion         int             `json:"run_version"`
	ContentLanguage    string          `json:"content_language"`
	Direction          string          `json:"direction"`
	Keywords           []string        `json:"keywords,omitempty"`
	VenueCodes         []string        `json:"venue_codes,omitempty"`
	VenueLabel         string          `json:"venue_label,omitempty"`
	StrategyProfileKey string          `json:"strategy_profile_key,omitempty"`
	Corpus             json.RawMessage `json:"corpus,omitempty"`
	PrevBundle         json.RawMessage `json:"prev_bundle,omitempty"`
	PrevAudit          json.RawMessage `json:"prev_audit,omitempty"`
	Round              int             `json:"round,omitempty"`
	MaxRounds          int             `json:"max_rounds,omitempty"`
}

type StrategyInvokeOptions struct {
	AuditLevelCode   string   `json:"audit_level_code"`
	IntensityCode    string   `json:"intensity_code"`
	ExecRounds       int      `json:"exec_rounds"`
	MaxIdeas         int      `json:"max_ideas"`
	MaxPapers        int      `json:"max_papers"`
	SourceCodes      []string `json:"source_codes,omitempty"`
	RandomReferences bool     `json:"random_references"`
}

type StrategyInvokeResponse struct {
	// 策略包约定 JSON（如 ideas/novelty 或 audit/literature_review）
	ResultJSON      json.RawMessage `json:"result_json"`
	ExternalTraceID string          `json:"external_trace_id,omitempty"`
}

// StrategyInvoker 由 inline 策略包实现：本地 libs/aris（子进程/CLI），不是 HTTP。
type StrategyInvoker interface {
	Invoke(ctx context.Context, req StrategyInvokeRequest) (*StrategyInvokeResponse, error)
}

func buildTopicDiscoveryOptions(input topicRunInput) StrategyInvokeOptions {
	return StrategyInvokeOptions{
		AuditLevelCode:   input.AuditLevel,
		IntensityCode:    input.Intensity,
		ExecRounds:       input.ExecRounds,
		MaxIdeas:         input.MaxIdeas,
		MaxPapers:        input.MaxPapers,
		SourceCodes:      input.SourceCodes,
		RandomReferences: input.RandomReferences,
	}
}

// BuildTopicDiscoveryInvokeRequest 选题发现单轮：generate 或 audit 各 Invoke 一次。
func BuildTopicDiscoveryInvokeRequest(
	behaviorKey, capabilityID string,
	input topicRunInput,
	ctx StrategyInvokeContext,
) StrategyInvokeRequest {
	return StrategyInvokeRequest{
		BehaviorKey:  behaviorKey,
		CapabilityID: capabilityID,
		Context:      ctx,
		Options:      buildTopicDiscoveryOptions(input),
	}
}

// InvokeTopicDiscoveryRound 选题发现单轮：generate route → audit route（capability_id 以 DB route 为准）。
func InvokeTopicDiscoveryRound(
	ctx context.Context,
	invoker StrategyInvoker,
	input topicRunInput,
	base StrategyInvokeContext,
	genRoute, auditRoute *entity.PaperStrategyRoute,
) (bundleJSON, auditJSON json.RawMessage, err error) {
	if invoker == nil {
		return nil, nil, fmt.Errorf("strategy invoker is nil")
	}
	if genRoute == nil || auditRoute == nil {
		return nil, nil, fmt.Errorf("strategy route is nil")
	}

	genReq := BuildTopicDiscoveryInvokeRequest(
		genRoute.BehaviorKey, genRoute.CapabilityID, input, base,
	)
	genResp, err := invoker.Invoke(ctx, genReq)
	if err != nil {
		return nil, nil, err
	}
	bundleJSON = genResp.ResultJSON

	auditCtx := base
	auditCtx.PrevBundle = bundleJSON
	auditReq := BuildTopicDiscoveryInvokeRequest(
		auditRoute.BehaviorKey, auditRoute.CapabilityID, input, auditCtx,
	)
	auditResp, err := invoker.Invoke(ctx, auditReq)
	if err != nil {
		return bundleJSON, nil, err
	}
	return bundleJSON, auditResp.ResultJSON, nil
}
