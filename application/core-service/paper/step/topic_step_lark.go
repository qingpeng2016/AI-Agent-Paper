package step

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/common/notification"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"go.uber.org/zap"
)

const (
	TopicStepOutcomeSuccess    = "success"
	TopicStepOutcomeFailed     = "failed"
	TopicStepOutcomeCheckpoint = "checkpoint"
)

// TopicStepExecNotify 选题流程单步执行结束后的 Lark 摘要（一步一条）。
type TopicStepExecNotify struct {
	StageCode        string
	Outcome          string // success | failed | checkpoint
	UserID           uint
	ManuscriptID     uint64
	ManuscriptTitle  string
	RunVersion       int
	StepID           uint64
	PromptTokens     int
	CompletionTokens int
	Detail           string
}

// TopicStageLabelZh 步骤 stage_code 展示名。
func TopicStageLabelZh(stageCode string) string {
	switch strings.TrimSpace(stageCode) {
	case "retrieve":
		return "文献检索与 PDF 下载"
	case "idea_discovery", "topic_discovery":
		return "选题发现-全流程"
	case "topic_confirm":
		return "确认并完成"
	default:
		if s := strings.TrimSpace(stageCode); s != "" {
			return s
		}
		return "未知步骤"
	}
}

func arisPhaseLabelZh(phase string) string {
	switch strings.TrimSpace(phase) {
	case "research-lit":
		return "文献调研"
	case "idea-creator":
		return "论文大纲"
	case "novelty-check":
		return "新颖性检验"
	case "research-review":
		return "外部审查"
	case "research-refine-pipeline":
		return "文献综述与实验规划"
	default:
		return phase
	}
}

// arisPhaseUiLabelZh 工作台 / 检查点对外展示（文献调研 + 论文大纲 合并为一步）。
func arisPhaseUiLabelZh(phase string) string {
	switch strings.TrimSpace(phase) {
	case "research-lit", "idea-creator":
		return "文献调研与选题大纲"
	case "novelty-check":
		return "新颖性检验"
	case "research-review":
		return "外部审查"
	case "research-refine-pipeline":
		return "文献综述与实验规划"
	default:
		return arisPhaseLabelZh(phase)
	}
}

// NotifyTopicStepExecutionResult 步骤执行结束后推送一条 Lark（成功 / 失败 / 检查点暂停）。
func NotifyTopicStepExecutionResult(
	ctx context.Context,
	manuscripts repository.PaperManuscriptRepo,
	n TopicStepExecNotify,
) {
	n.StageCode = strings.TrimSpace(n.StageCode)
	n.Outcome = strings.TrimSpace(n.Outcome)
	if n.ManuscriptTitle == "" && manuscripts != nil && n.ManuscriptID > 0 && n.UserID > 0 {
		if ms, err := manuscripts.GetByIDForUser(ctx, uint(n.ManuscriptID), n.UserID); err == nil && ms != nil {
			n.ManuscriptTitle = strings.TrimSpace(ms.Title)
		}
	}

	stageLabel := TopicStageLabelZh(n.StageCode)
	resultLabel, logMsg := topicStepOutcomeLabels(n.Outcome, stageLabel)

	total := n.PromptTokens + n.CompletionTokens
	userLine := fmt.Sprintf("user_id=%d", n.UserID)
	msLine := fmt.Sprintf("manuscript_id=%d", n.ManuscriptID)
	if t := strings.TrimSpace(n.ManuscriptTitle); t != "" {
		msLine = fmt.Sprintf("manuscript_id=%d · %s", n.ManuscriptID, t)
	}

	fields := []notification.FieldPair{
		{Key: "步骤", Value: stageLabel},
		{Key: "结果", Value: resultLabel},
		{Key: "用户", Value: userLine},
		{Key: "文章", Value: msLine},
		{Key: "run_version", Value: strconv.Itoa(n.RunVersion)},
		{Key: "step_id", Value: strconv.FormatUint(n.StepID, 10)},
		{Key: "Token", Value: fmt.Sprintf("prompt %d + completion %d = %d", n.PromptTokens, n.CompletionTokens, total)},
	}
	if d := strings.TrimSpace(n.Detail); d != "" {
		fields = append(fields, notification.FieldPair{Key: "说明", Value: d})
	}

	logger.InfoZ(ctx, logMsg,
		zap.String("stage_code", n.StageCode),
		zap.String("outcome", n.Outcome),
		zap.Uint("user_id", n.UserID),
		zap.Uint64("manuscript_id", n.ManuscriptID),
		zap.Int("prompt_tokens", n.PromptTokens),
		zap.Int("completion_tokens", n.CompletionTokens))

	nm := notification.GetGlobalNotificationManager()
	switch n.Outcome {
	case TopicStepOutcomeFailed:
		_ = nm.SendErrorAlertWithOrder(ctx, logMsg, fields)
	default:
		_ = nm.SendInfoAlert(ctx, logMsg, fields)
	}
}

func topicStepOutcomeLabels(outcome, stageLabel string) (resultLabel, logMsg string) {
	switch outcome {
	case TopicStepOutcomeSuccess:
		return "成功", fmt.Sprintf("选题流程 · %s · 成功", stageLabel)
	case TopicStepOutcomeFailed:
		return "失败", fmt.Sprintf("选题流程 · %s · 失败", stageLabel)
	case TopicStepOutcomeCheckpoint:
		return "等待人工确认", fmt.Sprintf("选题流程 · %s · 检查点", stageLabel)
	default:
		return outcome, fmt.Sprintf("选题流程 · %s · %s", stageLabel, outcome)
	}
}

// IdeaDiscoveryCheckpointDetail 段间检查点说明（含 ARIS 子阶段中文名）。
func IdeaDiscoveryCheckpointDetail(pauseAfterPhase string) string {
	p := strings.TrimSpace(pauseAfterPhase)
	if p == "" {
		return "Idea Discovery 段间检查点，待用户在平台确认并继续"
	}
	return fmt.Sprintf("Idea Discovery 已完成本轮 invoke，下一阶段「%s」（%s）需在平台确认后继续", arisPhaseUiLabelZh(p), p)
}
