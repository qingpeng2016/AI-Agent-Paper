package ideadiscoveryphases

import "context"

const stageResearchRefine = "research_refine"

// RunResearchRefine 仅处理 stage_code=research_refine（/research-refine-pipeline）。
func (j *IdeaDiscoveryPhasesJob) RunResearchRefine(ctx context.Context) {
	j.scanStage(ctx, stageResearchRefine)
}
