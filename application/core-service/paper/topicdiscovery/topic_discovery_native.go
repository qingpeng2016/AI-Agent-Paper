package topicdiscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/internal/jsonx"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/lang"
	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"gopkg.in/yaml.v3"
)

func mustIdeaDiscoveryConfig() ARISIdeaDiscoveryConfig {
	cfg, err := LoadARISIdeaDiscoveryConfig()
	if err != nil {
		return defaultARISIdeaDiscoveryConfig()
	}
	return cfg
}

// RunNativeIdeaDiscoveryLLM 原生 /idea-discovery（契约见 conf/aris_idea_discovery.yaml）。
func (s *TopicDiscoveryRunService) RunNativeIdeaDiscoveryLLM(
	ctx context.Context,
	userID uint,
	runVersion int,
	step *entity.PaperOutputTopicStep,
	input topicRunInput,
) error {
	cfg := mustIdeaDiscoveryConfig()

	route, err := s.requireTopicStrategyRoute(ctx, BehaviorIdeaDiscoveryRun)
	if err != nil {
		return err
	}

	paperRoot := ARISRunPaperRoot(uint64(userID), step.ManuscriptID, runVersion)
	runID := resolveIdeaDiscoveryRunID(paperRoot, step, input)

	invokeCtx := s.buildTopicStrategyInvokeContext(ctx, userID, runVersion, step, input, 1, 1, nil, nil)
	maxChain := cfg.Platform.MaxArisPerBot
	chainInTick := 0

	for {
		pendingPhase, err := ideaDiscoveryResumePhase(paperRoot, runID)
		if err != nil {
			return fmt.Errorf("run resume: %w", err)
		}
		if pendingPhase == "" {
			return s.finishIdeaDiscoveryAfterGate(ctx, step, userID, runVersion, route, paperRoot, runID, cfg)
		}

		meta := paperstep.LoadStepExtra(step)
		delete(meta, "aris_awaiting_checkpoint")
		meta["async_llm"] = true
		meta["aris_run_id"] = runID
		meta["native_run_id"] = runID
		step.Result = jsonx.MustJSON(meta)
		step.ExecProgressPct = ideaDiscoveryProgressPct(paperRoot, runID)
		_ = s.steps.SaveStep(ctx, step)

		syncIdeaDiscoveryRunProgressToMeta(meta, paperRoot, runID)
		// resume 与 auto_proceed 无关；CLI「— resume」由 conf platform.resume 控制。
		resuming := cfg.ShouldAppendCLIResume(paperRoot, runID)
		locale := invokeCtx.Language
		if err := PrepareNativeIdeaDiscoveryWorkspace(paperRoot, input, locale, cfg); err != nil {
			return err
		}
		argsLine := cfg.formatSkillArgumentsLine(paperRoot, compactDirectionForPrompt(input), strings.TrimSpace(input.Venue), locale)
		if resuming {
			argsLine += fmt.Sprintf(` — resume %s`, runID)
		}
		proceedPhase := ""
		if meta != nil {
			argsLine, proceedPhase = applyCheckpointProceedToArgsLine(meta, argsLine, pendingPhase)
			if proceedPhase != "" {
				step.Result = jsonx.MustJSON(meta)
				_ = s.steps.SaveStep(ctx, step)
			}
		}
		if err := writeNativePlatformArguments(paperRoot, argsLine); err != nil {
			return err
		}
		userExtra := buildNativeIdeaDiscoverySessionNotes(cfg, runID, pendingPhase, resuming, proceedPhase)
		req := BuildTopicDiscoveryInvokeRequest(BehaviorIdeaDiscoveryRun, route.CapabilityID, input, invokeCtx)

		raw, usage, err := s.invokeTopicDiscoveryBehavior(ctx, step, req, userExtra)
		if err != nil {
			return s.handleIdeaDiscoveryInvokeFailure(ctx, step, paperRoot, cfg, err)
		}

		meta = paperstep.LoadStepExtra(step)
		meta["aris_invoke_count"] = ideaDiscoveryInvokeCount(meta) + 1
		meta["aris_run_id"] = runID
		syncIdeaDiscoveryRunProgressToMeta(meta, paperRoot, runID)
		// 完整 CLI 输出见 paper_llm_call_logs；step.result 只保留摘要，避免 JSON 过大 + 字符集问题
		meta["last_invoke_preview"] = truncateUTF8(string(raw), 4096)
		meta["tokens_prompt"] = usage.PromptTokens
		meta["tokens_completion"] = usage.CompletionTokens
		delete(meta, "aris_invoke_last_error")
		delete(meta, "aris_invoke_retries")
		step.Result = jsonx.MustJSON(meta)

		pendingPhase, err = ideaDiscoveryResumePhase(paperRoot, runID)
		if err != nil {
			return err
		}
		if pendingPhase == "" {
			return s.finishIdeaDiscoveryAfterGate(ctx, step, userID, runVersion, route, paperRoot, runID, cfg)
		}
		if isResearchLitPhaseDone(paperRoot, runID) {
			_, _ = s.ingestARISWorkspaceLiterature(ctx, userID, runVersion, step.ManuscriptID, paperRoot, cfg)
		}

		if !cfg.Platform.AutoProceed {
			meta = paperstep.LoadStepExtra(step)
			syncIdeaDiscoveryRunProgressToMeta(meta, paperRoot, runID)
			meta["aris_awaiting_checkpoint"] = true
			meta["aris_pause_after_phase"] = pendingPhase
			delete(meta, "async_llm")
			meta["async_llm_done"] = false
			step.Result = jsonx.MustJSON(meta)
			step.Status = "running"
			step.ExecProgressPct = ideaDiscoveryProgressPct(paperRoot, runID)
			return s.steps.SaveStep(ctx, step)
		}

		chainInTick++
		if chainInTick >= maxChain {
			meta = paperstep.LoadStepExtra(step)
			meta["async_llm"] = true
			delete(meta, "aris_awaiting_checkpoint")
			step.Result = jsonx.MustJSON(meta)
			step.Status = "running"
			step.ExecProgressPct = ideaDiscoveryProgressPct(paperRoot, runID)
			return s.steps.SaveStep(ctx, step)
		}
	}
}

func ideaDiscoveryInvokeCount(meta map[string]any) int {
	if n, ok := meta["aris_invoke_count"].(float64); ok {
		return int(n)
	}
	if n, ok := meta["aris_invoke_count"].(int); ok {
		return n
	}
	return 0
}

func (s *TopicDiscoveryRunService) handleIdeaDiscoveryInvokeFailure(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	paperRoot string,
	cfg ARISIdeaDiscoveryConfig,
	invokeErr error,
) error {
	meta := paperstep.LoadStepExtra(step)
	retries := 0
	if n, ok := meta["aris_invoke_retries"].(float64); ok {
		retries = int(n)
	}
	retries++
	meta["aris_invoke_retries"] = retries
	meta["aris_invoke_last_error"] = invokeErr.Error()
	appendARISInvokeErrorLog(paperRoot, invokeErr, retries)

	retriable := retries <= ideaDiscoveryMaxInvokeRetries && isRetriableIdeaDiscoveryInvokeError(invokeErr)
	if retriable {
		meta["async_llm"] = true
		delete(meta, "aris_awaiting_checkpoint")
		step.Result = jsonx.MustJSON(meta)
		step.Status = "running"
		step.CompletedAt = nil
		_ = s.steps.SaveStep(ctx, step)
		return nil
	}
	reason := strings.TrimSpace(invokeErr.Error())
	if retries > ideaDiscoveryMaxInvokeRetries {
		reason = fmt.Sprintf("Idea Discovery CLI 已达最大重试（%d 次）: %s", ideaDiscoveryMaxInvokeRetries, reason)
	}
	paperstep.FailTopicStep(step, reason)
	step.ExecProgressPct = ideaDiscoveryProgressPct(paperRoot, strings.TrimSpace(fmt.Sprint(meta["aris_run_id"])))
	_ = s.steps.SaveStep(ctx, step)
	return invokeErr
}

func isRetriableIdeaDiscoveryInvokeError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if isNonRetriableIdeaDiscoveryInvokeError(msg) {
		return false
	}
	return strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "deadline") ||
		strings.Contains(msg, "killed") ||
		strings.Contains(msg, "context deadline exceeded")
}

func isNonRetriableIdeaDiscoveryInvokeError(msg string) bool {
	needles := []string{
		"apikeyhelper",
		"api key",
		"api_key",
		"unauthorized",
		"authentication",
		"not logged in",
		"permission denied",
		"无法解析 skill",
		"retrieve 未完成",
		"aris failed",
	}
	for _, n := range needles {
		if strings.Contains(msg, n) {
			return true
		}
	}
	return false
}

func (s *TopicDiscoveryRunService) finishIdeaDiscoveryAfterGate(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	userID uint,
	runVersion int,
	route *entity.PaperStrategyRoute,
	paperRoot, runID string,
	cfg ARISIdeaDiscoveryConfig,
) error {
	gate, gateErr := runIdeaDiscoveryEvidenceGate(paperRoot, runID, cfg.Artifacts.IdeaReport)
	if strings.TrimSpace(gate.ExecRecord) != "" {
		s.persistTopicStepExecCommand(ctx, step, gate.ExecRecord)
	}
	result := map[string]any{
		"pipeline":        "aris_idea_discovery_native",
		"behavior_key":    BehaviorIdeaDiscoveryRun,
		"capability_id":   route.CapabilityID,
		"aris_run_id":     runID,
		"gate_passed":     gate.Passed,
		"gate_stdout":     gate.Stdout,
		"gate_blocked":    gate.Reasons,
		"artifact_config": cfg.artifactConfigMap(),
	}
	mergeNativeIdeaDiscoveryArtifacts(result, paperRoot, cfg)
	aggregate, _ := s.BuildIdeaDiscoveryAggregateResult(ctx, userID, runVersion)
	for k, v := range aggregate {
		result[k] = v
	}
	step.Result = jsonx.MustJSON(result)

	meta := paperstep.LoadStepExtra(step)
	meta["strategy_route"] = strategyRouteMeta(route)
	meta["pipeline"] = "aris_idea_discovery_native"
	meta["aris_run_id"] = runID
	meta["gate_passed"] = gate.Passed
	delete(meta, "async_llm")
	delete(meta, "aris_awaiting_checkpoint")
	delete(meta, "aris_invoke_retries")
	meta["async_llm_done"] = true
	delete(meta, "awaiting_user_confirm")
	step.Result = jsonx.MustJSON(meta)

	done := time.Now()
	_, _ = s.ingestARISWorkspaceLiterature(ctx, userID, runVersion, step.ManuscriptID, paperRoot, cfg)
	if !gate.Passed {
		reason := "idea_discovery_gate 未 PASS"
		if gateErr != nil {
			reason = gateErr.Error()
		}
		paperstep.FailTopicStep(step, reason)
		step.ExecProgressPct = ideaDiscoveryProgressPct(paperRoot, runID)
		step.CompletedAt = &done
		if err := s.steps.SaveStep(ctx, step); err != nil {
			return err
		}
		if gateErr != nil {
			return gateErr
		}
		return fmt.Errorf("idea_discovery_gate 未 PASS")
	}
	step.Status = "completed"
	step.ExecProgressPct = 100
	step.CompletedAt = &done
	if err := s.steps.SaveStep(ctx, step); err != nil {
		return err
	}
	return nil
}

func ideaDiscoveryProgressPct(paperRoot, runID string) uint8 {
	statePath := mustIdeaDiscoveryConfig().RunStateFileAbs(paperRoot, runID)
	b, err := os.ReadFile(statePath)
	if err != nil {
		return 1
	}
	var st struct {
		Phases []struct {
			Status string `json:"status"`
		} `json:"phases"`
	}
	if json.Unmarshal(b, &st) != nil || len(st.Phases) == 0 {
		return 1
	}
	done := 0
	for _, p := range st.Phases {
		switch strings.TrimSpace(p.Status) {
		case "done", "accepted", "provisional":
			done++
		}
	}
	return uint8(min(99, max(1, done*100/len(st.Phases))))
}

// buildNativeIdeaDiscoverySessionNotes run_state 续跑 + 人工检查点 PROCEED；$ARGUMENTS 在 platform_user_prompt.txt。
func buildNativeIdeaDiscoverySessionNotes(cfg ARISIdeaDiscoveryConfig, runID, resumePhase string, resuming bool, checkpointProceedPhase string) string {
	checkpointProceedPhase = strings.TrimSpace(checkpointProceedPhase)
	var b strings.Builder
	fmt.Fprintf(&b, "run_id=%s", runID)
	if resumePhase != "" || resuming {
		fmt.Fprintf(&b, " resume_phase=%s resuming=%v", resumePhase, resuming)
	}
	if checkpointProceedPhase != "" {
		fmt.Fprintf(&b, "\n%s", checkpointProceedSessionNote(checkpointProceedPhase))
	}
	return b.String()
}

func checkpointProceedSessionNote(phase string) string {
	phase = strings.TrimSpace(phase)
	if phase == "" {
		phase = "research-lit"
	}
	return fmt.Sprintf(
		`Platform user checkpoint reply: PROCEED (confirm and continue). Under AUTO_PROCEED=false this is explicit approval to execute phase %q now — run the skill work (e.g. /research-lit search, verify, optional arxiv PDF download per contract). Do not stop at a pre-execution staging checkpoint or re-ask the same decision.`,
		phase,
	)
}

func applyCheckpointProceedToArgsLine(meta map[string]any, argsLine, defaultPhase string) (string, string) {
	if meta == nil {
		return argsLine, ""
	}
	v, ok := meta["aris_checkpoint_proceed"].(bool)
	if !ok || !v {
		return argsLine, ""
	}
	phase := strings.TrimSpace(fmt.Sprint(meta["aris_checkpoint_proceed_phase"]))
	if phase == "" {
		phase = strings.TrimSpace(defaultPhase)
	}
	delete(meta, "aris_checkpoint_proceed")
	delete(meta, "aris_checkpoint_proceed_phase")
	suffix := fmt.Sprintf(
		` — user checkpoint: PROCEED, execute phase %q now (do not re-stage; run /research-lit including search and ARXIV_DOWNLOAD per contract)`,
		phase,
	)
	return argsLine + suffix, phase
}

func mergeNativeIdeaDiscoveryArtifacts(result map[string]any, paperRoot string, cfg ARISIdeaDiscoveryConfig) {
	var found []string
	for _, rel := range cfg.artifactPathsForResult() {
		if st, err := os.Stat(filepath.Join(paperRoot, rel)); err == nil && st.Size() > 0 {
			found = append(found, rel)
		}
	}
	if len(found) > 0 {
		result["artifact_paths"] = found
	}
}

func writeNativePlatformArguments(paperRoot, argsLine string) error {
	cfg := mustIdeaDiscoveryConfig()
	invokeDir := cfg.InvokeDirAbs(paperRoot)
	if err := os.MkdirAll(invokeDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(invokeDir, "platform_user_prompt.txt"), []byte(strings.TrimSpace(argsLine)), 0o644)
}

func writeIdeaDiscoveryContractFile(paperRoot string, cfg ARISIdeaDiscoveryConfig) error {
	invokeDir := cfg.InvokeDirAbs(paperRoot)
	if err := os.MkdirAll(invokeDir, 0o755); err != nil {
		return err
	}
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(invokeDir, "idea_discovery_contract.yaml"), b, 0o644)
}

func PrepareNativeIdeaDiscoveryWorkspace(paperRoot string, input topicRunInput, locale string, cfg ARISIdeaDiscoveryConfig) error {
	if err := os.MkdirAll(paperRoot, 0o755); err != nil {
		return err
	}
	locale = lang.NormalizeContentLanguage(locale)
	dir := compactDirectionForPrompt(input)
	kw := keywordsLine(input)
	venue := strings.TrimSpace(input.Venue)
	var b strings.Builder
	fmt.Fprintf(&b, "# Research Brief\n\n")
	fmt.Fprintf(&b, "## Language\n%s\n\n", lang.DeliverableLanguageHint(locale))
	fmt.Fprintf(&b, "## Direction\n%s\n\n", dir)
	if kw != "" {
		fmt.Fprintf(&b, "## Keywords\n%s\n\n", kw)
	}
	if venue != "" {
		fmt.Fprintf(&b, "## Target venue\n%s\n\n", venue)
	}
	briefRel := strings.TrimSpace(cfg.Artifacts.ResearchBrief)
	if briefRel == "" {
		briefRel = "RESEARCH_BRIEF.md"
	}
	return os.WriteFile(filepath.Join(paperRoot, briefRel), []byte(b.String()), 0o644)
}

func (s *TopicDiscoveryRunService) BuildIdeaDiscoveryAggregateResult(ctx context.Context, userID uint, runVersion int) (map[string]any, error) {
	cfg := mustIdeaDiscoveryConfig()
	st, _ := s.steps.GetStep(ctx, uint64(userID), runVersion, "idea_discovery")
	mid := uint64(0)
	if st != nil {
		mid = st.ManuscriptID
	}
	paperRoot := ARISRunPaperRoot(uint64(userID), mid, runVersion)
	finalProposal := readFileMax(filepath.Join(paperRoot, cfg.Artifacts.FinalProposal), 512000)
	expPlan := readFileMax(filepath.Join(paperRoot, cfg.Artifacts.ExperimentPlan), 512000)
	ideaReport := readFileMax(filepath.Join(paperRoot, cfg.Artifacts.IdeaReport), 512000)

	litBody := finalProposal
	if strings.TrimSpace(litBody) == "" {
		litBody = ideaReport
	}
	generate := map[string]any{
		"ideas":             []any{},
		"novelty":           map[string]any{"synthesis": ideaReport},
		"literature_review": map[string]any{"title": "Final proposal", "content_medium": litBody},
		"experiment_plan":   map[string]any{"summary": expPlan, "content_medium": expPlan},
	}
	audit := map[string]any{"passed": true, "summary": "native idea-discovery + gate"}
	iterations := []map[string]any{{
		"round": 1, "generate": generate, "audit": audit, "passed": true,
	}}
	return map[string]any{
		"passed": true, "iterations": iterations,
		"ideas": generate["ideas"], "novelty": generate["novelty"],
		"literature_review": generate["literature_review"],
		"experiment_plan": generate["experiment_plan"],
		"final_audit": audit, "pipeline": "aris_idea_discovery_native",
	}, nil
}

func readFileMax(abs string, max int) string {
	b, err := os.ReadFile(abs)
	if err != nil || len(b) == 0 {
		return ""
	}
	if len(b) > max {
		b = b[:max]
	}
	return string(b)
}

// ResumeIdeaDiscoveryFromCheckpoint 用户 continue：Phase 内 AUTO_PROCEED=false 检查点。
func (s *TopicDiscoveryRunService) ResumeIdeaDiscoveryFromCheckpoint(ctx context.Context, userID uint, runVersion int) (bool, error) {
	step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "idea_discovery")
	if err != nil || step == nil {
		return false, err
	}
	meta := paperstep.LoadStepExtra(step)
	if v, ok := meta["aris_awaiting_checkpoint"].(bool); !ok || !v {
		return false, nil
	}
	delete(meta, "aris_awaiting_checkpoint")
	phase := strings.TrimSpace(fmt.Sprint(meta["aris_pause_after_phase"]))
	if phase == "" {
		phase = "research-lit"
	}
	meta["aris_checkpoint_proceed"] = true
	meta["aris_checkpoint_proceed_phase"] = phase
	meta["async_llm"] = true
	step.Result = jsonx.MustJSON(meta)
	step.Status = "running"
	step.CompletedAt = nil
	if err := s.steps.SaveStep(ctx, step); err != nil {
		return false, err
	}
	return true, nil
}

// ResumeIdeaDiscoveryExecution 原生 `-- resume <run_id>`：重新排队 bot（含超时/CLI 中断后续跑）。
func (s *TopicDiscoveryRunService) ResumeIdeaDiscoveryExecution(ctx context.Context, userID uint, runVersion int) (bool, error) {
	step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "idea_discovery")
	if err != nil || step == nil {
		return false, err
	}
	if strings.TrimSpace(step.Status) == "completed" || len(step.Result) > 0 {
		return false, nil
	}
	meta := paperstep.LoadStepExtra(step)
	if strings.TrimSpace(step.Status) == "failed" {
		if strings.TrimSpace(fmt.Sprint(meta["aris_run_id"])) == "" {
			return false, nil
		}
		step.Result = nil
	}
	meta["async_llm"] = true
	delete(meta, "aris_invoke_last_error")
	delete(meta, "error")
	step.Result = jsonx.MustJSON(meta)
	step.Status = "running"
	step.CompletedAt = nil
	if err := s.steps.SaveStep(ctx, step); err != nil {
		return false, err
	}
	return true, nil
}
