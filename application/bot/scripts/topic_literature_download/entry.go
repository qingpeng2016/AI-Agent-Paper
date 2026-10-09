package topicliteraturedownload

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"go.uber.org/zap"
)

const (
	ModuleAIAgentPaper          = "ai_agent_paper"
	TaskTopicLiteratureDownload = "topic_literature_pdf_download"
	scanLimit                   = 50
)

type TopicLiteratureDownloadJob struct {
	steps repository.PaperOutputTopicStepRepo
}

func NewTopicLiteratureDownloadJob(steps repository.PaperOutputTopicStepRepo) *TopicLiteratureDownloadJob {
	return &TopicLiteratureDownloadJob{steps: steps}
}

// Run 扫描 status=running 且 stage_code=retrieve，下载 PDF；成功 completed，失败 failed。
func (j *TopicLiteratureDownloadJob) Run(ctx context.Context) {
	rows, err := j.steps.ListByStatusAndStage(ctx, "running", "retrieve", scanLimit)
	if err != nil {
		logger.ErrorZ(ctx, "topic-literature-download-list-failed", zap.Error(err))
		return
	}
	if len(rows) == 0 {
		return
	}
	var completed, failed int
	for i := range rows {
		outcome := j.processOne(ctx, &rows[i])
		switch outcome {
		case "completed":
			completed++
		case "failed":
			failed++
		}
	}
	logger.InfoZ(ctx, "topic-literature-download-finished",
		zap.Int("scanned", len(rows)),
		zap.Int("completed", completed),
		zap.Int("failed", failed))
}

func (j *TopicLiteratureDownloadJob) processOne(ctx context.Context, step *entity.PaperOutputTopicStep) string {
	if len(step.Result) == 0 {
		j.markFailed(ctx, step, nil, "retrieve 无 result，无法执行 PDF 下载")
		return "failed"
	}

	items := papersvc.ParseLiteratureDownloads(step.Extra)
	if items == nil {
		items = []papersvc.LiteratureDownloadItem{}
	}
	hitCount := hitCountFromExtra(step.Extra)
	if hitCount > 0 && len(items) == 0 {
		j.markFailed(ctx, step, items, "缺少 literature_downloads，无法下载 PDF")
		return "failed"
	}

	if !papersvc.LiteratureDownloadsNeedWork(items) {
		return j.finishSuccess(ctx, step, items)
	}

	for idx := range items {
		it := &items[idx]
		pdfURL := strings.TrimSpace(it.PdfURL)
		if pdfURL == "" {
			continue
		}
		rel := strings.TrimSpace(it.LocalPath)
		if rel == "" {
			rel = papersvc.LiteratureLocalRelPath(it.ExternalKey)
			it.LocalPath = rel
		}
		abs := papersvc.LiteratureLocalAbsPath(rel)
		if st, err := os.Stat(abs); err == nil && st.Size() > 0 {
			it.DownloadError = ""
			continue
		}
		if err := papersvc.DownloadLiteraturePDF(ctx, pdfURL, abs); err != nil {
			it.DownloadError = err.Error()
			logger.WarnZ(ctx, "topic-literature-download-pdf-failed",
				zap.String("external_key", it.ExternalKey),
				zap.String("pdf_url", pdfURL),
				zap.Error(err))
			continue
		}
		it.DownloadError = ""
	}

	_ = j.mergeExtra(ctx, step, items, "")

	if papersvc.LiteratureDownloadsReady(items) {
		return j.finishSuccess(ctx, step, items)
	}
	reason := papersvc.LiteratureDownloadFailureReason(items)
	if reason == "" {
		reason = "PDF 下载未完成"
	}
	j.markFailed(ctx, step, items, reason)
	return "failed"
}

func (j *TopicLiteratureDownloadJob) mergeExtra(ctx context.Context, step *entity.PaperOutputTopicStep, items []papersvc.LiteratureDownloadItem, errMsg string) error {
	meta := map[string]any{}
	if len(step.Extra) > 0 {
		_ = json.Unmarshal(step.Extra, &meta)
	}
	meta["literature_downloads"] = items
	if strings.TrimSpace(errMsg) != "" {
		meta["error"] = errMsg
	}
	step.Extra = mustJSON(meta)
	return j.steps.SaveStep(ctx, step)
}

func (j *TopicLiteratureDownloadJob) finishSuccess(ctx context.Context, step *entity.PaperOutputTopicStep, items []papersvc.LiteratureDownloadItem) string {
	meta := map[string]any{}
	if len(step.Extra) > 0 {
		_ = json.Unmarshal(step.Extra, &meta)
	}
	meta["literature_downloads"] = items
	delete(meta, "error")
	step.Extra = mustJSON(meta)

	hitCount := hitCountFromMeta(meta)
	step.Status = "completed"
	now := time.Now()
	step.CompletedAt = &now
	sum := fmt.Sprintf("检索完成，共 %d 篇文献（PDF 已落盘）", hitCount)
	step.SummaryText = &sum
	if err := j.steps.SaveStep(ctx, step); err != nil {
		logger.ErrorZ(ctx, "topic-literature-download-save-failed", zap.Uint64("step_id", step.ID), zap.Error(err))
		return ""
	}
	return "completed"
}

func (j *TopicLiteratureDownloadJob) markFailed(ctx context.Context, step *entity.PaperOutputTopicStep, items []papersvc.LiteratureDownloadItem, reason string) {
	meta := map[string]any{}
	if len(step.Extra) > 0 {
		_ = json.Unmarshal(step.Extra, &meta)
	}
	if items != nil {
		meta["literature_downloads"] = items
	}
	meta["error"] = strings.TrimSpace(reason)
	step.Extra = mustJSON(meta)
	step.Status = "failed"
	now := time.Now()
	step.CompletedAt = &now
	sum := "文献 PDF 下载失败"
	step.SummaryText = &sum
	if err := j.steps.SaveStep(ctx, step); err != nil {
		logger.ErrorZ(ctx, "topic-literature-download-fail-save", zap.Uint64("step_id", step.ID), zap.Error(err))
	}
}

func hitCountFromMeta(meta map[string]any) int {
	if n, ok := meta["hit_count"].(float64); ok {
		return int(n)
	}
	return 0
}

func hitCountFromExtra(extra []byte) int {
	if len(extra) == 0 {
		return 0
	}
	var meta map[string]any
	if json.Unmarshal(extra, &meta) != nil {
		return 0
	}
	return hitCountFromMeta(meta)
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
