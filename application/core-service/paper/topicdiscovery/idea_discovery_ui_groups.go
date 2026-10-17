package topicdiscovery

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// IdeaDiscoveryUIGroup 由 checkpoint_phases.pause 合并：连续 pause=false 为一段，pause=true 单独一段。
type IdeaDiscoveryUIGroup struct {
	Index    int      `json:"index"`
	PhaseIDs []string `json:"phase_ids"`
	LabelZh  string   `json:"label_zh"`
}

// BuildIdeaDiscoveryUIGroups 按 run_state phase 顺序生成工作台展示分段。
func BuildIdeaDiscoveryUIGroups(cfg ARISIdeaDiscoveryConfig) []IdeaDiscoveryUIGroup {
	entries := normalizeIdeaDiscoveryCheckpointPhases(cfg.CheckpointPhases)
	var groups []IdeaDiscoveryUIGroup
	var buf []string
	var bufLabels []string

	flush := func() {
		if len(buf) == 0 {
			return
		}
		groups = append(groups, IdeaDiscoveryUIGroup{
			Index:    len(groups),
			PhaseIDs: append([]string(nil), buf...),
			LabelZh:  strings.Join(bufLabels, "、"),
		})
		buf = nil
		bufLabels = nil
	}

	for _, phase := range ideaDiscoveryRunStatePhases {
		e := entries[phase]
		if !e.Pause {
			buf = append(buf, phase)
			bufLabels = append(bufLabels, strings.TrimSpace(e.LabelZh))
			continue
		}
		flush()
		groups = append(groups, IdeaDiscoveryUIGroup{
			Index:    len(groups),
			PhaseIDs: []string{phase},
			LabelZh:  strings.TrimSpace(e.LabelZh),
		})
	}
	flush()
	return groups
}

func uiGroupIndexForPhase(groups []IdeaDiscoveryUIGroup, phase string) int {
	phase = strings.TrimSpace(phase)
	if phase == "" {
		return 0
	}
	for _, g := range groups {
		for _, id := range g.PhaseIDs {
			if id == phase {
				return g.Index
			}
		}
	}
	return 0
}

func uiGroupLabelForPhase(cfg ARISIdeaDiscoveryConfig, phase string) string {
	groups := BuildIdeaDiscoveryUIGroups(cfg)
	idx := uiGroupIndexForPhase(groups, phase)
	if idx >= 0 && idx < len(groups) {
		if s := strings.TrimSpace(groups[idx].LabelZh); s != "" {
			return s
		}
	}
	return cfg.PhaseLabelZh(phase)
}

func ideaDiscoveryRunStatePhaseDone(paperRoot, runID, phaseID string) bool {
	phaseID = strings.TrimSpace(phaseID)
	if phaseID == "" {
		return false
	}
	statusByPhase, err := readRunStatePhaseStatuses(paperRoot, runID)
	if err != nil {
		return false
	}
	return runStatePhaseCompleted(statusByPhase[phaseID])
}

func uiGroupFullyCompleted(paperRoot, runID string, g IdeaDiscoveryUIGroup) bool {
	for _, id := range g.PhaseIDs {
		if !ideaDiscoveryRunStatePhaseDone(paperRoot, runID, id) {
			return false
		}
	}
	return len(g.PhaseIDs) > 0
}

func countCompletedUIGroups(paperRoot, runID string, groups []IdeaDiscoveryUIGroup) int {
	n := 0
	for _, g := range groups {
		if uiGroupFullyCompleted(paperRoot, runID, g) {
			n++
		}
	}
	return n
}

func readRunStatePhaseStatuses(paperRoot, runID string) (map[string]string, error) {
	path := mustIdeaDiscoveryConfig().RunStateFileAbs(paperRoot, runID)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var st struct {
		Phases []struct {
			Phase  string `json:"phase"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"phases"`
	}
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, ph := range st.Phases {
		id := strings.TrimSpace(ph.Phase)
		if id == "" {
			id = strings.TrimSpace(ph.Name)
		}
		if id == "" {
			continue
		}
		out[id] = ph.Status
	}
	return out, nil
}

func syncIdeaDiscoveryUIGroupsToMeta(meta map[string]any, paperRoot, runID string, cfg ARISIdeaDiscoveryConfig) {
	if meta == nil {
		return
	}
	groups := BuildIdeaDiscoveryUIGroups(cfg)
	uiGroups := make([]map[string]any, 0, len(groups))
	for _, g := range groups {
		uiGroups = append(uiGroups, map[string]any{
			"index":     g.Index,
			"phase_ids": g.PhaseIDs,
			"label_zh":  g.LabelZh,
		})
	}
	meta["aris_ui_groups"] = uiGroups
	meta["aris_ui_group_count"] = len(groups)
	meta["aris_ui_groups_completed"] = countCompletedUIGroups(paperRoot, runID, groups)

	pending, err := ideaDiscoveryResumePhase(paperRoot, runID)
	if err == nil && strings.TrimSpace(pending) != "" {
		idx := uiGroupIndexForPhase(groups, pending)
		meta["aris_ui_group_index"] = idx
		if idx >= 0 && idx < len(groups) {
			meta["aris_ui_group_label"] = groups[idx].LabelZh
		}
	}
}

// IdeaDiscoveryCheckpointDetail 段间检查点 Lark/日志文案（按 conf 合并后的 UI 段名）。
func IdeaDiscoveryCheckpointDetail(pauseAfterPhase string) string {
	p := strings.TrimSpace(pauseAfterPhase)
	if p == "" {
		return "Idea Discovery 段间检查点，待用户在平台确认并继续"
	}
	cfg := mustIdeaDiscoveryConfig()
	label := uiGroupLabelForPhase(cfg, p)
	return fmt.Sprintf("Idea Discovery 已完成本轮 invoke，下一阶段「%s」（%s）需在平台确认后继续", label, p)
}

func ideaDiscoveryProgressPctByUIGroups(paperRoot, runID string, cfg ARISIdeaDiscoveryConfig) uint8 {
	groups := BuildIdeaDiscoveryUIGroups(cfg)
	if len(groups) == 0 {
		return 1
	}
	done := countCompletedUIGroups(paperRoot, runID, groups)
	if done >= len(groups) {
		return 99
	}
	return uint8(min(99, max(1, done*100/len(groups))))
}
