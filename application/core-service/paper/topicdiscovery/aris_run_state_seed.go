package topicdiscovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// seedIdeaDiscoveryRunStateFromWorkspace ARIS headless 常不写 .aris/runs/*.json；有 IDEA_REPORT 时平台补最小 run_state 供续跑/进度。
func seedIdeaDiscoveryRunStateFromWorkspace(paperRoot, runID string) error {
	runID = strings.TrimSpace(runID)
	if runID == "" || ideaDiscoveryRunStateFileExists(paperRoot, runID) {
		return nil
	}
	cfg := mustIdeaDiscoveryConfig()
	reportAbs := filepath.Join(paperRoot, strings.TrimSpace(cfg.Artifacts.IdeaReport))
	b, err := os.ReadFile(reportAbs)
	if err != nil || len(b) < 512 {
		return nil
	}
	text := string(b)
	doneThrough := 0
	if strings.Contains(text, "文献景观") || strings.Contains(text, "Literature Landscape") || strings.Contains(text, "## 📚") {
		doneThrough = 1
	}
	if doneThrough >= 1 && (strings.Contains(text, "想法候选") || strings.Contains(text, "Ranked Ideas") || strings.Contains(text, "推荐的Top")) {
		doneThrough = 2
	}
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
		"run_id": runID,
		"phases": phases,
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
