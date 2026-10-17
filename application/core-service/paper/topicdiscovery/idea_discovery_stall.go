package topicdiscovery

import (
	"fmt"
	"strings"
)

// ideaDiscoveryMaxStallInvokes：同一 pending 阶段在无平台 checkpoint 时连续 invoke 上限。
// 用于 ARIS 技能 AUTO_PROCEED=false 在内部停住、但 run_state 尚未前进时的 headless PROCEED 桥接；
// 超过上限则失败，避免 bot 每 tick 重调导致 Claude 会话上下文无限增长。
const ideaDiscoveryMaxStallInvokes = 8

func ideaDiscoveryStallMetaKey(phase string) string {
	phase = strings.TrimSpace(phase)
	if phase == "" {
		return "aris_stall_unknown"
	}
	return "aris_stall_" + strings.ReplaceAll(phase, "-", "_")
}

func ideaDiscoveryStallCount(meta map[string]any, phase string) int {
	if meta == nil {
		return 0
	}
	v, ok := meta[ideaDiscoveryStallMetaKey(phase)]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

func ideaDiscoveryStallBump(meta map[string]any, phase string) int {
	if meta == nil {
		return 0
	}
	n := ideaDiscoveryStallCount(meta, phase) + 1
	meta[ideaDiscoveryStallMetaKey(phase)] = n
	return n
}

func ideaDiscoveryStallClear(meta map[string]any, phase string) {
	if meta == nil {
		return
	}
	delete(meta, ideaDiscoveryStallMetaKey(phase))
}

func ideaDiscoveryStallExceededError(phase, runID string, attempts int) error {
	return fmt.Errorf(
		"idea discovery 阶段 %q 已连续 invoke %d 次仍未前进（run_id=%s）；请检查 .aris/runs/%s.json、idea-stage/IDEA_REPORT.md 与 ARIS run_state；或在前端 Resume 并选择 PROCEED",
		phase, attempts, runID, runID,
	)
}
