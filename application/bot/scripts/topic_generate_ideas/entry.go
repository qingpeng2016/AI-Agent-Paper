package topicgenerateideas

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
	ModuleAIAgentPaper       = "ai_agent_paper"
	TaskTopicGenerateIdeasLLM = "topic_generate_ideas_llm"
	scanLimit                = 20
)

type TopicGenerateIdeasJob struct {
	steps repository.PaperOutputTopicStepRepo
	runs  *papersvc.TopicDiscoveryRunService
}

func NewTopicGenerateIdeasJob(
	steps repository.PaperOutputTopicStepRepo,
	runs *papersvc.TopicDiscoveryRunService,
) *TopicGenerateIdeasJob {
	return &TopicGenerateIdeasJob{steps: steps, runs: runs}
}

// Run 扫描 status=running 且 stage_code=generate_ideas，调用 LLM；成功 completed，失败 failed。
func (j *TopicGenerateIdeasJob) Run(ctx context.Context) {
	rows, err := j.steps.ListByStatusAndStage(ctx, "running", "generate_ideas", scanLimit)
	if err != nil {
		logger.ErrorZ(ctx, "topic-generate-ideas-list-failed", zap.Error(err))
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
	logger.InfoZ(ctx, "topic-generate-ideas-finished",
		zap.Int("scanned", len(rows)),
		zap.Int("completed", completed),
		zap.Int("failed", failed))
}

func (j *TopicGenerateIdeasJob) processOne(ctx context.Context, step *entity.PaperOutputTopicStep) string {
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

	if err := j.runs.RunIdeasAndNoveltyLLM(ctx, uint(step.UserID), step.RunVersion, step, input); err != nil {
		logger.WarnZ(ctx, "topic-generate-ideas-llm-failed",
			zap.Uint64("step_id", step.ID),
			zap.Error(err))
		j.markFailed(ctx, step, err.Error())
		return "failed"
	}
	return "completed"
}

func (j *TopicGenerateIdeasJob) finishAlreadyDone(ctx context.Context, step *entity.PaperOutputTopicStep) string {
	if strings.TrimSpace(step.Status) == "completed" {
		return "completed"
	}
	now := time.Now()
	step.Status = "completed"
	step.CompletedAt = &now
	if err := j.steps.SaveStep(ctx, step); err != nil {
		logger.ErrorZ(ctx, "topic-generate-ideas-reconcile-save", zap.Error(err))
		return ""
	}
	return "completed"
}

func (j *TopicGenerateIdeasJob) markFailed(ctx context.Context, step *entity.PaperOutputTopicStep, reason string) {
	meta := papersvc.LoadStepExtra(step)
	meta["error"] = strings.TrimSpace(reason)
	delete(meta, "async_llm")
	step.Extra = papersvc.MustJSONStep(meta)
	step.Status = "failed"
	now := time.Now()
	step.CompletedAt = &now
	sum := "脑暴与新颖性分析失败"
	step.SummaryText = &sum
	if err := j.steps.SaveStep(ctx, step); err != nil {
		logger.ErrorZ(ctx, "topic-generate-ideas-fail-save", zap.Uint64("step_id", step.ID), zap.Error(err))
	}
}
