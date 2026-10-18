package topicdiscovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const BehaviorIdeaDiscoveryPhase = "topic.discovery.phase"

// researchLitLandscapeRel 分阶段 bot：/research-lit 段以该文件为唯一成功产物（paperRoot 下）。
const researchLitLandscapeRel = "LITERATURE_LANDSCAPE.md"

func researchLitLandscapePresent(paperRoot string) bool {
	return workspaceFileMinSize(paperRoot, researchLitLandscapeRel, 64)
}

// buildIdeaDiscoveryPhaseCLIPrompt 单段 ARIS skill：/research-lit 等 + 平台 args；后续段追加 composed 指向 IDEA_REPORT。
func buildIdeaDiscoveryPhaseCLIPrompt(phaseID, argsLine string, cfg ARISIdeaDiscoveryConfig) string {
	phaseID = strings.TrimPrefix(strings.TrimSpace(phaseID), "/")
	if phaseID == "" {
		return ""
	}
	argsLine = strings.TrimSpace(argsLine)
	line := "/" + phaseID
	if argsLine != "" {
		line += " " + argsLine
	}
	if ideaDiscoveryPhaseIndex(phaseID) > 0 {
		composed := strings.TrimSpace(cfg.Artifacts.IdeaReport)
		if composed != "" {
			line += fmt.Sprintf(` — composed: %s`, composed)
		}
	}
	return line
}

func ideaDiscoveryPhasedStageBehavior(behavior string) bool {
	behavior = strings.TrimSpace(behavior)
	if !strings.HasPrefix(behavior, "topic.stage.") {
		return false
	}
	code := strings.TrimPrefix(behavior, "topic.stage.")
	return isPhasedTopicStage(code)
}

// parseIdeaDiscoveryPhasedInvokeResult 解析 invoke 返回 JSON；success 仅当 result=="success"。
func parseIdeaDiscoveryPhasedInvokeResult(raw json.RawMessage) (success bool, reason string, err error) {
	obj, err := ideaDiscoveryPhasedResultJSONObject(raw)
	if err != nil {
		return false, "", err
	}
	switch v := obj["result"].(type) {
	case string:
		switch strings.TrimSpace(strings.ToLower(v)) {
		case "success":
			return true, "", nil
		case "false", "failed", "failure", "error":
			return false, phasedInvokeResultReason(obj), nil
		default:
			return false, "", fmt.Errorf("ARIS 返回 JSON 中 result=%q 无法识别，需要 success 或 false", v)
		}
	case bool:
		if v {
			return true, "", nil
		}
		return false, phasedInvokeResultReason(obj), nil
	default:
		return false, "", fmt.Errorf("ARIS 返回 JSON 缺少 result 字段（需要 {\"result\":\"success\"} 或 {\"result\":\"false\"}）")
	}
}

func phasedInvokeResultReason(obj map[string]any) string {
	if obj == nil {
		return "ARIS 报告失败"
	}
	for _, k := range []string{"reason", "error", "message", "summary"} {
		if s := strings.TrimSpace(fmt.Sprint(obj[k])); s != "" && s != "<nil>" {
			return s
		}
	}
	return "ARIS 报告失败"
}

func ideaDiscoveryPhasedResultJSONObject(raw json.RawMessage) (map[string]any, error) {
	text := strings.TrimSpace(string(raw))
	if text == "" {
		return nil, fmt.Errorf("ARIS 未返回可用于判定的 JSON")
	}
	obj, err := extractJSONObject(text)
	if err != nil {
		return nil, fmt.Errorf("ARIS 未返回可用于判定的 JSON: %w", err)
	}
	if s, ok := obj["raw"].(string); ok && strings.TrimSpace(s) != "" {
		if inner, innerErr := extractJSONObject(s); innerErr == nil && len(inner) > 0 {
			if _, has := inner["result"]; has {
				return inner, nil
			}
		}
	}
	return obj, nil
}

func executingIdeaDiscoveryPhaseFromMeta(meta map[string]any) string {
	return executingPhaseFromRuntimeMeta(meta)
}

func nextIdeaDiscoveryPhaseID(phase string) string {
	idx := ideaDiscoveryPhaseIndex(phase)
	if idx < 0 || idx+1 >= len(ideaDiscoveryRunStatePhases) {
		return ""
	}
	return ideaDiscoveryRunStatePhases[idx+1]
}

func ideaDiscoveryCheckpointResumePhase(meta map[string]any, paperRoot, runID string) string {
	if meta != nil {
		normalizeIdeaDiscoveryRuntimeMeta(meta)
		for _, id := range ideaDiscoveryRunStatePhases {
			if phaseStatusInMeta(meta, id) == IdeaDiscoveryPhaseAwaitingConfirm {
				return id
			}
		}
		if p := strings.TrimSpace(pendingPhaseFromMeta(meta)); p != "" {
			if phaseStatusInMeta(meta, p) == IdeaDiscoveryPhaseAwaitingConfirm {
				return p
			}
		}
	}
	pending, err := ideaDiscoveryResumePhase(paperRoot, runID)
	if err != nil {
		return ""
	}
	return pending
}

func ideaDiscoveryPhasedProgressPct(meta map[string]any) uint8 {
	if meta == nil {
		return 1
	}
	normalizeIdeaDiscoveryRuntimeMeta(meta)
	n := len(ideaDiscoveryRunStatePhases)
	if n == 0 {
		return 1
	}
	done := 0
	for _, id := range ideaDiscoveryRunStatePhases {
		if phaseStatusInMeta(meta, id) == IdeaDiscoveryPhaseCompleted {
			done++
		}
	}
	if done >= n {
		return 99
	}
	pct := done * 100 / n
	if pct < 1 {
		return 1
	}
	if pct > 99 {
		return 99
	}
	return uint8(pct)
}

func workspaceFileMinSize(paperRoot, rel string, minBytes int64) bool {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return false
	}
	st, err := os.Stat(filepath.Join(paperRoot, rel))
	if err != nil {
		return false
	}
	return st.Size() >= minBytes
}

func literaturePDFCountUnder(paperRoot, litDirRel string) int {
	litDirRel = strings.TrimSpace(litDirRel)
	if litDirRel == "" {
		return 0
	}
	entries, err := os.ReadDir(filepath.Join(paperRoot, litDirRel))
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(e.Name()), ".pdf") {
			n++
		}
	}
	return n
}

// researchLitStandaloneDeliverablesPresent：/research-lit 在不同 ARIS 版本下文件名不一致，任一达标即视为本段完成。
func researchLitStandaloneDeliverablesPresent(paperRoot string, cfg ARISIdeaDiscoveryConfig) bool {
	type fileMin struct {
		rel string
		min int64
	}
	candidates := []fileMin{
		{"LITERATURE_REVIEW.md", 256},
		{"LITERATURE_REVIEW_SUMMARY.txt", 128},
		{"RESEARCH_LIT_REPORT.md", 256},
		{"REFERENCE_INDEX.md", 256},
		{"EXECUTIVE_SUMMARY.md", 256},
		{"PROJECT_SUMMARY.txt", 128},
		{"references.bib", 64},
	}
	if brief := strings.TrimSpace(cfg.Artifacts.ResearchBrief); brief != "" {
		candidates = append(candidates, fileMin{brief, 128})
	} else {
		candidates = append(candidates, fileMin{"RESEARCH_BRIEF.md", 128})
	}
	for _, c := range candidates {
		if workspaceFileMinSize(paperRoot, c.rel, c.min) {
			return true
		}
	}
	return false
}

func ideaDiscoveryPhasedInvokeMetaKey(phase string) string {
	phase = strings.TrimSpace(phase)
	return "phased_invoke_" + strings.ReplaceAll(phase, "-", "_")
}

func ideaDiscoveryPhasedInvokeCount(meta map[string]any, phase string) int {
	if meta == nil {
		return 0
	}
	k := ideaDiscoveryPhasedInvokeMetaKey(phase)
	if n, ok := meta[k].(float64); ok {
		return int(n)
	}
	if n, ok := meta[k].(int); ok {
		return n
	}
	return 0
}

func bumpIdeaDiscoveryPhasedInvokeCount(meta map[string]any, phase string) {
	if meta == nil {
		return
	}
	meta[ideaDiscoveryPhasedInvokeMetaKey(phase)] = ideaDiscoveryPhasedInvokeCount(meta, phase) + 1
}

func ideaDiscoveryPhasedPhaseSucceeded(paperRoot, runID, phase string, cfg ARISIdeaDiscoveryConfig) bool {
	phase = strings.TrimSpace(phase)
	if phase == "" {
		return false
	}
	if ideaDiscoveryRunStateFileExists(paperRoot, runID) && isRunStatePhaseDone(paperRoot, runID, phase) {
		return true
	}
	switch phase {
	case "research-lit":
		if isResearchLitPhaseDone(paperRoot, runID) {
			return true
		}
		if researchLitStandaloneDeliverablesPresent(paperRoot, cfg) {
			return true
		}
		if workspaceFileMinSize(paperRoot, cfg.Artifacts.CorpusJSON, 64) {
			return true
		}
		if literaturePDFCountUnder(paperRoot, cfg.Artifacts.LiteratureDir) > 0 {
			return true
		}
		return workspaceFileMinSize(paperRoot, cfg.Artifacts.IdeaReport, 256)
	case "idea-creator", "novelty-check", "research-review":
		return workspaceFileMinSize(paperRoot, cfg.Artifacts.IdeaReport, 512)
	case "research-refine-pipeline":
		if workspaceFileMinSize(paperRoot, cfg.Artifacts.FinalProposal, 256) {
			return true
		}
		return workspaceFileMinSize(paperRoot, cfg.Artifacts.ExperimentPlan, 256)
	default:
		return false
	}
}

// IdeaDiscoveryPhasedBotMayRetryPhase 分阶段 bot：每 tick 只打一次 ARIS，不因缺 artifact 在同段上连打。
// （AutoProceed 只影响 skill 内行为；段间推进由 runtime_state.stage + 下轮 scheduler 负责。）
func IdeaDiscoveryPhasedBotMayRetryPhase(cfg ARISIdeaDiscoveryConfig) bool {
	_ = cfg
	return false
}

func ideaDiscoveryPhasedReinvokeFailureReason(paperRoot, phase string) string {
	phase = strings.TrimSpace(phase)
	if phase == "research-lit" {
		return fmt.Sprintf("Idea Discovery 阶段 /research-lit 已执行 ARIS 但未产出 %s", researchLitLandscapeRel)
	}
	return fmt.Sprintf("Idea Discovery 阶段 /%s 已执行 ARIS 但未收到 {\"result\":\"success\"}", phase)
}

func ideaDiscoveryPhasedPhaseInvokeSucceeded(paperRoot, phase string, raw json.RawMessage) (ok bool, failReason string, parseErr error) {
	phase = strings.TrimSpace(phase)
	if phase == "research-lit" {
		if researchLitLandscapePresent(paperRoot) {
			return true, "", nil
		}
		return false, "", fmt.Errorf("缺少必需产物 %s", researchLitLandscapeRel)
	}
	return parseIdeaDiscoveryPhasedInvokeResult(raw)
}

// ideaDiscoveryPhasedStepFailureReason 写入 step.error_msg 的文案（CLI 错误或缺产物）。
func ideaDiscoveryPhasedStepFailureReason(meta map[string]any, phase string, invokeErr error) string {
	phase = strings.TrimSpace(phase)
	if phase == "" {
		phase = executingIdeaDiscoveryPhaseFromMeta(meta)
	}
	if invokeErr != nil {
		return fmt.Sprintf("Idea Discovery 阶段 /%s 失败: %s", phase, strings.TrimSpace(invokeErr.Error()))
	}
	if meta != nil {
		if le, ok := meta["aris_invoke_last_error"].(string); ok && strings.TrimSpace(le) != "" {
			return fmt.Sprintf("Idea Discovery 阶段 /%s 失败: %s", phase, strings.TrimSpace(le))
		}
	}
	return fmt.Sprintf("Idea Discovery 阶段 /%s 已执行 ARIS 但未产出必需产物", phase)
}

func syncIdeaDiscoveryPhasedProgressMeta(meta map[string]any) {
	if meta == nil {
		return
	}
	meta["idea_discovery_invoke_mode"] = "phased"
	meta["exec_progress_phased"] = ideaDiscoveryPhasedProgressPct(meta)
}
