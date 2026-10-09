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
	ModuleAIAgentPaper            = "ai_agent_paper"
	TaskTopicLiteratureDownload   = "topic_literature_pdf_download"
	scanLimit                     = 50
)

type TopicLiteratureDownloadJob struct {
	steps repository.PaperOutputTopicStepRepo
}

func NewTopicLiteratureDownloadJob(steps repository.PaperOutputTopicStepRepo) *TopicLiteratureDownloadJob {
	return &TopicLiteratureDownloadJob{steps: steps}
}

// Run 扫描 status=running 且 stage_code=retrieve，下载 PDF 并更新 extra.local_path。
func (j *TopicLiteratureDownloadJob) Run(ctx context.Context) {
	rows, err := j.steps.ListByStatusAndStage(ctx, "running", "retrieve", scanLimit)
	if err != nil {
		logger.ErrorZ(ctx, "topic-literature-download-list-failed", zap.Error(err))
		return
	}
	if len(rows) == 0 {
		return
	}
	var completed int
	for i := range rows {
		if j.processOne(ctx, &rows[i]) {
			completed++
		}
	}
	logger.InfoZ(ctx, "topic-literature-download-finished",
		zap.Int("scanned", len(rows)),
		zap.Int("completed", completed))
}

func (j *TopicLiteratureDownloadJob) processOne(ctx context.Context, step *entity.PaperOutputTopicStep) bool {
	items := papersvc.ParseLiteratureDownloads(step.Extra)
	changed := false
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
			changed = true
		}
		abs := papersvc.LiteratureLocalAbsPath(rel)
		if st, err := os.Stat(abs); err == nil && st.Size() > 0 {
			continue
		}
		if err := papersvc.DownloadLiteraturePDF(ctx, pdfURL, abs); err != nil {
			logger.WarnZ(ctx, "topic-literature-download-pdf-failed",
				zap.String("external_key", it.ExternalKey),
				zap.String("pdf_url", pdfURL),
				zap.Error(err))
			continue
		}
		changed = true
	}
	if changed {
		_ = j.writeExtra(ctx, step, items)
	}
	if !papersvc.LiteratureDownloadsReady(items) {
		return false
	}
	return j.markCompleted(ctx, step, items)
}

func (j *TopicLiteratureDownloadJob) writeExtra(ctx context.Context, step *entity.PaperOutputTopicStep, items []papersvc.LiteratureDownloadItem) error {
	meta := map[string]any{}
	if len(step.Extra) > 0 {
		_ = json.Unmarshal(step.Extra, &meta)
	}
	meta["literature_downloads"] = items
	step.Extra = mustJSON(meta)
	return j.steps.SaveStep(ctx, step)
}

func (j *TopicLiteratureDownloadJob) markCompleted(ctx context.Context, step *entity.PaperOutputTopicStep, items []papersvc.LiteratureDownloadItem) bool {
	meta := map[string]any{}
	if len(step.Extra) > 0 {
		_ = json.Unmarshal(step.Extra, &meta)
	}
	meta["literature_downloads"] = items
	hitCount := 0
	if n, ok := meta["hit_count"].(float64); ok {
		hitCount = int(n)
	}
	step.Extra = mustJSON(meta)
	step.Status = "completed"
	now := time.Now()
	step.CompletedAt = &now
	sum := fmt.Sprintf("检索完成，共 %d 篇文献（PDF 已落盘）", hitCount)
	step.SummaryText = &sum
	if err := j.steps.SaveStep(ctx, step); err != nil {
		logger.ErrorZ(ctx, "topic-literature-download-save-failed", zap.Uint64("step_id", step.ID), zap.Error(err))
		return false
	}
	return true
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
