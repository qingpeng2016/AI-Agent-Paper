package topicaudit

import (
	"context"
	"strings"
	"time"

	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"go.uber.org/zap"
)

const (
	ModuleAIAgentPaper  = "ai_agent_paper"
	TaskTopicAuditLLM   = "topic_audit_llm"
	scanLimit           = 20
)

type TopicAuditJob struct {
	steps repository.PaperOutputTopicStepRepo
	runs  *papersvc.TopicDiscoveryRunService
}

func NewTopicAuditJob(
	steps repository.PaperOutputTopicStepRepo,
	runs *papersvc.TopicDiscoveryRunService,
) *TopicAuditJob {
	return &TopicAuditJob{steps: steps, runs: runs}
}

// Run 扫描 status=running 且 stage_code=audit，调用 LLM；成功 completed，失败 failed。
func (j *TopicAuditJob) Run(ctx context.Context) {
	rows, err := j.steps.ListByStatusAndStage(ctx, "running", "audit", scanLimit)
	if err != nil {
		logger.ErrorZ(ctx, "topic-audit-list-failed", zap.Error(err))
		return
	}
	if len(rows) == 0 {
		return
	}
	var completed, failed int
	for i := range rows {
		outcome := j.processOne(ctx, &rows[i])
		switch outcome {
		case "completed":
			completed++
		case "failed":
			failed++
		}
	}
	logger.InfoZ(ctx, "topic-audit-finished",
		zap.Int("scanned", len(rows)),
		zap.Int("completed", completed),
		zap.Int("failed", failed))
}

func (j *TopicAuditJob) processOne(ctx context.Context, step *entity.PaperOutputTopicStep) string {
	if j.runs == nil {
		j.markFailed(ctx, step, "TopicDiscoveryRunService 未注入")
		return "failed"
	}
	if len(step.Result) > 0 {
		return j.finishAlreadyDone(ctx, step)
	}

	input, err := j.runs.LoadRunInputForUser(ctx, uint(step.UserID), step.RunVersion)
	if err != nil {
		j.markFailed(ctx, step, "无法加载 run 入参: "+err.Error())
		return "failed"
	}

	if err := j.runs.RunAuditAndLiteratureReviewLLM(ctx, uint(step.UserID), step.RunVersion, step, input); err != nil {
		logger.WarnZ(ctx, "topic-audit-llm-failed",
			zap.Uint64("step_id", step.ID),
			zap.Error(err))
		j.markFailed(ctx, step, err.Error())
		return "failed"
	}
	return "completed"
}

func (j *TopicAuditJob) finishAlreadyDone(ctx context.Context, step *entity.PaperOutputTopicStep) string {
	if strings.TrimSpace(step.Status) == "completed" {
		return "completed"
	}
	now := time.Now()
	step.Status = "completed"
	step.CompletedAt = &now
	if err := j.steps.SaveStep(ctx, step); err != nil {
		logger.ErrorZ(ctx, "topic-audit-reconcile-save", zap.Error(err))
		return ""
	}
	return "completed"
}

func (j *TopicAuditJob) markFailed(ctx context.Context, step *entity.PaperOutputTopicStep, reason string) {
	meta := papersvc.LoadStepExtra(step)
	meta["error"] = strings.TrimSpace(reason)
	delete(meta, "async_llm")
	step.Extra = papersvc.MustJSONStep(meta)
	step.Status = "failed"
	now := time.Now()
	step.CompletedAt = &now
	sum := "审查与文献综述失败"
	step.SummaryText = &sum
	if err := j.steps.SaveStep(ctx, step); err != nil {
		logger.ErrorZ(ctx, "topic-audit-fail-save", zap.Uint64("step_id", step.ID), zap.Error(err))
	}
}
