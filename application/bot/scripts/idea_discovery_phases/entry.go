// idea_discovery_phases：entry 查库调度；各 stage 文件内完成 ARIS（调用 core-service/paper）。
package ideadiscoveryphases

import (
	"context"
	"strings"

	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"go.uber.org/zap"
)

const (
	TaskIdeaDiscoveryPhasesLLM = "llm_idea_discovery_phases"

	// paper_output_topic_step.status
	topicStepRunning   = "running"
	topicStepCompleted = "completed"
	topicStepFailed    = "failed"
	// process*Step 返回，非 DB status
	topicStepSkipped = "skipped"

	researchLitStageCode    = "research_lit"
	ideaCreatorStageCode    = "idea_creator"
	noveltyCheckStageCode   = "novelty_check"
	researchReviewStageCode = "research_review"
	researchRefineStageCode = "research_refine"
)

var (
	ideaCreatorPriorStages    = []string{researchLitStageCode}
	noveltyCheckPriorStages   = []string{researchLitStageCode, ideaCreatorStageCode}
	researchReviewPriorStages = []string{researchLitStageCode, ideaCreatorStageCode, noveltyCheckStageCode}
	researchRefinePriorStages = []string{researchLitStageCode, ideaCreatorStageCode, noveltyCheckStageCode, researchReviewStageCode}
)

// IdeaDiscoveryPhasesJob 五段分 step 的 ARIS bot 任务。
type IdeaDiscoveryPhasesJob struct {
	steps       repository.PaperOutputTopicStepRepo
	runs        *papersvc.TopicDiscoveryRunService
	manuscripts repository.PaperManuscriptRepo
}

// NewIdeaDiscoveryPhasesJob 构造 bot 任务。
func NewIdeaDiscoveryPhasesJob(
	steps repository.PaperOutputTopicStepRepo,
	runs *papersvc.TopicDiscoveryRunService,
	manuscripts repository.PaperManuscriptRepo,
) *IdeaDiscoveryPhasesJob {
	return &IdeaDiscoveryPhasesJob{steps: steps, runs: runs, manuscripts: manuscripts}
}

// Run 每个 user+文稿 只认全表最大 run_version，再在其中找 running 的五段队列并调度。
func (j *IdeaDiscoveryPhasesJob) Run(ctx context.Context) {
	rows, err := j.steps.ListByStatusOnMaxRunPerManuscript(ctx, topicStepRunning)
	if err != nil {
		logger.ErrorZ(ctx, "idea-discovery-phases-list-failed", zap.Error(err))
		return
	}
	for i := range rows {
		step := &rows[i]
		code := strings.TrimSpace(step.StageCode)

		var scanned, invoked, skipped, failed int
		switch code {
		case researchLitStageCode:
			scanned, invoked, skipped, failed = j.RunResearchLit(ctx, step)
		case ideaCreatorStageCode:
			scanned, invoked, skipped, failed = j.RunIdeaCreator(ctx, step)
		case noveltyCheckStageCode:
			scanned, invoked, skipped, failed = j.RunNoveltyCheck(ctx, step)
		case researchReviewStageCode:
			scanned, invoked, skipped, failed = j.RunResearchReview(ctx, step)
		case researchRefineStageCode:
			scanned, invoked, skipped, failed = j.RunResearchRefine(ctx, step)
		default:
			continue
		}
		if scanned == 0 {
			continue
		}
		logger.InfoZ(ctx, "idea-discovery-phases-finished",
			zap.String("stage_code", code),
			zap.Int("matched_queue_rows", scanned),
			zap.Int("completed", invoked),
			zap.Int("failed", failed),
			zap.Int("skipped", skipped))
		return
	}
}
