// idea_discovery_phases bot：按 paper_output_topic_step 分段调用 ARIS（/research-lit → … → /research-refine-pipeline）。
// 与 llm_topic_discovery（单步 /idea-discovery）并列；run 需 invoke_mode=phased 创建五段 step。
package ideadiscoveryphases

import (
	"context"
	"fmt"
	"strings"

	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/topicdiscovery"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/common/notification"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"go.uber.org/zap"
)

const (
	ModuleAIAgentPaper           = "ai_agent_paper"
	TaskIdeaDiscoveryPhasesLLM   = "llm_idea_discovery_phases"
	scanLimitPerStage            = 8
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
	for _, stageCode := range topicdiscovery.TopicDiscoveryPhasedStageCodes() {
		rows, err := j.steps.ListByStatusAndStage(ctx, "running", stageCode, scanLimitPerStage)
		if err != nil {
			notification.SendErrorLog(ctx, "idea-discovery-phases-list-failed",
				zap.String("stage", stageCode), zap.Error(err))
			continue
		}
		for i := range rows {
			j.processOne(ctx, &rows[i])
		}
	}
}

func (j *IdeaDiscoveryPhasesJob) processOne(ctx context.Context, step *entity.PaperOutputTopicStep) string {
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
	if papersvc.TopicStepHasPersistedOutput(step) && strings.TrimSpace(step.Status) == "completed" {
		return "completed"
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

	if strings.TrimSpace(step.Status) == "completed" {
		j.notifyStep(ctx, step, papersvc.TopicStepOutcomeSuccess,
			fmt.Sprintf("%s 阶段已完成", step.StageCode), deltaPrompt, deltaCompletion)
		return "completed"
	}
	if strings.TrimSpace(step.Status) == "failed" {
		reason := ""
		if step.ErrorMsg != nil {
			reason = strings.TrimSpace(*step.ErrorMsg)
		}
		j.notifyStep(ctx, step, papersvc.TopicStepOutcomeFailed, reason, deltaPrompt, deltaCompletion)
		return "failed"
	}

	if phase := strings.TrimSpace(fmt.Sprint(meta["aris_pending_phase"])); phase != "" &&
		papersvc.IdeaDiscoveryNeedsUserConfirm(meta, phase) {
		j.notifyStep(ctx, step, papersvc.TopicStepOutcomeCheckpoint,
			papersvc.IdeaDiscoveryCheckpointDetail(phase), deltaPrompt, deltaCompletion)
		return "skipped"
	}

	logger.InfoZ(ctx, "idea-discovery-phases-invoke-tick-done",
		zap.Uint64("step_id", step.ID),
		zap.String("stage_code", step.StageCode),
		zap.Int("exec_progress_pct", int(step.ExecProgressPct)))
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
