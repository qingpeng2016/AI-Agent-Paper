package topicdiscovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const BehaviorIdeaDiscoveryPhase = "topic.discovery.phase"

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

// IdeaDiscoveryPhasedBotMayRetryPhase 同段 invoke 未产出可验收 artifact 时是否保留 async_llm 重试。
func IdeaDiscoveryPhasedBotMayRetryPhase(cfg ARISIdeaDiscoveryConfig) bool {
	return cfg.Platform.AutoProceed
}

func syncIdeaDiscoveryPhasedProgressMeta(meta map[string]any) {
	if meta == nil {
		return
	}
	meta["idea_discovery_invoke_mode"] = "phased"
	meta["exec_progress_phased"] = ideaDiscoveryPhasedProgressPct(meta)
}
