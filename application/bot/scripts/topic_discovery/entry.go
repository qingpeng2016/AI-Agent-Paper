package topicdiscovery

import (
	"context"
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
	steps repository.PaperOutputTopicStepRepo
	runs  *papersvc.TopicDiscoveryRunService
}

func NewTopicDiscoveryJob(
	steps repository.PaperOutputTopicStepRepo,
	runs *papersvc.TopicDiscoveryRunService,
) *TopicDiscoveryJob {
	return &TopicDiscoveryJob{steps: steps, runs: runs}
}

// Run 扫描 paper_output_topic_step 第二步（stage_code=topic_discovery, status=running），
// 经 paper_llm_strategy_route → 本地 ARIS CLI（skill）调用生成包与审计，结果写入该步 result。
func (j *TopicDiscoveryJob) Run(ctx context.Context) {
	rows, err := j.steps.ListByStatusAndStage(ctx, "running", "topic_discovery", scanLimit)
	if err != nil {
		notification.SendErrorLog(ctx, "topic-discovery-list-failed", zap.Error(err))
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
	logger.InfoZ(ctx, "topic-discovery-finished",
		zap.Int("scanned", len(rows)),
		zap.Int("completed", completed),
		zap.Int("failed", failed))
}

func (j *TopicDiscoveryJob) processOne(ctx context.Context, step *entity.PaperOutputTopicStep) string {
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

	if err := j.runs.RunTopicDiscoveryLLM(ctx, uint(step.UserID), step.RunVersion, step, input); err != nil {
		notification.SendErrorLog(ctx, "topic-discovery-llm-failed",
			zap.Uint64("step_id", step.ID),
			zap.Error(err))
		j.markFailed(ctx, step, err.Error())
		return "failed"
	}
	return "completed"
}

func (j *TopicDiscoveryJob) finishAlreadyDone(ctx context.Context, step *entity.PaperOutputTopicStep) string {
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
	meta := papersvc.LoadStepExtra(step)
	meta["error"] = strings.TrimSpace(reason)
	delete(meta, "async_llm")
	step.Extra = papersvc.MustJSONStep(meta)
	step.Status = "failed"
	step.ExecProgressPct = 0
	now := time.Now()
	step.CompletedAt = &now
	sum := "选题发现失败"
	step.SummaryText = &sum
	if err := j.steps.SaveStep(ctx, step); err != nil {
		notification.SendErrorLog(ctx, "topic-discovery-fail-save", zap.Uint64("step_id", step.ID), zap.Error(err))
	}
}
