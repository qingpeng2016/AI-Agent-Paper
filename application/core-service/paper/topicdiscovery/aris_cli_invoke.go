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

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/llm"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/storage"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"go.uber.org/zap"
)

var arisTraceSanitize = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

type localARISInvokeResult struct {
	ResultJSON      json.RawMessage
	ExternalTraceID string
	RawText         string
	Usage           llm.LLMUsage
	CLIExecutor     string
	CLIModel        string
	ExecRecord      string
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
	if strings.Contains(behavior, "research_lit") {
		return "research-lit"
	}
	if strings.Contains(behavior, "idea_creator") {
		return "idea-creator"
	}
	if strings.Contains(behavior, "novelty_check") {
		return "novelty-check"
	}
	if strings.Contains(behavior, "research_review") || strings.Contains(behavior, "topic_audit") {
		return "research-review"
	}
	if strings.Contains(behavior, "research_refine") {
		return "research-refine-pipeline"
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

func buildHeadlessPrompt(capability, behavior, invokeDirRel string) string {
	capability = strings.TrimPrefix(strings.TrimSpace(capability), "/")
	invokeDirRel = strings.TrimSpace(filepath.ToSlash(invokeDirRel))
	if invokeDirRel == "" {
		invokeDirRel = ".paper/invoke"
	}
	// idea-discovery 走 buildIdeaDiscoveryCLIPrompt（与 ARIS 官方一致：/idea-discovery + $ARGUMENTS 同一 prompt）
	if capability == "idea-discovery" || strings.Contains(behavior, "discovery.run") {
		return ""
	}
	return fmt.Sprintf(`[AI-Agent-Paper headless invoke]
Execute research skill /%s in this workspace.

1. Read structured input: %s/request.json (context, options, behavior_key).
2. Follow the skill workflow; deliverables are Markdown per SKILL.md.
3. If .paper/corpus.json and literature/ exist, treat as optional local library per /research-lit (SKILL defaults).
4. Follow the skill workflow where practical, but **your final assistant message must be ONLY one valid JSON object**
   matching the platform schema described in %s/platform_schema.json.
5. behavior_key=%s
Do not wrap JSON in markdown fences.
`, capability, invokeDirRel, invokeDirRel, behavior)
}

// buildIdeaDiscoveryCLIPrompt 与 ARIS 文档一致：aris prompt '/idea-discovery <研究方向与开关…>'。
// sessionNotes 为平台检查点 PROCEED / resume 提示（等同用户在 chat 里回复），非空时接在参数行后。
func buildIdeaDiscoveryCLIPrompt(argsLine, sessionNotes string) string {
	argsLine = strings.TrimSpace(argsLine)
	var b strings.Builder
	b.WriteString("/idea-discovery")
	if argsLine != "" {
		b.WriteByte(' ')
		b.WriteString(argsLine)
	}
	sessionNotes = strings.TrimSpace(sessionNotes)
	if sessionNotes != "" {
		b.WriteString("\n\n")
		b.WriteString(sessionNotes)
	}
	return b.String()
}

func platformSchemaJSON(behavior string) ([]byte, error) {
	var schema any
	if strings.Contains(behavior, "discovery.run") {
		schema = map[string]any{
			"checkpoint_pause": false,
			"phase":            "",
			"run_id":           "",
			"summary":          "",
			"gate_passed":      false,
			"artifact_paths":   []string{},
			"blocked_reasons":  []string{},
		}
	} else if strings.Contains(behavior, "topic_audit") {
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

func parseLLMUsageFromCLIJSON(payload map[string]any) llm.LLMUsage {
	if payload == nil {
		return llm.LLMUsage{}
	}
	var usage llm.LLMUsage
	if u, ok := payload["usage"].(map[string]any); ok {
		usage.PromptTokens = intFromAnyUsage(u["prompt_tokens"], u["input_tokens"])
		usage.CompletionTokens = intFromAnyUsage(u["completion_tokens"], u["output_tokens"])
	}
	if usage.TotalTokens() == 0 {
		if u, ok := payload["metrics"].(map[string]any); ok {
			usage.PromptTokens = intFromAnyUsage(u["prompt_tokens"], u["input_tokens"])
			usage.CompletionTokens = intFromAnyUsage(u["completion_tokens"], u["output_tokens"])
		}
	}
	return usage
}

func intFromAnyUsage(primary, fallback any) int {
	if n := intFromAnyUsageVal(primary); n > 0 {
		return n
	}
	return intFromAnyUsageVal(fallback)
}

func intFromAnyUsageVal(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	default:
		return 0
	}
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
	// 默认不截断 CLI session（24h 兜底防泄漏）；可用 PAPER_ARIS_INVOKE_TIMEOUT_SEC 覆盖
	return 24 * time.Hour
}

// buildARISNativeCLIArgs 对齐 `aris --help`：非交互为 `aris [--model M] --output-format json prompt TEXT`。
func buildARISNativeCLIArgs(prompt string) []string {
	args := []string{"--output-format", "json"}
	if m := strings.TrimSpace(os.Getenv("PAPER_ARIS_CLI_MODEL")); m != "" {
		args = append(args, "--model", m)
	}
	args = append(args, "prompt", prompt)
	return args
}

// arisCLIExecutable 写入 exec_command 的可复制 CLI 名（与真实 invoke 一致）。
func arisCLIExecutable() string {
	if cli := strings.TrimSpace(os.Getenv("PAPER_ARIS_CLI")); cli != "" {
		return cli
	}
	return "aris"
}

// FormatARISInvokeShellCommand 完整 shell 单行：aris --output-format json prompt '…'。
func FormatARISInvokeShellCommand(cliPrompt string) string {
	cli, args := resolveARISCLIInvocation(cliPrompt)
	return formatARISCLIExecShellLine(cli, args)
}

func resolveARISCLIInvocation(prompt string) (cli string, args []string) {
	executor := strings.ToLower(strings.TrimSpace(os.Getenv("PAPER_ARIS_EXECUTOR")))
	if executor == "" {
		executor = "aris"
	}
	cli = strings.TrimSpace(os.Getenv("PAPER_ARIS_CLI"))
	if cli == "" {
		if executor == "aris" {
			cli = "aris"
		} else {
			cli = "claude"
		}
	}
	if executor == "aris" || filepath.Base(cli) == "aris" {
		args = buildARISNativeCLIArgs(prompt)
	} else {
		args = []string{"-p", prompt, "--output-format", "json"}
		if strings.TrimSpace(os.Getenv("PAPER_ARIS_CLAUDE_TOOLS")) == "0" {
			args = append(args, "--tools", "")
		}
	}
	return cli, args
}

func runARISCLI(ctx context.Context, workspace, prompt string) (stdout, stderr, execRecord string, err error) {
	cli, args := resolveARISCLIInvocation(prompt)
	execRecord = formatARISCLIExecRecord(workspace, cli, args, prompt)

	env := os.Environ()

	cmd := exec.CommandContext(ctx, cli, args...)
	cmd.Dir = workspace
	cmd.Env = env

	var logFile *os.File
	var logPath string
	if arisCLIWorkspaceLogEnabled() {
		if f, p, logErr := newARISCLILogFile(workspace, cli); logErr == nil {
			logFile = f
			logPath = p
			writeARISCLILogHeader(f, workspace, cli, args, prompt)
		} else {
			logger.WarnZ(ctx, "aris-cli-workspace-log-open", zap.Error(logErr), zap.String("workspace", workspace))
		}
	}
	if logFile != nil {
		defer func() {
			_ = logFile.Close()
			finalizeARISCLILog(workspace, logPath)
		}()
	}

	outStr, errStr, runErr := runARISCLIWithStream(ctx, cmd, logFile)
	if runErr != nil {
		msg := strings.TrimSpace(errStr)
		if msg == "" {
			msg = strings.TrimSpace(outStr)
		}
		if msg == "" {
			msg = runErr.Error()
		}
		logger.ErrorZ(ctx, "aris-cli-exit",
			zap.String("cli", cli),
			zap.String("workspace", workspace),
			zap.Error(runErr))
		return outStr, errStr, execRecord, fmt.Errorf("%s failed: %s", cli, msg)
	}
	return outStr, errStr, execRecord, nil
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
	sessionDir = arisRunSessionDir(uid, mid, rv)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return "", func() {}, err
	}
	return sessionDir, func() {}, nil
}

// arisRunSessionDir inout/user_*/manuscript_*/run_*。
func arisRunSessionDir(userID uint64, manuscriptID uint64, runVersion int) string {
	base := storage.ArisInvokeWorkRoot()
	return filepath.Join(
		base,
		fmt.Sprintf("user_%d", userID),
		fmt.Sprintf("manuscript_%d", manuscriptID),
		fmt.Sprintf("run_%d", runVersion),
	)
}

// ARISRunPaperRoot 本 run 的 ARIS 工作区根（…/run_N/paper/）。
func ARISRunPaperRoot(userID uint64, manuscriptID uint64, runVersion int) string {
	return filepath.Join(arisRunSessionDir(userID, manuscriptID, runVersion), arisRunWorkspaceDir)
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

	ideaCfg := mustIdeaDiscoveryConfig()
	paperRoot := filepath.Join(sessionDir, arisRunWorkspaceDir)
	var discoveryInput topicRunInput
	if strings.Contains(behavior, "discovery.run") {
		discoveryInput = topicRunInput{
			Direction:   req.Context.Direction,
			Description: req.Context.Direction,
			Keywords:    req.Context.Keywords,
			Venue:       req.Context.VenueLabel,
		}
		if err := PrepareNativeIdeaDiscoveryWorkspace(paperRoot, discoveryInput, req.Context.Language, ideaCfg); err != nil {
			return nil, err
		}
		if err := writeIdeaDiscoveryContractFile(paperRoot, ideaCfg); err != nil {
			return nil, err
		}
	}
	invokeDir := ideaCfg.InvokeDirAbs(paperRoot)
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

	platformUser = strings.TrimSpace(platformUser)
	promptPath := filepath.Join(invokeDir, "platform_user_prompt.txt")
	if strings.Contains(behavior, "discovery.run") {
		runID := newestIdeaDiscoveryRunStateID(paperRoot)
		if runID == "" {
			runID = resolveIdeaDiscoveryRunID(paperRoot, nil, discoveryInput)
		}
		withResume := ideaCfg.ShouldAppendCLIResume(paperRoot, runID)
		argsLine := ideaCfg.FormatIdeaDiscoveryArgsLine(discoveryInput, req.Context.VenueLabel, req.Context.Language, runID, withResume)
		if err := os.WriteFile(promptPath, []byte(argsLine), 0o644); err != nil {
			return nil, err
		}
		if platformUser != "" {
			_ = os.WriteFile(filepath.Join(invokeDir, "platform_session_notes.txt"), []byte(platformUser), 0o644)
		}
	} else if platformUser != "" {
		if err := os.WriteFile(filepath.Join(invokeDir, "platform_user_prompt.txt"), []byte(platformUser), 0o644); err != nil {
			return nil, err
		}
	}

	runCtx, cancel := context.WithTimeout(ctx, arisInvokeTimeout())
	defer cancel()

	cliPrompt := buildHeadlessPrompt(capability, behavior, ideaCfg.invokeDirRel())
	if strings.Contains(behavior, "discovery.run") {
		argsLine := ""
		if b, err := os.ReadFile(promptPath); err == nil {
			argsLine = strings.TrimSpace(string(b))
		}
		cliPrompt = buildIdeaDiscoveryCLIPrompt(argsLine, "")
	}
	if strings.TrimSpace(cliPrompt) == "" {
		return nil, fmt.Errorf("empty ARIS CLI prompt for behavior_key=%q", behavior)
	}
	stdout, _, execRecord, err := runARISCLI(runCtx, paperRoot, cliPrompt)
	if err != nil {
		if runCtx.Err() == context.DeadlineExceeded {
			return &localARISInvokeResult{ExecRecord: execRecord}, fmt.Errorf("local ARIS/Claude invoke timeout")
		}
		return &localARISInvokeResult{ExecRecord: execRecord}, err
	}

	var rawText string
	var cliUsage llm.LLMUsage
	if cliJSON, _ := extractJSONObject(stdout); cliJSON != nil {
		rawText = claudeResultText(cliJSON)
		cliUsage = parseLLMUsageFromCLIJSON(cliJSON)
	}
	if rawText == "" {
		rawText = stdout
	}
	execLabel, cliModel := arisCLIExecutorModel()

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
		Usage:           cliUsage,
		CLIExecutor:     execLabel,
		CLIModel:        cliModel,
		ExecRecord:      execRecord,
	}, nil
}

func arisCLIExecutorModel() (executor, model string) {
	executor = strings.ToLower(strings.TrimSpace(os.Getenv("PAPER_ARIS_EXECUTOR")))
	if executor == "" {
		executor = "claude"
	}
	model = strings.TrimSpace(os.Getenv("PAPER_ARIS_CLI_MODEL"))
	if model == "" {
		model = "claude-cli"
		if executor == "codex" || executor == "aris" {
			model = "codex-cli"
		}
	}
	return executor, model
}
