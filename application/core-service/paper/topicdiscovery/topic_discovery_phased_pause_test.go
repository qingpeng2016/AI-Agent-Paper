package topicdiscovery

import (
	"testing"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

func TestPhasedPauseAfterStage_returnsNextPending(t *testing.T) {
	rows := []entity.PaperOutputTopicStep{
		{StageCode: "research_lit", Sort: 1, Status: "completed"},
		{StageCode: "idea_creator", Sort: 2, Status: "pending"},
		{StageCode: "novelty_check", Sort: 3, Status: "pending"},
	}
	got := PhasedPauseAfterStage(rows, true)
	if got != "idea_creator" {
		t.Fatalf("pause stage = %q, want idea_creator", got)
	}
}
