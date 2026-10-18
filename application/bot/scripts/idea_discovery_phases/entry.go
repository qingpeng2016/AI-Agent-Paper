// idea_discovery_phases bot：五段 ARIS 各一条 paper_output_topic_step，按 stage_code 分脚本扫 running+async_llm。
package ideadiscoveryphases

import (
	"context"

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
		s, i, sk, f := j.dispatchStage(ctx, stageCode)
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

func (j *IdeaDiscoveryPhasesJob) dispatchStage(ctx context.Context, stageCode string) (scanned, invoked, skipped, failed int) {
	switch stageCode {
	case stageResearchLit:
		return j.scanStage(ctx, stageResearchLit)
	case stageIdeaCreator:
		return j.scanStage(ctx, stageIdeaCreator)
	case stageNoveltyCheck:
		return j.scanStage(ctx, stageNoveltyCheck)
	case stageResearchReview:
		return j.scanStage(ctx, stageResearchReview)
	case stageResearchRefine:
		return j.scanStage(ctx, stageResearchRefine)
	default:
		return 0, 0, 0, 0
	}
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
