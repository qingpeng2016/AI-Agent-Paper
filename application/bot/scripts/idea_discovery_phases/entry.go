// idea_discovery_phases bot：五段 ARIS 各一条 paper_output_topic_step，按 stage_code 扫 running+async_llm。
package ideadiscoveryphases

import (
	"context"
	"fmt"
	"strings"

	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/topicdiscovery"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"go.uber.org/zap"
)

const (
	ModuleAIAgentPaper         = "ai_agent_paper"
	TaskIdeaDiscoveryPhasesLLM = "llm_idea_discovery_phases"
	scanLimit                  = 20
)

type IdeaDiscoveryPhasesJob struct {
	steps       repository.PaperOutputTopicStepRepo
	runs        *papersvc.TopicDiscoveryRunService
	manuscripts repository.PaperManuscriptRepo
}

func NewIdeaDiscoveryPhasesJob(
	steps repository.PaperOutputTopicStepRepo,
	runs *papersvc.TopicDiscoveryRunService,
	manuscripts repository.PaperManuscriptRepo,
) *IdeaDiscoveryPhasesJob {
	return &IdeaDiscoveryPhasesJob{steps: steps, runs: runs, manuscripts: manuscripts}
}

func (j *IdeaDiscoveryPhasesJob) Run(ctx context.Context) {
	var scanned, invoked, skipped, failed int
	for _, stageCode := range topicdiscovery.TopicDiscoveryPhasedStageCodes() {
		s, i, sk, f := j.scanStage(ctx, stageCode)
		scanned += s
		invoked += i
		skipped += sk
		failed += f
	}
	if scanned > 0 {
		logger.InfoZ(ctx, "idea-discovery-phases-finished",
			zap.Int("matched_queue_rows", scanned),
			zap.Int("completed", invoked),
			zap.Int("failed", failed),
			zap.Int("skipped", skipped))
	}
}

func (j *IdeaDiscoveryPhasesJob) scanStage(ctx context.Context, stageCode string) (scanned, invoked, skipped, failed int) {
	rows, err := j.steps.ListByStatusAndStage(ctx, "running", stageCode, scanLimit)
	if err != nil {
		logger.ErrorZ(ctx, "idea-discovery-phases-list-failed",
			zap.String("stage_code", stageCode), zap.Error(err))
		return 0, 0, 0, 0
	}
	for i := range rows {
		if !topicdiscovery.PhasedTopicStageQueueStep(&rows[i]) {
			continue
		}
		scanned++
		switch j.processPhasedStep(ctx, &rows[i]) {
		case "completed":
			invoked++
		case "failed":
			failed++
		default:
			skipped++
		}
	}
	return scanned, invoked, skipped, failed
}

func (j *IdeaDiscoveryPhasesJob) processPhasedStep(ctx context.Context, step *entity.PaperOutputTopicStep) string {
	if step == nil || j.runs == nil {
		return "skipped"
	}
	if strings.TrimSpace(step.Status) == "failed" {
		return "skipped"
	}
	if step.ErrorMsg != nil && strings.TrimSpace(*step.ErrorMsg) != "" {
		if strings.TrimSpace(step.Status) != "failed" {
			papersvc.FailTopicStep(step, *step.ErrorMsg)
			_ = j.steps.SaveStep(ctx, step)
		}
		return "skipped"
	}

	meta := papersvc.LoadStepExtra(step)
	if async, ok := meta["async_llm"].(bool); !ok || !async {
		return "skipped"
	}

	input, err := j.runs.LoadRunInputForUser(ctx, uint(step.UserID), step.RunVersion)
	if err != nil {
		j.markFailed(ctx, step, "无法加载 run 入参: "+err.Error(), 0, 0)
		return "failed"
	}

	invoke, err := j.runs.PrepareTopicStageBotInvoke(ctx, uint(step.UserID), step.RunVersion, step, input)
	if err != nil {
		j.markFailed(ctx, step, err.Error(), 0, 0)
		return "failed"
	}
	if !invoke {
		return "skipped"
	}

	meta = papersvc.LoadStepExtra(step)
	beforePrompt := intMeta(meta["tokens_prompt"])
	beforeCompletion := intMeta(meta["tokens_completion"])

	if err = j.runs.RunNativeTopicStageLLM(ctx, uint(step.UserID), step.RunVersion, step, input); err != nil {
		if strings.TrimSpace(step.Status) != "failed" {
			j.markFailed(ctx, step, err.Error(), 0, 0)
		} else {
			meta = papersvc.LoadStepExtra(step)
			j.notifyStep(ctx, step, papersvc.TopicStepOutcomeFailed, err.Error(),
				intMeta(meta["tokens_prompt"])-beforePrompt, intMeta(meta["tokens_completion"])-beforeCompletion)
		}
		return "failed"
	}

	meta = papersvc.LoadStepExtra(step)
	deltaPrompt := intMeta(meta["tokens_prompt"]) - beforePrompt
	deltaCompletion := intMeta(meta["tokens_completion"]) - beforeCompletion
	phaseID := topicdiscovery.TopicStageCodeToPhaseID(step.StageCode)

	if strings.TrimSpace(step.Status) == "completed" {
		j.notifyStep(ctx, step, papersvc.TopicStepOutcomeSuccess,
			fmt.Sprintf("Idea Discovery · %s 已完成", phaseID), deltaPrompt, deltaCompletion)
		if topicdiscovery.TopicStageNeedsUserCheckpoint(phaseID) {
			j.notifyStep(ctx, step, papersvc.TopicStepOutcomeCheckpoint,
				topicdiscovery.TopicStageCheckpointDetail(phaseID), deltaPrompt, deltaCompletion)
		}
		return "completed"
	}
	if strings.TrimSpace(step.Status) == "failed" {
		j.notifyStep(ctx, step, papersvc.TopicStepOutcomeFailed,
			fmt.Sprintf("Idea Discovery · %s 失败", phaseID), deltaPrompt, deltaCompletion)
		return "failed"
	}

	logger.InfoZ(ctx, "idea-discovery-phases-invoke-tick",
		zap.Uint64("step_id", step.ID),
		zap.String("stage_code", step.StageCode),
		zap.String("phase_id", phaseID))
	return "skipped"
}

func (j *IdeaDiscoveryPhasesJob) notifyStep(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	outcome, detail string,
	promptTokens, completionTokens int,
) {
	if step == nil {
		return
	}
	papersvc.NotifyTopicStepExecutionResult(ctx, j.manuscripts, papersvc.TopicStepExecNotify{
		StageCode:        step.StageCode,
		Outcome:          outcome,
		UserID:           uint(step.UserID),
		ManuscriptID:     step.ManuscriptID,
		RunVersion:       step.RunVersion,
		StepID:           step.ID,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		Detail:           detail,
	})
}

func intMeta(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	default:
		return 0
	}
}

func (j *IdeaDiscoveryPhasesJob) markFailed(ctx context.Context, step *entity.PaperOutputTopicStep, reason string, promptTokens, completionTokens int) {
	papersvc.FailTopicStep(step, reason)
	if err := j.steps.SaveStep(ctx, step); err != nil {
		logger.ErrorZ(ctx, "idea-discovery-phases-fail-save", zap.Uint64("step_id", step.ID), zap.Error(err))
		return
	}
	j.notifyStep(ctx, step, papersvc.TopicStepOutcomeFailed, reason, promptTokens, completionTokens)
}
