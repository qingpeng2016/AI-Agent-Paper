package topicdiscovery

import (
	"context"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/internal/jsonx"
)

// skipRetrieveForNativeRun 原生模式：不跑平台检索，标记 retrieve 完成以便进入 idea_discovery。
func (s *TopicDiscoveryRunService) skipRetrieveForNativeRun(ctx context.Context, userID uint, runVersion int) error {
	step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "retrieve")
	if err != nil || step == nil {
		return err
	}
	st := strings.TrimSpace(step.Status)
	if st == "completed" || st == "running" {
		return nil
	}
	step.Status = "completed"
	step.ExecProgressPct = 100
	now := time.Now()
	step.CompletedAt = &now
	step.Result = jsonx.MustJSON(map[string]any{
		"skipped": true,
		"reason":  "native_aris_web_literature",
		"note":    "文献由 /idea-discovery → /research-lit（sources: web）拉取并入库",
	})
	return s.steps.SaveStep(ctx, step)
}

// OnRetrieveCompleted retrieve 完成后若 idea_discovery 仍 pending 则启动（兼容旧平台 PDF bot 路径）。
func (s *TopicDiscoveryRunService) OnRetrieveCompleted(ctx context.Context, userID uint, runVersion int) error {
	idea, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "idea_discovery")
	if err != nil || idea == nil {
		return err
	}
	if strings.TrimSpace(idea.Status) != "pending" {
		return nil
	}
	ret, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "retrieve")
	if err != nil || ret == nil || strings.TrimSpace(ret.Status) != "completed" {
		return nil
	}
	return s.stageBeginTopicDiscoveryAsync(ctx, userID, runVersion, idea)
}
