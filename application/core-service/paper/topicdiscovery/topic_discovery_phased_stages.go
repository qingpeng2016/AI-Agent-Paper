package topicdiscovery

import (
	"context"
	"fmt"
	"sort"
	"strings"

	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

// topicDiscoveryPhasedStages：工作流 1 五段 ARIS skill，每段一条 paper_output_topic_step；最后 topic_confirm。
var topicDiscoveryPhasedStages = []string{
	"research_lit",
	"idea_creator",
	"novelty_check",
	"research_review",
	"research_refine",
	"topic_confirm",
}

// TopicDiscoveryPhasedStageCodes 供 bot 按序扫描 running step。
func TopicDiscoveryPhasedStageCodes() []string {
	out := make([]string, 0, len(topicDiscoveryPhasedStages))
	for _, s := range topicDiscoveryPhasedStages {
		if s == "topic_confirm" {
			continue
		}
		out = append(out, s)
	}
	return out
}

func isPhasedTopicStage(stageCode string) bool {
	for _, s := range topicDiscoveryPhasedStages {
		if s == stageCode {
			return stageCode != "topic_confirm"
		}
	}
	return false
}

func ideaDiscoveryPhaseIDToStageCode(phaseID string) string {
	switch strings.TrimSpace(phaseID) {
	case "research-lit":
		return "research_lit"
	case "idea-creator":
		return "idea_creator"
	case "novelty-check":
		return "novelty_check"
	case "research-review":
		return "research_review"
	case "research-refine-pipeline":
		return "research_refine"
	default:
		return strings.ReplaceAll(strings.TrimSpace(phaseID), "-", "_")
	}
}

// TopicStageCodeToPhaseID 供 bot / API 展示。
func TopicStageCodeToPhaseID(stageCode string) string {
	return topicStageCodeToPhaseID(stageCode)
}

func topicStageCodeToPhaseID(stageCode string) string {
	switch strings.TrimSpace(stageCode) {
	case "research_lit":
		return "research-lit"
	case "idea_creator":
		return "idea-creator"
	case "novelty_check":
		return "novelty-check"
	case "research_review":
		return "research-review"
	case "research_refine":
		return "research-refine-pipeline"
	default:
		return strings.ReplaceAll(stageCode, "_", "-")
	}
}

func topicStageBehaviorKey(stageCode string) string {
	return "topic.stage." + strings.TrimSpace(stageCode)
}

func nextPhasedTopicStageCode(stageCode string) string {
	for i, s := range topicDiscoveryPhasedStages {
		if s == stageCode && i+1 < len(topicDiscoveryPhasedStages) {
			return topicDiscoveryPhasedStages[i+1]
		}
	}
	return ""
}

func (s *TopicDiscoveryRunService) stagesForRun(ctx context.Context, userID uint, runVersion int) []string {
	step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "research_lit")
	if err == nil && step != nil {
		return topicDiscoveryPhasedStages
	}
	return topicDiscoveryStages
}

func fallbackStageSortOrder(stageCode string) int {
	for i, s := range topicDiscoveryPhasedStages {
		if s == stageCode {
			return i + 1
		}
	}
	for i, s := range topicDiscoveryStages {
		if s == stageCode {
			return i + 1
		}
	}
	return 0
}

func topicStepSortKey(step entity.PaperOutputTopicStep) int {
	if step.Sort > 0 {
		return step.Sort
	}
	return fallbackStageSortOrder(step.StageCode)
}

// SortTopicSteps 按 sort 升序；sort=0 的旧行按 stage 默认顺序。
func SortTopicSteps(rows []entity.PaperOutputTopicStep) []entity.PaperOutputTopicStep {
	out := append([]entity.PaperOutputTopicStep(nil), rows...)
	sort.SliceStable(out, func(i, j int) bool {
		ki, kj := topicStepSortKey(out[i]), topicStepSortKey(out[j])
		if ki != kj {
			return ki < kj
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func isPhasedPerStepRun(rows []entity.PaperOutputTopicStep) bool {
	for _, r := range rows {
		if r.StageCode == "research_lit" {
			return true
		}
	}
	return false
}

// PhasedTopicStageQueueStep 分 step 五段 bot：单条 ARIS phase step，status=running 且 result.async_llm。
func PhasedTopicStageQueueStep(step *entity.PaperOutputTopicStep) bool {
	if step == nil || strings.TrimSpace(step.Status) != "running" {
		return false
	}
	if !isPhasedTopicStage(step.StageCode) {
		return false
	}
	meta := paperstep.LoadStepExtra(step)
	async, ok := meta["async_llm"].(bool)
	return ok && async
}

// PhasedPauseAfterStage 段间人工检查点：上一段 completed、下一段 pending 时，返回待用户确认后启动的 stage_code（下一段）。
func PhasedPauseAfterStage(rows []entity.PaperOutputTopicStep, humanCheckpoint bool) string {
	if !humanCheckpoint {
		return ""
	}
	ordered := SortTopicSteps(rows)
	var phased []entity.PaperOutputTopicStep
	for _, r := range ordered {
		if isPhasedTopicStage(r.StageCode) {
			phased = append(phased, r)
		}
	}
	for i := 0; i < len(phased)-1; i++ {
		if strings.TrimSpace(phased[i].Status) == "completed" &&
			strings.TrimSpace(phased[i+1].Status) == "pending" {
			return phased[i+1].StageCode
		}
	}
	if len(phased) > 0 && strings.TrimSpace(phased[len(phased)-1].Status) == "completed" {
		for _, r := range ordered {
			if r.StageCode == "topic_confirm" && strings.TrimSpace(r.Status) == "pending" {
				return "topic_confirm"
			}
		}
	}
	return ""
}

// IdeaDiscoveryRuntimeStageQueueStep bot 扫描：idea_discovery running，且 runtime_state 为
// other.async_llm=true + stage 中有一段 status=executing（见 idea_discovery_runtime_state.go）。
func IdeaDiscoveryRuntimeStageQueueStep(step *entity.PaperOutputTopicStep) bool {
	if step == nil || strings.TrimSpace(step.Status) != "running" {
		return false
	}
	switch strings.TrimSpace(step.StageCode) {
	case "idea_discovery", "topic_discovery":
	default:
		return false
	}
	if !runtimeStateOtherAsyncLLM(step.RuntimeState) {
		return false
	}
	return ExecutingPhaseFromRuntimeState(step.RuntimeState) != ""
}

// StepUsesRuntimeStagePhasedInvoke 是否走五段 ARIS（非 bundled），以 runtime_state.stage 为准。
func StepUsesRuntimeStagePhasedInvoke(step *entity.PaperOutputTopicStep) bool {
	if step == nil {
		return false
	}
	switch strings.TrimSpace(step.StageCode) {
	case "idea_discovery", "topic_discovery":
	default:
		return false
	}
	return runtimeStateStageMap(step.RuntimeState) != nil
}

// TopicStageNeedsUserCheckpoint 该 ARIS phase 是否在 yaml 中 pause=true。
func TopicStageNeedsUserCheckpoint(phaseID string) bool {
	return mustIdeaDiscoveryConfig().PhaseNeedsUserCheckpoint(strings.TrimSpace(phaseID))
}

// TopicStageCheckpointDetail 通知文案。
func TopicStageCheckpointDetail(phaseID string) string {
	cfg := mustIdeaDiscoveryConfig()
	phaseID = strings.TrimSpace(phaseID)
	label := phaseID
	if cp, ok := cfg.CheckpointPhases[phaseID]; ok && strings.TrimSpace(cp.LabelZh) != "" {
		label = strings.TrimSpace(cp.LabelZh)
	}
	return fmt.Sprintf("Idea Discovery · %s 已完成，请确认后继续下一段", label)
}

func (s *TopicDiscoveryRunService) phasedStagesPriorCompleted(ctx context.Context, userID uint, runVersion int, stageCode string) (bool, error) {
	for _, code := range topicDiscoveryPhasedStages {
		if code == stageCode {
			return true, nil
		}
		if code == "topic_confirm" {
			break
		}
		step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, code)
		if err != nil {
			return false, err
		}
		if step == nil || strings.TrimSpace(step.Status) != "completed" {
			return false, nil
		}
	}
	return true, nil
}
