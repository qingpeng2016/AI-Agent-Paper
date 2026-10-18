package topicdiscovery

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// 平台 / bot 认的 idea_discovery.runtime_state 形状（与 DB 存库 JSON 一致）：
//
//	{
//	  "other": { "async_llm": true },
//	  "stage": {
//	    "1": { "name": "research-lit", "status": "executing" },
//	    "2": { "name": "idea-creator", "status": "not_started" },
//	    ...
//	  }
//	}
//
// 五段 name 固定为 ideaDiscoveryRunStatePhases；status 为四态英文字符串。
// bot 是否排队：other.async_llm=true 且 stage 中恰好一段 status=executing。
// 打哪条 ARIS：executing 段的 name → /research-lit 等。

func parseRuntimeStateRoot(raw []byte) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var rt map[string]any
	if err := json.Unmarshal(raw, &rt); err != nil || rt == nil {
		return nil
	}
	return rt
}

func runtimeStateOtherAsyncLLM(raw []byte) bool {
	rt := parseRuntimeStateRoot(raw)
	if rt == nil {
		return false
	}
	other, ok := rt["other"].(map[string]any)
	if !ok || other == nil {
		return false
	}
	switch v := other["async_llm"].(type) {
	case bool:
		return v
	case float64:
		return v != 0
	default:
		return false
	}
}

func runtimeStateStageMap(raw []byte) map[string]any {
	rt := parseRuntimeStateRoot(raw)
	if rt == nil {
		return nil
	}
	stage, ok := rt["stage"].(map[string]any)
	if !ok || len(stage) == 0 {
		return nil
	}
	return stage
}

func sortedStageSlotKeys(stage map[string]any) []string {
	if stage == nil {
		return nil
	}
	keys := make([]string, 0, len(stage))
	for k := range stage {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		ni, ei := strconv.Atoi(keys[i])
		nj, ej := strconv.Atoi(keys[j])
		if ei == nil && ej == nil {
			return ni < nj
		}
		return keys[i] < keys[j]
	})
	return keys
}

func stageSlotPhaseName(entry map[string]any) string {
	if entry == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(entry["name"]))
}

func stageSlotPhaseStatus(entry map[string]any) string {
	if entry == nil {
		return IdeaDiscoveryPhaseNotStarted
	}
	return normalizeIdeaDiscoveryPhaseStatus(fmt.Sprint(entry["status"]))
}

// ExecutingPhaseFromRuntimeState 只读 runtime_state.stage，返回当前 executing 的 ARIS phase name（如 research-lit）。
func ExecutingPhaseFromRuntimeState(raw []byte) string {
	stage := runtimeStateStageMap(raw)
	if stage == nil {
		return ""
	}
	for _, slot := range sortedStageSlotKeys(stage) {
		entry, _ := stage[slot].(map[string]any)
		if stageSlotPhaseStatus(entry) == IdeaDiscoveryPhaseExecuting {
			return stageSlotPhaseName(entry)
		}
	}
	return ""
}

// PhaseStatusFromRuntimeState 读某 phase 在 stage 槽位中的 status。
func PhaseStatusFromRuntimeState(raw []byte, phaseName string) string {
	phaseName = strings.TrimSpace(phaseName)
	if phaseName == "" {
		return IdeaDiscoveryPhaseNotStarted
	}
	stage := runtimeStateStageMap(raw)
	if stage == nil {
		return IdeaDiscoveryPhaseNotStarted
	}
	for _, slot := range sortedStageSlotKeys(stage) {
		entry, _ := stage[slot].(map[string]any)
		if stageSlotPhaseName(entry) == phaseName {
			return stageSlotPhaseStatus(entry)
		}
	}
	return IdeaDiscoveryPhaseNotStarted
}

func executingPhaseFromRuntimeMeta(meta map[string]any) string {
	if meta == nil {
		return ""
	}
	if stage, ok := meta["stage"].(map[string]any); ok && len(stage) > 0 {
		for _, slot := range sortedStageSlotKeys(stage) {
			entry, _ := stage[slot].(map[string]any)
			if stageSlotPhaseStatus(entry) == IdeaDiscoveryPhaseExecuting {
				return stageSlotPhaseName(entry)
			}
		}
	}
	normalizeIdeaDiscoveryRuntimeMeta(meta)
	for _, id := range ideaDiscoveryRunStatePhases {
		if phaseStatusInMeta(meta, id) == IdeaDiscoveryPhaseExecuting {
			return id
		}
	}
	return ""
}

func phaseStatusFromRuntimeMeta(meta map[string]any, phaseName string) string {
	phaseName = strings.TrimSpace(phaseName)
	if meta != nil {
		if stage, ok := meta["stage"].(map[string]any); ok {
			for _, slot := range sortedStageSlotKeys(stage) {
				entry, _ := stage[slot].(map[string]any)
				if stageSlotPhaseName(entry) == phaseName {
					return stageSlotPhaseStatus(entry)
				}
			}
		}
	}
	return phaseStatusInMeta(meta, phaseName)
}
