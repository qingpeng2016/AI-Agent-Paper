package ideadiscoveryphases

import "context"

const stageNoveltyCheck = "novelty_check"

// RunNoveltyCheck 仅处理 stage_code=novelty_check（/novelty-check）。
func (j *IdeaDiscoveryPhasesJob) RunNoveltyCheck(ctx context.Context) {
	j.scanStage(ctx, stageNoveltyCheck)
}
