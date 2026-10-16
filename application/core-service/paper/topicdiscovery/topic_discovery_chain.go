package topicdiscovery

import (
	"context"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/internal/jsonx"
)

// skipRetrieveForNativeRun 原生模式：不跑平台检索，标记 retrieve 完成以便进入 idea_discovery。
func (s *TopicDiscoveryRunService) skipRetrieveForNativeRun(ctx context.Context, userID uint, runVersion int) error {
	cfg := mustIdeaDiscoveryConfig()
	if cfg.Platform.RequireRetrieveBeforeIdeaDiscovery {
		return nil
	}
	step, err := s.steps.GetStep(ctx, uint64(userID), runVersion, "retrieve")
	if err != nil || step == nil {
		return err
	}
	st := strings.TrimSpace(step.Status)
	if st == "completed" || st == "running" {
		return nil
	}
	sum := "原生 /idea-discovery：跳过平台 retrieve"
	step.SummaryText = &sum
	step.Status = "completed"
	step.ExecProgressPct = 100
	now := time.Now()
	step.CompletedAt = &now
	step.Result = jsonx.MustJSON(map[string]any{"skipped": true, "reason": "native_no_platform_retrieve"})
	return s.steps.SaveStep(ctx, step)
}

// OnRetrieveCompleted retrieve 完成后自动进入原生 /idea-discovery（需 auto_chain_retrieve 且未开平台级检查点）。
func (s *TopicDiscoveryRunService) OnRetrieveCompleted(ctx context.Context, userID uint, runVersion int) error {
	cfg := mustIdeaDiscoveryConfig()
	if cfg.Platform.InterStepHumanCheckpoint || !cfg.Platform.AutoChainRetrieve || !cfg.Platform.RequireRetrieveBeforeIdeaDiscovery {
		return nil
	}
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
