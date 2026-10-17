package topicdiscovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// seedIdeaDiscoveryRunStateFromWorkspace ARIS headless 常不写 .aris/runs/*.json；有 IDEA_REPORT 时平台补最小 run_state 供续跑/进度。
func seedIdeaDiscoveryRunStateFromWorkspace(paperRoot, runID string) error {
	if ideaDiscoveryRunStateFileExists(paperRoot, strings.TrimSpace(runID)) {
		return refreshIdeaDiscoveryRunStateFromWorkspace(paperRoot, runID)
	}
	return writeIdeaDiscoveryRunStateFromReport(paperRoot, runID)
}

// refreshIdeaDiscoveryRunStateFromWorkspace 根据 IDEA_REPORT 抬升 done 阶段（不降级已有 done）。
func refreshIdeaDiscoveryRunStateFromWorkspace(paperRoot, runID string) error {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return nil
	}
	cfg := mustIdeaDiscoveryConfig()
	path := cfg.RunStateFileAbs(paperRoot, runID)
	b, err := os.ReadFile(path)
	if err != nil {
		return writeIdeaDiscoveryRunStateFromReport(paperRoot, runID)
	}
	var doc struct {
		RunID    string `json:"run_id"`
		SeededBy string `json:"seeded_by"`
		Phases   []struct {
			Phase  string `json:"phase"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"phases"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return err
	}
	doneThrough := ideaDiscoveryDoneThroughFromReport(paperRoot, cfg)
	if doneThrough <= 0 {
		return nil
	}
	statusByID := map[string]string{}
	for _, ph := range doc.Phases {
		id := runStateEntryPhaseID(ph.Phase, ph.Name, "")
		if id != "" {
			statusByID[id] = ph.Status
		}
	}
	changed := false
	newPhases := make([]map[string]string, 0, len(ideaDiscoveryRunStatePhases))
	for i, id := range ideaDiscoveryRunStatePhases {
		st := strings.TrimSpace(statusByID[id])
		if st == "" {
			st = "pending"
		}
		if i < doneThrough && !runStatePhaseCompleted(st) {
			st = "done"
			changed = true
		}
		if statusByID[id] != st {
			changed = true
		}
		newPhases = append(newPhases, map[string]string{"phase": id, "status": st})
	}
	if !changed {
		return nil
	}
	doc.Phases = nil
	for _, p := range newPhases {
		doc.Phases = append(doc.Phases, struct {
			Phase  string `json:"phase"`
			Name   string `json:"name"`
			Status string `json:"status"`
		}{Phase: p["phase"], Status: p["status"]})
	}
	if strings.TrimSpace(doc.RunID) == "" {
		doc.RunID = runID
	}
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func writeIdeaDiscoveryRunStateFromReport(paperRoot, runID string) error {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return nil
	}
	cfg := mustIdeaDiscoveryConfig()
	doneThrough := ideaDiscoveryDoneThroughFromReport(paperRoot, cfg)
	if doneThrough == 0 {
		return nil
	}
	phases := make([]map[string]string, 0, len(ideaDiscoveryRunStatePhases))
	for i, id := range ideaDiscoveryRunStatePhases {
		st := "pending"
		if i < doneThrough {
			st = "done"
		}
		phases = append(phases, map[string]string{"phase": id, "status": st})
	}
	payload := map[string]any{
		"run_id":    runID,
		"phases":    phases,
		"seeded_by": "ai-agent-paper",
	}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	path := cfg.RunStateFileAbs(paperRoot, runID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func ideaDiscoveryDoneThroughFromReport(paperRoot string, cfg ARISIdeaDiscoveryConfig) int {
	reportAbs := filepath.Join(paperRoot, strings.TrimSpace(cfg.Artifacts.IdeaReport))
	b, err := os.ReadFile(reportAbs)
	if err != nil || len(b) < 512 {
		return 0
	}
	text := string(b)
	doneThrough := 0
	if strings.Contains(text, "文献景观") || strings.Contains(text, "Literature Landscape") ||
		strings.Contains(text, "## 📚") || strings.Contains(text, "文献综述与研究景观") ||
		strings.Contains(text, "PHASE 1") || strings.Contains(text, "Phase 1") {
		doneThrough = 1
	}
	if doneThrough >= 1 && ideaReportSignalsIdeaCreatorDone(text) {
		doneThrough = 2
	}
	if doneThrough >= 2 && (strings.Contains(text, "新颖性") || strings.Contains(text, "Novelty Verification") ||
		strings.Contains(text, "#novelty")) {
		doneThrough = 3
	}
	if doneThrough >= 3 && (strings.Contains(text, "外部审查") || strings.Contains(text, "External Critical Review") ||
		strings.Contains(text, "#external-critical-review")) {
		doneThrough = 4
	}
	if doneThrough >= 4 && (fileExistsNonEmpty(filepath.Join(paperRoot, cfg.Artifacts.FinalProposal)) ||
		strings.Contains(text, "FINAL_PROPOSAL") || strings.Contains(text, "实验规划")) {
		doneThrough = 5
	}
	if doneThrough > len(ideaDiscoveryRunStatePhases) {
		doneThrough = len(ideaDiscoveryRunStatePhases)
	}
	return doneThrough
}

func ideaReportSignalsIdeaCreatorDone(text string) bool {
	markers := []string{
		"想法候选", "Ranked Ideas", "推荐的Top", "创意生成", "研究创意",
		"初步想法", "Idea 1.", "Idea 1:", "💡 初步", "Phase 2", "PHASE 2",
		"想法评估与排名", "想法池",
	}
	for _, m := range markers {
		if strings.Contains(text, m) {
			return true
		}
	}
	return false
}

func phaseIndexInPipeline(phaseID string) int {
	phaseID = strings.TrimSpace(phaseID)
	for i, id := range ideaDiscoveryRunStatePhases {
		if id == phaseID {
			return i
		}
	}
	return -1
}

func fileExistsNonEmpty(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st != nil && !st.IsDir() && st.Size() > 0
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
