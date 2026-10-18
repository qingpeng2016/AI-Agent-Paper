package topicdiscovery

import (
	"fmt"
	"strings"
)

// runtime_state 各 ARIS 段状态（英文），与前端一致：
// not_started | awaiting_confirm | executing | completed

const (
	IdeaDiscoveryPhaseNotStarted      = "not_started"
	IdeaDiscoveryPhaseAwaitingConfirm = "awaiting_confirm"
	IdeaDiscoveryPhaseExecuting       = "executing"
	IdeaDiscoveryPhaseCompleted       = "completed"
)

func ideaDiscoveryPhaseConfirmKey(phase string) bool {
	phase = strings.TrimSpace(phase)
	if phase == "" {
		return false
	}
	for _, id := range ideaDiscoveryRunStatePhases {
		if id == phase {
			return true
		}
	}
	return false
}

func ideaDiscoveryPhaseIndex(phase string) int {
	phase = strings.TrimSpace(phase)
	for i, id := range ideaDiscoveryRunStatePhases {
		if id == phase {
			return i
		}
	}
	return -1
}

func phaseStatusInMeta(meta map[string]any, phase string) string {
	if meta == nil {
		return IdeaDiscoveryPhaseNotStarted
	}
	normalizeIdeaDiscoveryRuntimeMeta(meta)
	v, ok := meta[strings.TrimSpace(phase)]
	if !ok {
		return IdeaDiscoveryPhaseNotStarted
	}
	switch t := v.(type) {
	case string:
		return normalizeIdeaDiscoveryPhaseStatus(t)
	case bool:
		if t {
			return IdeaDiscoveryPhaseExecuting
		}
		return IdeaDiscoveryPhaseNotStarted
	case float64:
		if t != 0 {
			return IdeaDiscoveryPhaseExecuting
		}
		return IdeaDiscoveryPhaseNotStarted
	default:
		return IdeaDiscoveryPhaseNotStarted
	}
}

func normalizeIdeaDiscoveryPhaseStatus(s string) string {
	switch strings.TrimSpace(strings.ToLower(s)) {
	case IdeaDiscoveryPhaseNotStarted, "pending":
		return IdeaDiscoveryPhaseNotStarted
	case IdeaDiscoveryPhaseAwaitingConfirm, "waiting_confirm", "awaiting_user":
		return IdeaDiscoveryPhaseAwaitingConfirm
	case IdeaDiscoveryPhaseExecuting, "running", "confirmed", "in_progress", "true", "1":
		return IdeaDiscoveryPhaseExecuting
	case IdeaDiscoveryPhaseCompleted, "done", "complete":
		return IdeaDiscoveryPhaseCompleted
	default:
		if s == "" {
			return IdeaDiscoveryPhaseNotStarted
		}
		return s
	}
}

// SeedIdeaDiscoveryRuntimePhaseStatuses 创建 run：五段 key 全写入，仅 first 为 executing，其余 not_started。
func SeedIdeaDiscoveryRuntimePhaseStatuses(meta map[string]any, firstExecuting string) {
	if meta == nil {
		return
	}
	firstExecuting = strings.TrimSpace(firstExecuting)
	if firstExecuting == "" {
		firstExecuting = ideaDiscoveryRunStatePhases[0]
	}
	for _, id := range ideaDiscoveryRunStatePhases {
		if id == firstExecuting {
			setPhaseStatusInMeta(meta, id, IdeaDiscoveryPhaseExecuting)
		} else {
			setPhaseStatusInMeta(meta, id, IdeaDiscoveryPhaseNotStarted)
		}
	}
}

func setPhaseStatusInMeta(meta map[string]any, phase, status string) {
	if meta == nil {
		return
	}
	phase = strings.TrimSpace(phase)
	if !ideaDiscoveryPhaseConfirmKey(phase) {
		return
	}
	meta[phase] = normalizeIdeaDiscoveryPhaseStatus(status)
}

// IdeaDiscoveryBotMayInvokePending：只读工作区 run_state（ARIS 写入，平台不写）。
// DB 该 pending 为 executing；有 JSON 且该段未 done 才打 aris（AUTO_PROCEED=true 时 drain）。
func IdeaDiscoveryBotMayInvokePending(meta map[string]any, pendingPhase, paperRoot, runID string) bool {
	cfg := mustIdeaDiscoveryConfig()
	pendingPhase = strings.TrimSpace(pendingPhase)
	if pendingPhase == "" {
		return false
	}
	if phaseStatusInMeta(meta, pendingPhase) != IdeaDiscoveryPhaseExecuting {
		return false
	}
	if ideaDiscoveryNeedsUserConfirm(meta, pendingPhase, cfg) {
		return false
	}
	if !cfg.Platform.AutoProceed {
		return false
	}
	if !ideaDiscoveryRunStateFileExists(paperRoot, runID) {
		return false
	}
	return !isRunStatePhaseDone(paperRoot, runID, pendingPhase)
}

func hasLegacyIdeaDiscoveryRuntimeKeys(meta map[string]any) bool {
	if meta == nil {
		return false
	}
	for _, k := range []string{
		"aris_phase_drain", "aris_checkpoint_proceed", "aris_awaiting_checkpoint",
		"aris_pause_after_phase", "aris_between_invokes",
	} {
		if _, ok := meta[k]; ok {
			return true
		}
	}
	for _, id := range ideaDiscoveryRunStatePhases {
		if v, ok := meta[id]; ok {
			switch v.(type) {
			case bool, float64:
				return true
			}
		}
	}
	return false
}

func normalizeIdeaDiscoveryRuntimeMeta(meta map[string]any) {
	if meta == nil {
		return
	}
	if p := strings.TrimSpace(fmt.Sprint(meta["aris_phase_drain"])); p != "" && ideaDiscoveryPhaseConfirmKey(p) {
		meta[p] = IdeaDiscoveryPhaseExecuting
	}
	if p := strings.TrimSpace(fmt.Sprint(meta["aris_checkpoint_proceed"])); p != "" && ideaDiscoveryPhaseConfirmKey(p) {
		meta[p] = IdeaDiscoveryPhaseExecuting
	}
	for _, id := range ideaDiscoveryRunStatePhases {
		if v, ok := meta[id]; ok {
			switch t := v.(type) {
			case bool:
				if t {
					meta[id] = IdeaDiscoveryPhaseExecuting
				} else {
					meta[id] = IdeaDiscoveryPhaseNotStarted
				}
			case float64:
				if t != 0 {
					meta[id] = IdeaDiscoveryPhaseExecuting
				} else {
					meta[id] = IdeaDiscoveryPhaseNotStarted
				}
			case string:
				meta[id] = normalizeIdeaDiscoveryPhaseStatus(t)
			}
		}
	}
	delete(meta, "aris_phase_drain")
	delete(meta, "aris_checkpoint_proceed")
	delete(meta, "aris_awaiting_checkpoint")
	delete(meta, "aris_pause_after_phase")
	delete(meta, "aris_between_invokes")
}

func pendingPhaseFromMeta(meta map[string]any) string {
	if meta == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(meta["aris_pending_phase"]))
}

// syncIdeaDiscoveryPhaseStatusesToMeta 按 run_state pending + 用户确认位，写齐各 phase 四态字符串。
func syncIdeaDiscoveryPhaseStatusesToMeta(meta map[string]any, paperRoot, runID string, cfg ARISIdeaDiscoveryConfig) {
	if meta == nil {
		return
	}
	normalizeIdeaDiscoveryRuntimeMeta(meta)
	pending, _ := ideaDiscoveryResumePhase(paperRoot, runID)
	pendingIdx := ideaDiscoveryPhaseIndex(pending)

	for i, phase := range ideaDiscoveryRunStatePhases {
		if isRunStatePhaseDone(paperRoot, runID, phase) {
			meta[phase] = IdeaDiscoveryPhaseCompleted
			continue
		}
		if pending == "" {
			if i == 0 && phaseStatusInMeta(meta, phase) == IdeaDiscoveryPhaseExecuting {
				meta[phase] = IdeaDiscoveryPhaseExecuting
			} else if phaseStatusInMeta(meta, phase) == IdeaDiscoveryPhaseCompleted {
				meta[phase] = IdeaDiscoveryPhaseCompleted
			} else if i > 0 {
				meta[phase] = IdeaDiscoveryPhaseNotStarted
			}
			continue
		}
		if i < pendingIdx {
			if isRunStatePhaseDone(paperRoot, runID, phase) {
				meta[phase] = IdeaDiscoveryPhaseCompleted
			} else {
				cur := phaseStatusInMeta(meta, phase)
				switch cur {
				case IdeaDiscoveryPhaseExecuting:
					meta[phase] = IdeaDiscoveryPhaseExecuting
				case IdeaDiscoveryPhaseCompleted:
					meta[phase] = IdeaDiscoveryPhaseCompleted
				case IdeaDiscoveryPhaseAwaitingConfirm:
					meta[phase] = IdeaDiscoveryPhaseAwaitingConfirm
				default:
					meta[phase] = IdeaDiscoveryPhaseNotStarted
				}
			}
			continue
		}
		if i > pendingIdx {
			meta[phase] = IdeaDiscoveryPhaseNotStarted
			continue
		}
		// pending 段
		cur := phaseStatusInMeta(meta, phase)
		if cur == IdeaDiscoveryPhaseExecuting {
			meta[phase] = IdeaDiscoveryPhaseExecuting
			continue
		}
		if cfg.PhaseNeedsUserCheckpoint(phase) {
			meta[phase] = IdeaDiscoveryPhaseAwaitingConfirm
		} else {
			meta[phase] = IdeaDiscoveryPhaseExecuting
		}
	}
	// 任意时刻最多一段 executing（pending）；其余若误标 executing 则收回。
	for _, id := range ideaDiscoveryRunStatePhases {
		if id == pending {
			continue
		}
		if phaseStatusInMeta(meta, id) != IdeaDiscoveryPhaseExecuting {
			continue
		}
		if isRunStatePhaseDone(paperRoot, runID, id) {
			meta[id] = IdeaDiscoveryPhaseCompleted
		} else {
			meta[id] = IdeaDiscoveryPhaseNotStarted
		}
	}
}

// IdeaDiscoveryNeedsUserConfirm 当前 pending 段需 pause 且状态不是 executing/completed。
func IdeaDiscoveryNeedsUserConfirm(meta map[string]any, pendingPhase string) bool {
	return ideaDiscoveryNeedsUserConfirm(meta, pendingPhase, mustIdeaDiscoveryConfig())
}

// IdeaDiscoveryRunAwaitingUserConfirm run 是否处于段间人工检查点（无 executing、无 async_llm 排队）。
func IdeaDiscoveryRunAwaitingUserConfirm(meta map[string]any) bool {
	if meta == nil {
		return false
	}
	normalizeIdeaDiscoveryRuntimeMeta(meta)
	if v, ok := meta["async_llm"].(bool); ok && v {
		return false
	}
	for _, id := range ideaDiscoveryRunStatePhases {
		if phaseStatusInMeta(meta, id) == IdeaDiscoveryPhaseExecuting {
			return false
		}
	}
	pending := pendingPhaseFromMeta(meta)
	return IdeaDiscoveryNeedsUserConfirm(meta, pending)
}

func ideaDiscoveryNeedsUserConfirm(meta map[string]any, pendingPhase string, cfg ARISIdeaDiscoveryConfig) bool {
	pendingPhase = strings.TrimSpace(pendingPhase)
	if pendingPhase == "" || !cfg.PhaseNeedsUserCheckpoint(pendingPhase) {
		return false
	}
	normalizeIdeaDiscoveryRuntimeMeta(meta)
	st := phaseStatusFromRuntimeMeta(meta, pendingPhase)
	return st != IdeaDiscoveryPhaseExecuting && st != IdeaDiscoveryPhaseCompleted
}
