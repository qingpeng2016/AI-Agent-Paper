package ideadiscoveryphases

import "context"

const stageIdeaCreator = "idea_creator"

// RunIdeaCreator 仅处理 stage_code=idea_creator（/idea-creator）。
func (j *IdeaDiscoveryPhasesJob) RunIdeaCreator(ctx context.Context) {
	j.scanStage(ctx, stageIdeaCreator)
}
