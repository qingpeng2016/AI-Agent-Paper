package topicdiscovery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/internal/jsonx"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/llm"
	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"gorm.io/datatypes"
)


type arisInvokePayload struct {
	StrategyInvokeRequest
	PlatformUserPrompt string `json:"platform_user_prompt,omitempty"`
}

type localARISInvokeStdout struct {
	ResultJSON      json.RawMessage `json:"result_json"`
	ExternalTraceID string          `json:"external_trace_id"`
	RawText         string          `json:"raw_text"`
	Error           string          `json:"error"`
}

func paperARISInvokeScriptPath() string {
	if p := strings.TrimSpace(os.Getenv("PAPER_ARIS_INVOKE_SCRIPT")); p != "" {
		return p
	}
	candidates := []string{
		"libs/aris/tools/paper_strategy_invoke.py",
		filepath.Join("AI-Agent-Paper", "libs/aris/tools/paper_strategy_invoke.py"),
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "libs/aris/tools/paper_strategy_invoke.py"))
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return candidates[0]
}

func paperARISSkillsRoot() string {
	if p := strings.TrimSpace(os.Getenv("PAPER_ARIS_SKILLS_ROOT")); p != "" {
		return p
	}
	for _, c := range []string{"libs/aris/skills", filepath.Join("AI-Agent-Paper", "libs/aris/skills")} {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return ""
}

func (s *TopicDiscoveryRunService) invokeLocalARIS(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	req StrategyInvokeRequest,
	route *entity.PaperStrategyRoute,
	userExtra string,
) (json.RawMessage, llm.LLMUsage, error) {
	script := paperARISInvokeScriptPath()
	if st, err := os.Stat(script); err != nil || st.IsDir() {
		return nil, llm.LLMUsage{}, errorx.ErrLLMNotConfigured.WithDetail(
			"本地 ARIS 脚本不存在: " + script + "（设置 PAPER_ARIS_INVOKE_SCRIPT）",
		)
	}

	payload := arisInvokePayload{
		StrategyInvokeRequest: req,
		PlatformUserPrompt:    strings.TrimSpace(userExtra),
	}
	stdinJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, llm.LLMUsage{}, err
	}

	timeout := 3600 * time.Second
	if v := strings.TrimSpace(os.Getenv("PAPER_ARIS_INVOKE_TIMEOUT_SEC")); v != "" {
		if sec, parseErr := parsePositiveInt(v); parseErr == nil && sec > 0 {
			timeout = time.Duration(sec) * time.Second
		}
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, "python3", script)
	cmd.Stdin = bytes.NewReader(stdinJSON)
	cmd.Env = os.Environ()
	if root := paperARISSkillsRoot(); root != "" {
		cmd.Env = append(cmd.Env, "PAPER_ARIS_SKILLS_ROOT="+root)
	}

	start := time.Now()
	out, err := cmd.Output()
	latencyMs := int(time.Since(start).Milliseconds())
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && len(ee.Stderr) > 0 {
			err = fmt.Errorf("%w: %s", err, strings.TrimSpace(string(ee.Stderr)))
		}
		s.persistARISInvokeLog(ctx, step, route, req, string(stdinJSON), "", latencyMs, err)
		return nil, llm.LLMUsage{}, errorx.ErrLLMCallFailed.WithDetail(err.Error())
	}

	var parsed localARISInvokeStdout
	if err := json.Unmarshal(out, &parsed); err != nil {
		s.persistARISInvokeLog(ctx, step, route, req, string(stdinJSON), string(out), latencyMs, err)
		return nil, llm.LLMUsage{}, errorx.ErrLLMCallFailed.WithDetail("ARIS stdout 非 JSON: " + err.Error())
	}
	if strings.TrimSpace(parsed.Error) != "" {
		err = fmt.Errorf("%s", parsed.Error)
		s.persistARISInvokeLog(ctx, step, route, req, string(stdinJSON), string(out), latencyMs, err)
		return nil, llm.LLMUsage{}, errorx.ErrLLMCallFailed.WithDetail(parsed.Error)
	}
	if len(parsed.ResultJSON) == 0 {
		err = fmt.Errorf("ARIS 未返回 result_json")
		s.persistARISInvokeLog(ctx, step, route, req, string(stdinJSON), string(out), latencyMs, err)
		return nil, llm.LLMUsage{}, errorx.ErrLLMCallFailed.WithDetail(err.Error())
	}

	stageCode := topicStageCodeFromBehaviorKey(req.BehaviorKey)
	result := stageResultJSONFromObject(stageCode, parsed.ResultJSON, parsed.RawText)
	s.persistARISInvokeLog(ctx, step, route, req, string(stdinJSON), string(out), latencyMs, nil)
	return result, llm.LLMUsage{}, nil
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
	row := &entity.PaperLLMCallLog{
		ManuscriptID:      step.ManuscriptID,
		StageCode:         &stage,
		BehaviorKey:       &bk,
		StrategyPackageID: pkgID,
		CapabilityID:      &capID,
		ModelName:         "local-aris-cli",
		RequestJSON:       datatypes.JSON(requestJSON),
		ResponseJSON:      datatypes.JSON(responseJSON),
		LatencyMs:         latencyMs,
		Status:            status,
		CreatedAt:         time.Now(),
	}
	_ = s.llmRepo.InsertCallLog(ctx, row)
}
