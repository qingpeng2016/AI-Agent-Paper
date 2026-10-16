package topicdiscovery

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/internal/jsonx"
	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

const resultKeyRetrieveSearchPending = "retrieve_search_pending"

// RetrieveSearchPending step.result 是否仍待 bot 执行第三方文献检索。
func RetrieveSearchPending(resultJSON []byte) bool {
	if len(resultJSON) == 0 {
		return true
	}
	var m map[string]any
	if json.Unmarshal(resultJSON, &m) != nil {
		return false
	}
	v, ok := m[resultKeyRetrieveSearchPending].(bool)
	return ok && v
}

// stageRetrieveScheduleForBot API 路径：只把 retrieve 标为 running，检索由 bot 执行。
func (s *TopicDiscoveryRunService) stageRetrieveScheduleForBot(ctx context.Context, step *entity.PaperOutputTopicStep, input topicRunInput) error {
	_ = ctx
	query := literatureSearchQueryFromInput(input)
	step.Result = jsonx.MustJSON(map[string]any{
		resultKeyRetrieveSearchPending: true,
		"search_query":                 query,
	})
	step.Status = "running"
	step.ExecProgressPct = 0
	return nil
}

// RunRetrieveLiteratureSearch bot：对 retrieve 步执行第三方检索并写入 result（不含 PDF 落盘）。
func (s *TopicDiscoveryRunService) RunRetrieveLiteratureSearch(ctx context.Context, step *entity.PaperOutputTopicStep) error {
	if s == nil || step == nil {
		return fmt.Errorf("retrieve search: nil service or step")
	}
	input, err := s.LoadRunInputForUser(ctx, uint(step.UserID), step.RunVersion)
	if err != nil {
		return err
	}
	if err := s.stageRetrieveExecuteSearch(ctx, step, input); err != nil {
		if step.StartedAt == nil {
			now := time.Now()
			step.StartedAt = &now
		}
		paperstep.FailTopicStep(step, err.Error())
		if saveErr := s.steps.SaveStep(ctx, step); saveErr != nil {
			return saveErr
		}
		return err
	}
	return s.steps.SaveStep(ctx, step)
}
