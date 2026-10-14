package paper

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/common/notification"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	restresponse "github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
	httpx "github.com/qingpeng2016/ai-agent-paper/infrastructure/http"
	"go.uber.org/zap"
)

// LiteratureRetrieveLogScope 选题 retrieve 关联字段，便于在 info.log 里检索。
type LiteratureRetrieveLogScope struct {
	StepID       uint64
	ManuscriptID uint64
	UserID       uint64
	RunVersion   int
}

func enrichLiteraturePlatformRequestLog(entry *LiteraturePlatformRequestLog, err error) {
	if entry == nil || err == nil {
		return
	}
	var he *httpx.LiteratureHTTPError
	if !errors.As(err, &he) || he == nil {
		return
	}
	if entry.Op == "" {
		entry.Op = he.Operation
	}
	if entry.RequestURL == "" {
		entry.RequestURL = he.RequestURL
	}
	if entry.HTTPStatus == 0 {
		entry.HTTPStatus = he.HTTPStatus
	}
	if entry.ResponsePreview == "" {
		entry.ResponsePreview = he.ResponsePreview
	}
}

func literaturePlatformHitCount(respAny any) int {
	if respAny == nil {
		return 0
	}
	if r, ok := respAny.(*restresponse.PaperLiteratureSearchResult); ok && r != nil {
		return len(r.Hits)
	}
	return 0
}

func logLiteraturePlatformRequest(ctx context.Context, scope LiteratureRetrieveLogScope, entry LiteraturePlatformRequestLog) {
	fields := []zap.Field{
		zap.Uint64("step_id", scope.StepID),
		zap.Uint64("manuscript_id", scope.ManuscriptID),
		zap.Uint64("user_id", scope.UserID),
		zap.Int("run_version", scope.RunVersion),
		zap.String("source_code", entry.SourceCode),
		zap.String("op", entry.Op),
	}
	if entry.RequestURL != "" {
		fields = append(fields, zap.String("request_url", entry.RequestURL))
	}
	if entry.HTTPStatus > 0 {
		fields = append(fields, zap.Int("http_status", entry.HTTPStatus))
	}
	if entry.ResponsePreview != "" {
		fields = append(fields, zap.String("response_preview", entry.ResponsePreview))
	}
	if entry.Error != "" {
		fields = append(fields, zap.String("error", entry.Error))
		notification.SendErrorLog(ctx, "literature-platform-request", fields...)
		return
	}
	fields = append(fields, zap.Int("hit_count", literaturePlatformHitCount(entry.Response)))
	logger.InfoZ(ctx, "literature-platform-request", fields...)
}

// RecordLiteraturePlatformRequest 写入 extra 并打 literature-platform-request 日志（含 HTTP 摘要）。
func RecordLiteraturePlatformRequest(
	ctx context.Context,
	meta map[string]any,
	scope LiteratureRetrieveLogScope,
	entry LiteraturePlatformRequestLog,
	err error,
) {
	if err != nil && entry.Error == "" {
		entry.Error = err.Error()
	}
	enrichLiteraturePlatformRequestLog(&entry, err)
	logLiteraturePlatformRequest(ctx, scope, entry)
	AppendLiteraturePlatformRequest(meta, entry)
}

func retrieveLogScope(step *entity.PaperOutputTopicStep) LiteratureRetrieveLogScope {
	if step == nil {
		return LiteratureRetrieveLogScope{}
	}
	return LiteratureRetrieveLogScope{
		StepID:       step.ID,
		ManuscriptID: step.ManuscriptID,
		UserID:       step.UserID,
		RunVersion:   step.RunVersion,
	}
}
