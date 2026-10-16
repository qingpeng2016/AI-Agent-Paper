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
	"github.com/qingpeng2016/ai-agent-paper/common/notification"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"go.uber.org/zap"
)

const (
	ModuleAIAgentPaper          = "ai_agent_paper"
	TaskTopicLiteratureDownload = "topic_literature_pdf_download"
	scanLimit                   = 50
	pdfDownloadMaxAttempts      = 5
	pdfDownloadRetryDelay       = time.Second
)

type TopicLiteratureDownloadJob struct {
	steps    repository.PaperOutputTopicStepRepo
	userLit  repository.PaperUserLiteratureRepo
	runs     *papersvc.TopicDiscoveryRunService
}

func NewTopicLiteratureDownloadJob(
	steps repository.PaperOutputTopicStepRepo,
	userLit repository.PaperUserLiteratureRepo,
	runs *papersvc.TopicDiscoveryRunService,
) *TopicLiteratureDownloadJob {
	return &TopicLiteratureDownloadJob{steps: steps, userLit: userLit, runs: runs}
}

// Run 扫描 status=running 且 stage_code=retrieve，下载 PDF；成功 completed，失败 failed。
func (j *TopicLiteratureDownloadJob) Run(ctx context.Context) {
	rows, err := j.steps.ListByStatusAndStage(ctx, "running", "retrieve", scanLimit)
	if err != nil {
		notification.SendErrorLog(ctx, "topic-literature-download-list-failed", zap.Error(err))
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
		notification.SendErrorLog(ctx, "topic-literature-download-skip-no-result",
			zap.Uint64("step_id", step.ID),
			zap.Int("run_version", step.RunVersion))
		return ""
	}

	items := papersvc.ParseLiteratureDownloads(step.Result)
	if items == nil {
		items = []papersvc.LiteratureDownloadItem{}
	}
	hitCount := hitCountFromResult(step.Result)
	if hitCount == 0 {
		j.markFailed(ctx, step, items, "未检索到文献")
		return "failed"
	}
	if hitCount > 0 && len(items) == 0 {
		j.markFailed(ctx, step, items, "缺少 literature_downloads，无法下载 PDF")
		return "failed"
	}

	if !papersvc.LiteratureDownloadsNeedWork(items) {
		reason := fmt.Sprintf("检索到 %d 篇，但均无 PDF 下载链接（如 PubMed 元数据），未落盘", hitCount)
		j.markFailed(ctx, step, items, reason)
		return "failed"
	}

	notification.SendNotifyLog(ctx, "topic-literature-download-start",
		zap.Uint64("step_id", step.ID),
		zap.Uint64("manuscript_id", step.ManuscriptID),
		zap.Int("run_version", step.RunVersion),
		zap.Int("hit_count", hitCount))

	planned := papersvc.CountPlannedPDFDownloads(items)
	done := 0
	_ = j.saveRetrieveDownloadProgress(ctx, step, items, done, planned, planned > 0)

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
		} else if err := downloadPDFWithRetry(ctx, pdfURL, abs, it.ExternalKey); err != nil {
			it.DownloadError = err.Error()
		} else {
			it.DownloadError = ""
		}
		done++
		_ = j.saveRetrieveDownloadProgress(ctx, step, items, done, planned, false)
	}

	_ = j.mergeDownloadsResult(ctx, step, items)
	j.syncUserLiterature(ctx, step, items)

	if papersvc.LiteratureRetrieveCorpusReady(items) {
		return j.finishSuccess(ctx, step, items)
	}
	reason := papersvc.LiteratureDownloadFailureReason(items)
	if reason == "" {
		reason = "PDF 未落盘"
	}
	j.syncUserLiterature(ctx, step, items)
	j.markFailed(ctx, step, items, reason)
	return "failed"
}

func downloadPDFWithRetry(ctx context.Context, pdfURL, absPath, externalKey string) error {
	var lastErr error
	for attempt := 1; attempt <= pdfDownloadMaxAttempts; attempt++ {
		if st, err := os.Stat(absPath); err == nil && st.Size() > 0 {
			return nil
		}
		lastErr = papersvc.DownloadLiteraturePDF(ctx, pdfURL, absPath)
		if lastErr == nil {
			return nil
		}
		notification.SendErrorLog(ctx, "topic-literature-download-pdf-failed",
			zap.String("external_key", externalKey),
			zap.String("pdf_url", pdfURL),
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", pdfDownloadMaxAttempts),
			zap.Error(lastErr))
		if attempt >= pdfDownloadMaxAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pdfDownloadRetryDelay):
		}
	}
	return lastErr
}

func (j *TopicLiteratureDownloadJob) mergeDownloadsResult(ctx context.Context, step *entity.PaperOutputTopicStep, items []papersvc.LiteratureDownloadItem) error {
	meta := papersvc.LoadStepExtra(step)
	patch := map[string]any{"literature_downloads": items}
	step.Result = mustJSON(papersvc.MergeStepExtra(meta, patch))
	papersvc.SyncStepFilesAppend(step, items)
	return j.steps.SaveStep(ctx, step)
}

func (j *TopicLiteratureDownloadJob) syncUserLiterature(ctx context.Context, step *entity.PaperOutputTopicStep, items []papersvc.LiteratureDownloadItem) {
	if j.userLit == nil || step == nil {
		return
	}
	result := step.Result
	for _, err := range papersvc.SyncDownloadedUserLiterature(ctx, j.userLit, step.UserID, result, items) {
		notification.SendErrorLog(ctx, "topic-literature-user-lit-persist-failed", zap.Error(err))
	}
}

func (j *TopicLiteratureDownloadJob) saveRetrieveDownloadProgress(
	ctx context.Context,
	step *entity.PaperOutputTopicStep,
	items []papersvc.LiteratureDownloadItem,
	done, planned int,
	beforeFirst bool,
) error {
	step.ExecProgressPct = papersvc.ExecProgressPct(done, planned, beforeFirst)
	meta := papersvc.LoadStepExtra(step)
	meta["exec_progress_planned_downloads"] = planned
	meta["exec_progress_completed_downloads"] = done
	step.Result = mustJSON(papersvc.MergeStepExtra(meta, map[string]any{
		"literature_downloads": items,
	}))
	papersvc.SyncStepFilesAppend(step, items)
	return j.steps.SaveStep(ctx, step)
}

func (j *TopicLiteratureDownloadJob) finishSuccess(ctx context.Context, step *entity.PaperOutputTopicStep, items []papersvc.LiteratureDownloadItem) string {
	j.syncUserLiterature(ctx, step, items)
	meta := papersvc.LoadStepExtra(step)
	meta = papersvc.MergeStepExtra(meta, map[string]any{"literature_downloads": items})
	step.Result = mustJSON(meta)
	papersvc.SyncStepFilesAppend(step, items)

	hitCount := hitCountFromMeta(meta)
	if hitCount == 0 {
		j.markFailed(ctx, step, items, "未检索到文献")
		return "failed"
	}
	step.Status = "completed"
	now := time.Now()
	step.CompletedAt = &now
	pdfN := papersvc.CountLiteraturePDFsOnDisk(items)
	if pdfN == 0 {
		j.markFailed(ctx, step, items, "PDF 未落盘")
		return "failed"
	}
	step.ExecProgressPct = 100
	if pdfN < hitCount {
		meta = papersvc.MergeStepExtra(meta, map[string]any{
			"pdf_partial": true,
			"pdf_on_disk": pdfN,
		})
		step.Result = mustJSON(meta)
	}
	if err := j.steps.SaveStep(ctx, step); err != nil {
		notification.SendErrorLog(ctx, "topic-literature-download-save-failed", zap.Uint64("step_id", step.ID), zap.Error(err))
		return ""
	}
	if j.runs != nil {
		staged, syncErr := j.runs.SyncRunLiteratureToInoutWorkspace(ctx, uint(step.UserID), step.RunVersion, j.userLit)
		if syncErr != nil {
			notification.SendErrorLog(ctx, "topic-literature-inout-sync-failed",
				zap.Uint64("step_id", step.ID),
				zap.Error(syncErr))
		} else {
			logger.InfoZ(ctx, "topic-literature-inout-sync",
				zap.Uint64("step_id", step.ID),
				zap.Int("staged_pdfs", staged),
				zap.String("paper_root_hint", fmt.Sprintf("user_%d/manuscript_%d/run_%d/paper/literature", step.UserID, step.ManuscriptID, step.RunVersion)))
		}
	}
	notification.SendNotifyLog(ctx, "topic-literature-download-completed",
		zap.Uint64("step_id", step.ID),
		zap.Int("pdf_on_disk", pdfN),
		zap.Int("hit_count", hitCount))
	if j.runs != nil {
		if err := j.runs.OnRetrieveCompleted(ctx, uint(step.UserID), step.RunVersion); err != nil {
			notification.SendErrorLog(ctx, "topic-retrieve-chain-idea-discovery", zap.Error(err))
		}
	}
	return "completed"
}

func (j *TopicLiteratureDownloadJob) markFailed(ctx context.Context, step *entity.PaperOutputTopicStep, items []papersvc.LiteratureDownloadItem, reason string) {
	if items != nil {
		papersvc.SyncStepFilesAppend(step, items)
	}
	papersvc.FailTopicStep(step, reason)
	meta := papersvc.LoadStepExtra(step)
	patch := map[string]any{}
	if items != nil {
		patch["literature_downloads"] = items
	}
	if len(patch) > 0 {
		step.Result = mustJSON(papersvc.MergeStepExtra(meta, patch))
	}
	if err := j.steps.SaveStep(ctx, step); err != nil {
		notification.SendErrorLog(ctx, "topic-literature-download-fail-save", zap.Uint64("step_id", step.ID), zap.Error(err))
	}
}

func hitCountFromMeta(meta map[string]any) int {
	if n, ok := meta["hit_count"].(float64); ok {
		return int(n)
	}
	return 0
}

func hitCountFromResult(raw []byte) int {
	if len(raw) == 0 {
		return 0
	}
	var meta map[string]any
	if json.Unmarshal(raw, &meta) != nil {
		return 0
	}
	return hitCountFromMeta(meta)
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
