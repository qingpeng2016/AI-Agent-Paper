package topicdiscovery

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var errNoRunStateFile = errors.New("run state file missing")

// ideaDiscoveryResumePhase 只读工作区 .aris/runs/{run_id}.json（由 aris /idea-discovery 写入），不调用 libs/aris 脚本。
func ideaDiscoveryResumePhase(paperRoot, runID string) (string, error) {
	phase, complete, err := readRunStateResumeTarget(paperRoot, runID)
	if err != nil {
		if errors.Is(err, errNoRunStateFile) {
			if len(ideaDiscoveryRunStatePhases) == 0 {
				return "", nil
			}
			return ideaDiscoveryRunStatePhases[0], nil
		}
		return "", err
	}
	if complete {
		return "", nil
	}
	return phase, nil
}

func readRunStateResumeTarget(paperRoot, runID string) (phase string, complete bool, err error) {
	path := mustIdeaDiscoveryConfig().RunStateFileAbs(paperRoot, runID)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, errNoRunStateFile
		}
		return "", false, err
	}
	var st struct {
		Phases []struct {
			Phase  string `json:"phase"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"phases"`
	}
	if err := json.Unmarshal(b, &st); err != nil {
		return "", false, err
	}
	for _, ph := range st.Phases {
		id := runStateEntryPhaseID(ph.Phase, ph.Name, "")
		if id == "" {
			continue
		}
		if runStatePhaseNeedsWork(ph.Status) {
			return id, false, nil
		}
	}
	return "", true, nil
}

func runStatePhaseNeedsWork(status string) bool {
	// ARIS run_state: done = artifact produced (may be ✓ unaccepted); resume skips to next pending phase.
	return !runStatePhaseCompleted(status)
}

func runStatePhaseCompleted(status string) bool {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "done", "accepted", "provisional", "skipped":
		return true
	default:
		return false
	}
}

func runStateEntryPhaseID(phase, name, id string) string {
	if s := strings.TrimSpace(phase); s != "" {
		return s
	}
	if s := strings.TrimSpace(name); s != "" {
		return s
	}
	return strings.TrimSpace(id)
}

func runStatePhaseStatus(paperRoot, runID, phaseID string) (status string, found bool) {
	phaseID = strings.TrimSpace(phaseID)
	if phaseID == "" {
		return "", false
	}
	path := mustIdeaDiscoveryConfig().RunStateFileAbs(paperRoot, runID)
	b, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	var st struct {
		Phases []struct {
			Phase  string `json:"phase"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"phases"`
	}
	if json.Unmarshal(b, &st) != nil {
		return "", false
	}
	for _, ph := range st.Phases {
		id := runStateEntryPhaseID(ph.Phase, ph.Name, "")
		if id == phaseID {
			return ph.Status, true
		}
	}
	return "", false
}

func isRunStatePhaseDone(paperRoot, runID, phaseID string) bool {
	st, ok := runStatePhaseStatus(paperRoot, runID, phaseID)
	return ok && runStatePhaseCompleted(st)
}

// ideaDiscoveryCompletedPhaseCount 读工作区 run_state，已完成 phase 数量（DB 未同步时的依据）。
func ideaDiscoveryCompletedPhaseCount(paperRoot, runID string) int {
	path := mustIdeaDiscoveryConfig().RunStateFileAbs(paperRoot, runID)
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var st struct {
		Phases []struct {
			Status string `json:"status"`
		} `json:"phases"`
	}
	if json.Unmarshal(b, &st) != nil {
		return 0
	}
	n := 0
	for _, ph := range st.Phases {
		if runStatePhaseCompleted(ph.Status) {
			n++
		}
	}
	return n
}

// newestIdeaDiscoveryRunStateID 工作区 .aris/runs/*.json 中最新一份的 run_id（ARIS 内建 slug 可能与平台 meta 不同）。
func newestIdeaDiscoveryRunStateID(paperRoot string) string {
	cfg := mustIdeaDiscoveryConfig()
	rel := strings.TrimSpace(cfg.Workspace.RunStateDir)
	if rel == "" {
		rel = ".aris/runs"
	}
	dir := filepath.Join(paperRootJoin(paperRoot, rel), "")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	type cand struct {
		id  string
		mod int64
	}
	var cands []cand
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		if strings.TrimSpace(id) == "" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		cands = append(cands, cand{id: id, mod: info.ModTime().UnixNano()})
	}
	if len(cands) == 0 {
		return ""
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].mod > cands[j].mod })
	return cands[0].id
}

// ideaDiscoveryRunStateFileExists 工作区是否已有 .aris/runs/{run_id}.json（ARIS 首次 invoke 前通常不存在）。
func ideaDiscoveryRunStateFileExists(paperRoot, runID string) bool {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return false
	}
	path := mustIdeaDiscoveryConfig().RunStateFileAbs(paperRoot, runID)
	st, err := os.Stat(path)
	return err == nil && st != nil && !st.IsDir()
}

// syncIdeaDiscoveryRunProgressToMeta 把 run_state 进度写入 step.result（展示/排查用）。
func syncIdeaDiscoveryRunProgressToMeta(meta map[string]any, paperRoot, runID string) {
	if meta == nil {
		return
	}
	completed := ideaDiscoveryCompletedPhaseCount(paperRoot, runID)
	meta["aris_phases_completed"] = completed
	if pending, err := ideaDiscoveryResumePhase(paperRoot, runID); err == nil && strings.TrimSpace(pending) != "" {
		meta["aris_pending_phase"] = pending
	}
	cfg := mustIdeaDiscoveryConfig()
	syncIdeaDiscoveryPhaseStatusesToMeta(meta, paperRoot, runID, cfg)
	syncIdeaDiscoveryUIGroupsToMeta(meta, paperRoot, runID, cfg)
}
