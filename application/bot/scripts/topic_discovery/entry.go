package topicdiscovery

import (
	"context"
	"fmt"
	"strings"
	"time"

	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/common/notification"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"go.uber.org/zap"
)

const (
	ModuleAIAgentPaper    = "ai_agent_paper"
	TaskTopicDiscoveryLLM = "llm_topic_discovery"
	scanLimit             = 20
)

type TopicDiscoveryJob struct {
	steps       repository.PaperOutputTopicStepRepo
	runs        *papersvc.TopicDiscoveryRunService
	manuscripts repository.PaperManuscriptRepo
}

func NewTopicDiscoveryJob(
	steps repository.PaperOutputTopicStepRepo,
	runs *papersvc.TopicDiscoveryRunService,
	manuscripts repository.PaperManuscriptRepo,
) *TopicDiscoveryJob {
	return &TopicDiscoveryJob{steps: steps, runs: runs, manuscripts: manuscripts}
}

// Run 扫描 idea_discovery（status=running 且 async_llm），每轮至多一次 ARIS invoke（auto_proceed=false + 检查点）。
func (j *TopicDiscoveryJob) Run(ctx context.Context) {
	rows, err := j.steps.ListByStatusAndStage(ctx, "running", "idea_discovery", scanLimit)
	if err != nil {
		notification.SendErrorLog(ctx, "topic-discovery-list-failed", zap.Error(err))
		return
	}
	if len(rows) == 0 {
		return
	}
	var completed, failed, skipped int
	for i := range rows {
		outcome := j.processOne(ctx, &rows[i])
		switch outcome {
		case "completed":
			completed++
		case "failed":
			failed++
		case "skipped":
			skipped++
		}
	}
	logger.InfoZ(ctx, "topic-discovery-finished",
		zap.Int("scanned", len(rows)),
		zap.Int("completed", completed),
		zap.Int("failed", failed),
		zap.Int("skipped", skipped))
}

func (j *TopicDiscoveryJob) processOne(ctx context.Context, step *entity.PaperOutputTopicStep) string {
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
	if j.runs == nil {
		j.markFailed(ctx, step, "TopicDiscoveryRunService 未注入", 0, 0)
		return "failed"
	}
	if papersvc.TopicStepHasPersistedOutput(step) {
		return j.finishAlreadyDone(ctx, step)
	}
	meta := papersvc.LoadStepExtra(step)
	if async, ok := meta["async_llm"].(bool); !ok || !async {
		return "skipped"
	}
	if v, ok := meta["aris_awaiting_checkpoint"].(bool); ok && v {
		return "skipped"
	}

	input, err := j.runs.LoadRunInputForUser(ctx, uint(step.UserID), step.RunVersion)
	if err != nil {
		j.markFailed(ctx, step, "无法加载 run 入参: "+err.Error(), 0, 0)
		return "failed"
	}

	beforePrompt, beforeCompletion := tokenUsageFromMeta(meta)

	if err := j.runs.RunNativeIdeaDiscoveryLLM(ctx, uint(step.UserID), step.RunVersion, step, input); err != nil {
		if strings.TrimSpace(step.Status) != "failed" {
			j.markFailed(ctx, step, err.Error(), 0, 0)
		} else {
			meta = papersvc.LoadStepExtra(step)
			deltaPrompt := intMeta(meta["tokens_prompt"]) - beforePrompt
			deltaCompletion := intMeta(meta["tokens_completion"]) - beforeCompletion
			j.notifyStep(ctx, step, papersvc.TopicStepOutcomeFailed, err.Error(), deltaPrompt, deltaCompletion)
		}
		return "failed"
	}
	if strings.TrimSpace(step.Status) == "failed" {
		meta = papersvc.LoadStepExtra(step)
		reason := ""
		if step.ErrorMsg != nil {
			reason = strings.TrimSpace(*step.ErrorMsg)
		}
		deltaPrompt := intMeta(meta["tokens_prompt"]) - beforePrompt
		deltaCompletion := intMeta(meta["tokens_completion"]) - beforeCompletion
		j.notifyStep(ctx, step, papersvc.TopicStepOutcomeFailed, reason, deltaPrompt, deltaCompletion)
		return "failed"
	}

	meta = papersvc.LoadStepExtra(step)
	deltaPrompt := intMeta(meta["tokens_prompt"]) - beforePrompt
	deltaCompletion := intMeta(meta["tokens_completion"]) - beforeCompletion

	if v, ok := meta["aris_awaiting_checkpoint"].(bool); ok && v {
		phase := strings.TrimSpace(fmt.Sprint(meta["aris_pause_after_phase"]))
		j.notifyStep(ctx, step, papersvc.TopicStepOutcomeCheckpoint, papersvc.IdeaDiscoveryCheckpointDetail(phase), deltaPrompt, deltaCompletion)
		return "skipped"
	}
	if strings.TrimSpace(step.Status) == "completed" {
		j.notifyStep(ctx, step, papersvc.TopicStepOutcomeSuccess, "Idea Discovery 全流程已完成", deltaPrompt, deltaCompletion)
		return "completed"
	}
	// 本 tick 已跑完一次 invoke，仍排队后续段：只打日志，不推 Lark（避免逐步刷屏）
	logger.InfoZ(ctx, "topic-discovery-invoke-tick-done",
		zap.Uint64("step_id", step.ID),
		zap.Uint64("manuscript_id", step.ManuscriptID),
		zap.Int("exec_progress_pct", int(step.ExecProgressPct)),
		zap.Int("prompt_tokens_delta", deltaPrompt),
		zap.Int("completion_tokens_delta", deltaCompletion))
	return "skipped"
}

func (j *TopicDiscoveryJob) notifyStep(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	outcome, detail string,
	promptTokens, completionTokens int,
) {
	if step == nil {
		return
	}
	papersvc.NotifyTopicStepExecutionResult(ctx, j.manuscripts, papersvc.TopicStepExecNotify{
		StageCode:        "idea_discovery",
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

func tokenUsageFromMeta(meta map[string]any) (prompt, completion int) {
	return intMeta(meta["tokens_prompt"]), intMeta(meta["tokens_completion"])
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

func (j *TopicDiscoveryJob) finishAlreadyDone(ctx context.Context, step *entity.PaperOutputTopicStep) string {
	if strings.TrimSpace(step.Status) == "failed" {
		return "skipped"
	}
	if strings.TrimSpace(step.Status) == "completed" {
		return "completed"
	}
	now := time.Now()
	step.Status = "completed"
	step.CompletedAt = &now
	if err := j.steps.SaveStep(ctx, step); err != nil {
		logger.ErrorZ(ctx, "topic-discovery-reconcile-save", zap.Error(err))
		return ""
	}
	return "completed"
}

func (j *TopicDiscoveryJob) markFailed(ctx context.Context, step *entity.PaperOutputTopicStep, reason string, promptTokens, completionTokens int) {
	papersvc.FailTopicStep(step, reason)
	if err := j.steps.SaveStep(ctx, step); err != nil {
		logger.ErrorZ(ctx, "topic-discovery-fail-save", zap.Uint64("step_id", step.ID), zap.Error(err))
		return
	}
	j.notifyStep(ctx, step, papersvc.TopicStepOutcomeFailed, reason, promptTokens, completionTokens)
}
