package ideadiscoveryphases

import "context"

const stageResearchLit = "research_lit"

// RunResearchLit 仅处理 stage_code=research_lit 的 running 队列（/research-lit）。
func (j *IdeaDiscoveryPhasesJob) RunResearchLit(ctx context.Context) {
	j.scanStage(ctx, stageResearchLit)
}
