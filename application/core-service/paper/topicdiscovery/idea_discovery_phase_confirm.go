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

func phaseUserConfirmedForInvoke(meta map[string]any, phase string) bool {
	st := phaseStatusInMeta(meta, phase)
	return st == IdeaDiscoveryPhaseExecuting || st == IdeaDiscoveryPhaseCompleted
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
				meta[phase] = IdeaDiscoveryPhaseExecuting
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
}

// IdeaDiscoveryNeedsUserConfirm 当前 pending 段需 pause 且状态不是 executing/completed。
func IdeaDiscoveryNeedsUserConfirm(meta map[string]any, pendingPhase string) bool {
	return ideaDiscoveryNeedsUserConfirm(meta, pendingPhase, mustIdeaDiscoveryConfig())
}

func ideaDiscoveryNeedsUserConfirm(meta map[string]any, pendingPhase string, cfg ARISIdeaDiscoveryConfig) bool {
	pendingPhase = strings.TrimSpace(pendingPhase)
	if pendingPhase == "" || !cfg.PhaseNeedsUserCheckpoint(pendingPhase) {
		return false
	}
	normalizeIdeaDiscoveryRuntimeMeta(meta)
	st := phaseStatusInMeta(meta, pendingPhase)
	return st != IdeaDiscoveryPhaseExecuting && st != IdeaDiscoveryPhaseCompleted
}
