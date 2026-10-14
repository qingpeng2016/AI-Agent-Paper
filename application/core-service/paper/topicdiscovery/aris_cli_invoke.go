package topicdiscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/storage"
)

var arisTraceSanitize = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

type localARISInvokeResult struct {
	ResultJSON      json.RawMessage
	ExternalTraceID string
	RawText         string
}

func resolveCapability(req StrategyInvokeRequest) string {
	cap := strings.TrimSpace(req.CapabilityID)
	behavior := strings.TrimSpace(req.BehaviorKey)
	if cap != "" && !strings.HasPrefix(cap, "topic.stage.") {
		return strings.TrimPrefix(cap, "/")
	}
	if strings.Contains(behavior, "topic_audit") || strings.HasSuffix(behavior, "topic_audit") {
		return "research-review"
	}
	if strings.Contains(behavior, "topic_generate") || strings.HasSuffix(behavior, "topic_generate") {
		return "idea-creator"
	}
	if strings.Contains(behavior, "experiment_plan") || strings.HasSuffix(behavior, "experiment_plan") {
		return "experiment-bridge"
	}
	if strings.Contains(behavior, "discovery.run") || strings.HasSuffix(behavior, "discovery.run") {
		return "idea-discovery"
	}
	if cap != "" && !strings.HasPrefix(cap, "topic.") {
		return strings.TrimPrefix(cap, "/")
	}
	return ""
}

func buildHeadlessPrompt(capability, behavior string) string {
	return fmt.Sprintf(`[AI-Agent-Paper headless invoke]
Execute research skill /%s in this workspace.

1. Read structured input: .paper/invoke/request.json (context, options, behavior_key).
2. Read literature corpus if present: .paper/corpus.json.
3. Follow the skill workflow where practical, but **your final assistant message must be ONLY one valid JSON object**
   matching the platform schema described in .paper/invoke/platform_schema.json.
4. behavior_key=%s
Do not wrap JSON in markdown fences.
`, capability, behavior)
}

func platformSchemaJSON(behavior string) ([]byte, error) {
	var schema any
	if strings.Contains(behavior, "topic_audit") {
		schema = map[string]any{
			"audit": map[string]any{
				"passed": true,
				"issues": []map[string]string{
					{"gate": "citation|claim|kill", "severity": "blocker|major|minor", "claim": "", "fix": ""},
				},
				"summary":   "",
				"off_topic": map[string]any{},
			},
		}
	} else {
		schema = map[string]any{
			"ideas": []map[string]any{
				{"title": "", "problem": "", "approach": "", "contribution": "", "reference_keys": []string{}},
			},
			"novelty": map[string]any{"lines": []any{}, "risks": []any{}, "synthesis": ""},
			"literature_review": map[string]any{
				"title": "", "summary": "", "structure": "thematic", "content_medium": "",
				"citations": []map[string]string{{"ref_id": "", "note": ""}},
			},
			"experiment_plan": nil,
		}
	}
	return json.MarshalIndent(schema, "", "  ")
}

func extractJSONObject(text string) (map[string]any, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("empty")
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(text), &m); err == nil {
		return m, nil
	}
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(text[start:end+1]), &m); err == nil {
			return m, nil
		}
	}
	return nil, fmt.Errorf("no json object")
}

func claudeResultText(payload map[string]any) string {
	if payload == nil {
		return ""
	}
	if s, ok := payload["result"].(string); ok {
		return s
	}
	var msg any
	if m, ok := payload["message"]; ok {
		msg = m
	} else {
		msg = payload["content"]
	}
	switch t := msg.(type) {
	case string:
		return t
	case []any:
		var parts []string
		for _, block := range t {
			bm, ok := block.(map[string]any)
			if !ok {
				continue
			}
			if bm["type"] == "text" {
				parts = append(parts, fmt.Sprint(bm["text"]))
			}
		}
		return strings.Join(parts, "\n")
	}
	return ""
}

func arisInvokeTimeout() time.Duration {
	if v := strings.TrimSpace(os.Getenv("PAPER_ARIS_INVOKE_TIMEOUT_SEC")); v != "" {
		if sec, err := parsePositiveInt(v); err == nil && sec > 0 {
			return time.Duration(sec) * time.Second
		}
	}
	return 3600 * time.Second
}

func runARISCLI(ctx context.Context, workspace, prompt string) (stdout, stderr string, err error) {
	executor := strings.ToLower(strings.TrimSpace(os.Getenv("PAPER_ARIS_EXECUTOR")))
	if executor == "" {
		executor = "claude"
	}
	cli := strings.TrimSpace(os.Getenv("PAPER_ARIS_CLI"))
	if cli == "" {
		if executor == "aris" {
			cli = "aris"
		} else {
			cli = "claude"
		}
	}

	env := os.Environ()
	if root := paperARISSkillsRoot(); root != "" {
		env = append(env, "CLAUDE_CODE_ADDITIONAL_DIRECTORIES="+root)
	}

	var args []string
	if executor == "aris" || filepath.Base(cli) == "aris" {
		args = []string{"--print", "--output-format", "json", prompt}
	} else {
		args = []string{"-p", prompt, "--output-format", "json"}
		if strings.TrimSpace(os.Getenv("PAPER_ARIS_CLAUDE_TOOLS")) == "0" {
			args = append(args, "--tools", "")
		}
	}

	cmd := exec.CommandContext(ctx, cli, args...)
	cmd.Dir = workspace
	cmd.Env = env
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errBuf.String())
		if msg == "" {
			msg = strings.TrimSpace(outBuf.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return outBuf.String(), errBuf.String(), fmt.Errorf("%s failed: %s", cli, msg)
	}
	return outBuf.String(), errBuf.String(), nil
}

func sanitizePathSegment(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "unknown"
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "unknown"
	}
	if len(out) > 64 {
		return out[:64]
	}
	return out
}

// arisInvokeSessionDir 每次 invoke 独立目录：user / manuscript / run / behavior+uuid，避免多用户与同用户并发互相踩文件。
func arisInvokeSessionDir(req StrategyInvokeRequest) (sessionDir string, cleanup func(), err error) {
	uid := req.Context.UserID
	mid := req.Context.ManuscriptID
	rv := req.Context.RunVersion
	if uid == 0 {
		return "", func() {}, fmt.Errorf("invoke context 缺少 user_id")
	}
	behavior := sanitizePathSegment(req.BehaviorKey)
	invokeID := uuid.NewString()
	base := storage.ArisInvokeWorkRoot()
	sessionDir = filepath.Join(
		base,
		fmt.Sprintf("user_%d", uid),
		fmt.Sprintf("manuscript_%d", mid),
		fmt.Sprintf("run_%d", rv),
		fmt.Sprintf("%s_%s", behavior, invokeID),
	)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return "", func() {}, err
	}
	keep := strings.TrimSpace(os.Getenv("PAPER_ARIS_KEEP_WORKSPACE")) == "1"
	if keep {
		return sessionDir, func() {}, nil
	}
	return sessionDir, func() { _ = os.RemoveAll(sessionDir) }, nil
}

func runLocalStrategyInvoke(ctx context.Context, req StrategyInvokeRequest, platformUser string) (*localARISInvokeResult, error) {
	behavior := strings.TrimSpace(req.BehaviorKey)
	capability := resolveCapability(req)
	if capability == "" {
		return nil, fmt.Errorf("无法解析 skill：capability_id=%q behavior_key=%q（请在 paper_llm_strategy_route 配置 ARIS skill 名，如 idea-creator）", req.CapabilityID, req.BehaviorKey)
	}

	sessionDir, cleanup, err := arisInvokeSessionDir(req)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	paperRoot := filepath.Join(sessionDir, "paper")
	invokeDir := filepath.Join(paperRoot, ".paper", "invoke")
	if err := os.MkdirAll(invokeDir, 0o755); err != nil {
		return nil, err
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(invokeDir, "request.json"), reqBytes, 0o644); err != nil {
		return nil, err
	}
	schemaBytes, err := platformSchemaJSON(behavior)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(invokeDir, "platform_schema.json"), schemaBytes, 0o644); err != nil {
		return nil, err
	}

	if len(req.Context.Corpus) > 0 && string(req.Context.Corpus) != "null" {
		corpusDir := filepath.Join(paperRoot, ".paper")
		if err := os.MkdirAll(corpusDir, 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(filepath.Join(corpusDir, "corpus.json"), req.Context.Corpus, 0o644); err != nil {
			return nil, err
		}
	}
	platformUser = strings.TrimSpace(platformUser)
	if platformUser != "" {
		if err := os.WriteFile(filepath.Join(invokeDir, "platform_user_prompt.txt"), []byte(platformUser), 0o644); err != nil {
			return nil, err
		}
	}

	runCtx, cancel := context.WithTimeout(ctx, arisInvokeTimeout())
	defer cancel()

	stdout, _, err := runARISCLI(runCtx, paperRoot, buildHeadlessPrompt(capability, behavior))
	if err != nil {
		if runCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("local ARIS/Claude invoke timeout")
		}
		return nil, err
	}

	var rawText string
	if cliJSON, _ := extractJSONObject(stdout); cliJSON != nil {
		rawText = claudeResultText(cliJSON)
	}
	if rawText == "" {
		rawText = stdout
	}

	resultObj, err := extractJSONObject(rawText)
	var resultBytes json.RawMessage
	if err != nil || len(resultObj) == 0 {
		trunc := rawText
		if len(trunc) > 8000 {
			trunc = trunc[:8000]
		}
		resultBytes, _ = json.Marshal(map[string]string{"raw": trunc})
	} else {
		resultBytes, _ = json.Marshal(resultObj)
	}

	trace := arisTraceSanitize.ReplaceAllString(capability, "")
	if len(trace) > 64 {
		trace = trace[:64]
	}
	rt := rawText
	if len(rt) > 16000 {
		rt = rt[:16000]
	}

	return &localARISInvokeResult{
		ResultJSON:      resultBytes,
		ExternalTraceID: trace,
		RawText:         rt,
	}, nil
}
