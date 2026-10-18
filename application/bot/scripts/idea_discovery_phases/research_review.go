package ideadiscoveryphases

import "context"

const stageResearchReview = "research_review"

// RunResearchReview 仅处理 stage_code=research_review（/research-review）。
func (j *IdeaDiscoveryPhasesJob) RunResearchReview(ctx context.Context) {
	j.scanStage(ctx, stageResearchReview)
}
