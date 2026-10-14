package literature

import (
	"context"
	"errors"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/common/notification"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	restresponse "github.com/qingpeng2016/ai-agent-paper/domain/rest/response"
	httpx "github.com/qingpeng2016/ai-agent-paper/infrastructure/http"
	"go.uber.org/zap"
)

const ExtraKeyLiteraturePlatformRequests = "literature_platform_requests"

// LiteraturePlatformRequestLog 文献平台一次检索的请求/响应（extra 数组，只追加）。
type LiteraturePlatformRequestLog struct {
	At              string `json:"at"`
	SourceCode      string `json:"source_code"`
	Op              string `json:"op,omitempty"` // count | search
	Input           any    `json:"input"`
	RequestURL      string `json:"request_url,omitempty"`
	HTTPStatus      int    `json:"http_status,omitempty"`
	ResponsePreview string `json:"response_preview,omitempty"`
	Response        any    `json:"response,omitempty"`
	Error           string `json:"error,omitempty"`
}

func appendExtraArray(meta map[string]any, key string, item any) {
	if meta == nil {
		return
	}
	var arr []any
	switch cur := meta[key].(type) {
	case []any:
		arr = cur
	case nil:
		arr = nil
	default:
		arr = []any{cur}
	}
	meta[key] = append(arr, item)
}

// AppendLiteraturePlatformRequest 向 step extra 追加一条平台请求日志。
func AppendLiteraturePlatformRequest(meta map[string]any, entry LiteraturePlatformRequestLog) {
	if entry.At == "" {
		entry.At = time.Now().UTC().Format(time.RFC3339)
	}
	appendExtraArray(meta, ExtraKeyLiteraturePlatformRequests, entry)
}

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

// RetrieveLogScope 从选题步骤行构造 retrieve 日志 scope。
func RetrieveLogScope(topicStep *entity.PaperOutputTopicStep) LiteratureRetrieveLogScope {
	if topicStep == nil {
		return LiteratureRetrieveLogScope{}
	}
	return LiteratureRetrieveLogScope{
		StepID:       topicStep.ID,
		ManuscriptID: topicStep.ManuscriptID,
		UserID:       topicStep.UserID,
		RunVersion:   topicStep.RunVersion,
	}
}
