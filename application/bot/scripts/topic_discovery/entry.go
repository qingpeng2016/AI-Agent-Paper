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
	ideaDiscoveryMaxRetriesNotify = 32
)

type TopicDiscoveryJob struct {
	steps repository.PaperOutputTopicStepRepo
	runs  *papersvc.TopicDiscoveryRunService
}

func NewTopicDiscoveryJob(
	steps repository.PaperOutputTopicStepRepo,
	runs *papersvc.TopicDiscoveryRunService,
) *TopicDiscoveryJob {
	return &TopicDiscoveryJob{steps: steps, runs: runs}
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
		j.markFailed(ctx, step, "TopicDiscoveryRunService 未注入")
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
		j.markFailed(ctx, step, "无法加载 run 入参: "+err.Error())
		return "failed"
	}

	runID := strings.TrimSpace(fmt.Sprint(meta["aris_run_id"]))
	if runID == "" {
		runID = "(new)"
	}
	notification.SendNotifyLog(ctx, "topic-discovery-invoke-start",
		zap.Uint64("step_id", step.ID),
		zap.Uint64("manuscript_id", step.ManuscriptID),
		zap.Int("run_version", step.RunVersion),
		zap.String("aris_run_id", runID))

	beforePrompt, beforeCompletion := tokenUsageFromMeta(meta)

	if err := j.runs.RunNativeIdeaDiscoveryLLM(ctx, uint(step.UserID), step.RunVersion, step, input); err != nil {
		notification.SendErrorLog(ctx, "topic-discovery-llm-failed",
			zap.Uint64("step_id", step.ID),
			zap.Uint64("manuscript_id", step.ManuscriptID),
			zap.Int("run_version", step.RunVersion),
			zap.Error(err))
		if strings.TrimSpace(step.Status) != "failed" {
			j.markFailed(ctx, step, err.Error())
		}
		notification.SendNotifyLog(ctx, "topic-discovery-invoke-failed",
			zap.Uint64("step_id", step.ID),
			zap.String("reason", err.Error()))
		return "failed"
	}
	if strings.TrimSpace(step.Status) == "failed" {
		return "failed"
	}

	meta = papersvc.LoadStepExtra(step)
	if lastErr, ok := meta["aris_invoke_last_error"].(string); ok && strings.TrimSpace(lastErr) != "" {
		retries := intMeta(meta["aris_invoke_retries"])
		if retries == 1 || retries >= ideaDiscoveryMaxRetriesNotify {
			notification.SendErrorLog(ctx, "topic-discovery-aris-invoke-error",
				zap.Uint64("step_id", step.ID),
				zap.Uint64("manuscript_id", step.ManuscriptID),
				zap.Int("run_version", step.RunVersion),
				zap.Int("retry", retries),
				zap.String("error", strings.TrimSpace(lastErr)),
				zap.String("workspace_log", "inout/.../paper/.paper/invoke/logs/latest_invoke_error.txt"))
		}
	}
	deltaPrompt := intMeta(meta["tokens_prompt"]) - beforePrompt
	deltaCompletion := intMeta(meta["tokens_completion"]) - beforeCompletion
	if v, ok := meta["aris_awaiting_checkpoint"].(bool); ok && v {
		phase := strings.TrimSpace(fmt.Sprint(meta["aris_pause_after_phase"]))
		notification.SendNotifyLog(ctx, "topic-discovery-checkpoint",
			zap.Uint64("step_id", step.ID),
			zap.String("pause_after_phase", phase),
			zap.Int("prompt_tokens", deltaPrompt),
			zap.Int("completion_tokens", deltaCompletion))
		return "skipped"
	}
	if strings.TrimSpace(step.Status) != "completed" {
		notification.SendNotifyLog(ctx, "topic-discovery-invoke-progress",
			zap.Uint64("step_id", step.ID),
			zap.Int("exec_progress_pct", int(step.ExecProgressPct)),
			zap.Int("prompt_tokens", deltaPrompt),
			zap.Int("completion_tokens", deltaCompletion))
		return "skipped"
	}
	notification.SendNotifyLog(ctx, "topic-discovery-completed",
		zap.Uint64("step_id", step.ID),
		zap.Uint64("manuscript_id", step.ManuscriptID),
		zap.Int("prompt_tokens", deltaPrompt),
		zap.Int("completion_tokens", deltaCompletion))
	return "completed"
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
		notification.SendErrorLog(ctx, "topic-discovery-reconcile-save", zap.Error(err))
		return ""
	}
	return "completed"
}

func (j *TopicDiscoveryJob) markFailed(ctx context.Context, step *entity.PaperOutputTopicStep, reason string) {
	papersvc.FailTopicStep(step, reason)
	if err := j.steps.SaveStep(ctx, step); err != nil {
		notification.SendErrorLog(ctx, "topic-discovery-fail-save", zap.Uint64("step_id", step.ID), zap.Error(err))
	}
}
