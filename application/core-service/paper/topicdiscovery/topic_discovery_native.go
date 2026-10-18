package topicdiscovery

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

// RunNativeIdeaDiscoveryLLM 原生 /idea-discovery：bot 经 PrepareIdeaDiscoveryBotInvoke 放行后，本函数只打一条 aris 并落库。
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
	invokeCtx := s.buildTopicStrategyInvokeContext(ctx, userID, runVersion, step, input, 1, 1, nil, nil)

	runID := resolveIdeaDiscoveryRunID(paperRoot, step, input)
	pendingPhase, err := ideaDiscoveryResumePhase(paperRoot, runID)
	if err != nil {
		return fmt.Errorf("run resume: %w", err)
	}
	if pendingPhase == "" {
		return s.finishIdeaDiscoveryAfterGate(ctx, step, userID, runVersion, route, paperRoot, runID, cfg)
	}
	meta := paperstep.LoadStepExtra(step)
	normalizeIdeaDiscoveryRuntimeMeta(meta)
	meta["aris_run_id"] = runID
	meta["native_run_id"] = runID
	paperstep.PersistStepExtra(step, meta)
	step.ExecProgressPct = ideaDiscoveryProgressPct(paperRoot, runID)
	_ = s.steps.SaveStep(ctx, step)

	syncIdeaDiscoveryRunProgressToMeta(meta, paperRoot, runID)
	// 与手动一致：每次只打一条命令；续跑/重试也先清会话，避免 bot 连 tick 堆上下文到 200k。
	cfg.ClearCLIWorkspaceSessions(paperRoot)
	runID = resolveIdeaDiscoveryRunID(paperRoot, step, input)
	locale := invokeCtx.Language
	if err := PrepareNativeIdeaDiscoveryWorkspace(paperRoot, input, locale, cfg); err != nil {
		return err
	}
	argsLine := cfg.FormatIdeaDiscoveryArgsLine(input, strings.TrimSpace(input.Venue), locale)
	if err := writeNativePlatformArguments(paperRoot, argsLine); err != nil {
		return err
	}
	req := BuildTopicDiscoveryInvokeRequest(BehaviorIdeaDiscoveryRun, route.CapabilityID, input, invokeCtx)

	raw, usage, err := s.invokeTopicDiscoveryBehavior(ctx, step, req, "")
	if err != nil {
		return s.handleIdeaDiscoveryInvokeFailure(ctx, step, paperRoot, cfg, err)
	}

	meta = paperstep.LoadStepExtra(step)
	meta["aris_invoke_count"] = ideaDiscoveryInvokeCount(meta) + 1
	meta["aris_run_id"] = runID
	syncIdeaDiscoveryRunProgressToMeta(meta, paperRoot, runID)
	meta["last_invoke_preview"] = truncateUTF8(string(raw), 4096)
	meta["tokens_prompt"] = usage.PromptTokens
	meta["tokens_completion"] = usage.CompletionTokens
	delete(meta, "aris_invoke_last_error")
	delete(meta, "aris_invoke_retries")
	paperstep.PersistStepExtra(step, meta)

	runID = resolveIdeaDiscoveryRunID(paperRoot, step, input)
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

	pendingPhase, err = ideaDiscoveryResumePhase(paperRoot, runID)
	if err != nil {
		return err
	}
	if pendingPhase == "" {
		return s.finishIdeaDiscoveryAfterGate(ctx, step, userID, runVersion, route, paperRoot, runID, cfg)
	}
	meta = paperstep.LoadStepExtra(step)
	normalizeIdeaDiscoveryRuntimeMeta(meta)
	syncIdeaDiscoveryRunProgressToMeta(meta, paperRoot, runID)

	// AUTO_PROCEED=true：run_state 已存在且该段未 done 时可 drain（同段内 bot 连 tick）。
	// AUTO_PROCEED=false：skill 内 checkpoint 需真人/前端回复，禁止 bot 自动重打 /idea-discovery。
	if cfg.Platform.AutoProceed &&
		ideaDiscoveryRunStateFileExists(paperRoot, runID) &&
		IdeaDiscoveryBotMayInvokePending(meta, pendingPhase, paperRoot, runID) {
		meta["async_llm"] = true
		meta["async_llm_done"] = false
		paperstep.PersistStepExtra(step, meta)
		step.Status = "running"
		step.ExecProgressPct = ideaDiscoveryProgressPct(paperRoot, runID)
		return s.steps.SaveStep(ctx, step)
	}

	if ideaDiscoveryNeedsUserConfirm(meta, pendingPhase, cfg) ||
		(!cfg.Platform.AutoProceed && !isRunStatePhaseDone(paperRoot, runID, pendingPhase)) {
		return s.saveIdeaDiscoveryAwaitingUserContinue(ctx, step, paperRoot, runID, pendingPhase)
	}
	delete(meta, "async_llm")
	return s.saveIdeaDiscoveryAfterInvokeIdle(ctx, step, paperRoot, runID)
}

func (s *TopicDiscoveryRunService) saveIdeaDiscoveryAfterInvokeIdle(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	paperRoot, runID string,
) error {
	meta := paperstep.LoadStepExtra(step)
	syncIdeaDiscoveryRunProgressToMeta(meta, paperRoot, runID)
	delete(meta, "async_llm")
	meta["async_llm_done"] = false
	normalizeIdeaDiscoveryRuntimeMeta(meta)
	if ideaDiscoveryRunStateFileExists(paperRoot, runID) {
		meta["aris_run_state_ready"] = true
	} else {
		meta["aris_run_state_ready"] = false
	}
	paperstep.PersistStepExtra(step, meta)
	step.Status = "running"
	step.ExecProgressPct = ideaDiscoveryProgressPct(paperRoot, runID)
	return s.steps.SaveStep(ctx, step)
}

func (s *TopicDiscoveryRunService) saveIdeaDiscoveryAwaitingUserContinue(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	paperRoot, runID, pendingPhase string,
) error {
	meta := paperstep.LoadStepExtra(step)
	syncIdeaDiscoveryRunProgressToMeta(meta, paperRoot, runID)
	setPhaseStatusInMeta(meta, pendingPhase, IdeaDiscoveryPhaseAwaitingConfirm)
	normalizeIdeaDiscoveryRuntimeMeta(meta)
	delete(meta, "async_llm")
	meta["async_llm_done"] = false
	paperstep.PersistStepExtra(step, meta)
	step.Status = "running"
	step.ExecProgressPct = ideaDiscoveryProgressPct(paperRoot, runID)
	return s.steps.SaveStep(ctx, step)
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
	if StepUsesRuntimeStagePhasedInvoke(step) || strings.TrimSpace(fmt.Sprint(meta["idea_discovery_invoke_mode"])) == "phased" {
		phase := executingIdeaDiscoveryPhaseFromMeta(meta)
		reason := ideaDiscoveryPhasedStepFailureReason(meta, phase, invokeErr)
		if err := s.persistIdeaDiscoveryPhasedStepFailed(ctx, step, reason); err != nil {
			return err
		}
		return invokeErr
	}
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
		normalizeIdeaDiscoveryRuntimeMeta(meta)
		paperstep.PersistStepExtra(step, meta)
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
		"prompt is too long",
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
	meta := paperstep.LoadStepExtra(step)
	for k, v := range result {
		meta[k] = v
	}
	meta["strategy_route"] = strategyRouteMeta(route)
	delete(meta, "async_llm")
	normalizeIdeaDiscoveryRuntimeMeta(meta)
	delete(meta, "aris_invoke_retries")
	meta["async_llm_done"] = true
	delete(meta, "awaiting_user_confirm")
	paperstep.PersistStepExtra(step, meta)

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
	return ideaDiscoveryProgressPctByUIGroups(paperRoot, runID, mustIdeaDiscoveryConfig())
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

// ReconcileIdeaDiscoveryBotAwaiting 把旧版 aris_* 运行态迁移为 runtime_state 内 {"research-lit": true} 形式。
func (s *TopicDiscoveryRunService) ReconcileIdeaDiscoveryBotAwaiting(
	ctx context.Context,
	userID uint,
	runVersion int,
	step *entity.PaperOutputTopicStep,
	input topicRunInput,
) error {
	if step == nil {
		return nil
	}
	meta := paperstep.LoadStepExtra(step)
	if !hasLegacyIdeaDiscoveryRuntimeKeys(meta) {
		return nil
	}
	normalizeIdeaDiscoveryRuntimeMeta(meta)
	paperstep.PersistStepExtra(step, meta)
	return s.steps.SaveStep(ctx, step)
}

// ResumeIdeaDiscoveryFromCheckpoint 用户点「确认并继续」：runtime_state 里当前 pending 段置 true，并 async_llm 排队 bot。
func (s *TopicDiscoveryRunService) ResumeIdeaDiscoveryFromCheckpoint(ctx context.Context, userID uint, runVersion int) (bool, error) {
	step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "idea_discovery")
	if err != nil || step == nil {
		return false, err
	}
	if strings.TrimSpace(step.Status) == "completed" {
		return false, nil
	}
	input, err := s.loadRunInput(ctx, userID, runVersion)
	if err != nil {
		return false, err
	}
	paperRoot := ARISRunPaperRoot(uint64(userID), step.ManuscriptID, runVersion)
	runID := resolveIdeaDiscoveryRunID(paperRoot, step, input)
	meta := paperstep.LoadStepExtra(step)
	normalizeIdeaDiscoveryRuntimeMeta(meta)

	pending, err := ideaDiscoveryResumePhase(paperRoot, runID)
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(pending) == "" {
		pending = pendingPhaseFromMeta(meta)
	}
	if strings.TrimSpace(pending) == "" {
		return false, nil
	}
	setPhaseStatusInMeta(meta, pending, IdeaDiscoveryPhaseExecuting)
	meta["async_llm"] = true
	paperstep.PersistStepExtra(step, meta)
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
	paperstep.PersistStepExtra(step, meta)
	step.Status = "running"
	step.CompletedAt = nil
	if err := s.steps.SaveStep(ctx, step); err != nil {
		return false, err
	}
	return true, nil
}
