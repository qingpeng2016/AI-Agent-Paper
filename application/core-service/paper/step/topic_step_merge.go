package step

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/internal/jsonx"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/literature"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"gorm.io/datatypes"
)

var resultAppendOnlyKeys = map[string]struct{}{
	literature.ExtraKeyLiteraturePlatformRequests: {},
}

var topicStepRuntimeResultKeys = map[string]struct{}{
	"async_llm":                  {},
	"async_llm_done":             {},
	"aris_awaiting_checkpoint":   {},
	"aris_pause_after_phase":     {},
	"aris_invoke_retries":        {},
	"aris_invoke_last_error":     {},
	"aris_invoke_count":          {},
	"last_invoke_result":         {},
	"native_run_id":              {},
	"exec_progress_planned_calls": {},
	"exec_progress_completed_calls": {},
	"exec_progress_planned_downloads": {},
	"exec_progress_completed_downloads": {},
	"awaiting_user_confirm":      {},
	"error":                      {},
}

// StepFileEntry 步骤关联文件（PDF 等），files 列 JSON 数组，只追加。
type StepFileEntry struct {
	Kind        string `json:"kind"`
	Path        string `json:"path"`
	ExternalKey string `json:"external_key,omitempty"`
	SourceCode  string `json:"source_code,omitempty"`
	At          string `json:"at"`
	Note        string `json:"note,omitempty"`
}

const StepFileKindLiteraturePDF = "literature_pdf"

const stepFileKindLiteraturePDF = StepFileKindLiteraturePDF

func unmarshalJSONMap(raw datatypes.JSON) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil || m == nil {
		return map[string]any{}
	}
	return m
}

// FailTopicStep 将步骤置为 failed 并写入 error_msg；清掉 result 内 bot 续跑标记。
func FailTopicStep(step *entity.PaperOutputTopicStep, reason string) {
	if step == nil {
		return
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "unknown error"
	}
	step.Status = "failed"
	step.ErrorMsg = &reason
	step.ExecProgressPct = 0
	meta := LoadStepResultMap(step)
	clearTopicStepRuntimeKeys(meta)
	step.Result = jsonx.MustJSON(meta)
	now := time.Now()
	if step.CompletedAt == nil {
		step.CompletedAt = &now
	}
}

// LoadStepResultMap 解析 step.result JSON 对象。
func LoadStepResultMap(step *entity.PaperOutputTopicStep) map[string]any {
	if step == nil {
		return map[string]any{}
	}
	return unmarshalJSONMap(step.Result)
}

// LoadStepExtra 兼容旧名：数据均在 result 列。
func LoadStepExtra(step *entity.PaperOutputTopicStep) map[string]any {
	return LoadStepResultMap(step)
}

// MergeStepResultMap 将 patch 合入 base；平台请求日志等 append-only 键只追加。
func MergeStepResultMap(base, patch map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range patch {
		if _, appendOnly := resultAppendOnlyKeys[k]; appendOnly {
			switch items := v.(type) {
			case []any:
				for _, it := range items {
					appendResultArray(out, k, it)
				}
			case []literature.LiteraturePlatformRequestLog:
				for _, it := range items {
					appendResultArray(out, k, it)
				}
			default:
				appendResultArray(out, k, v)
			}
			continue
		}
		out[k] = v
	}
	return out
}

// MergeStepExtra 兼容旧名。
func MergeStepExtra(base, patch map[string]any) map[string]any {
	return MergeStepResultMap(base, patch)
}

// TopicStepHasPersistedOutput 是否已有本步业务产出（区别于仅含 async_llm 等运行态 key）。
func TopicStepHasPersistedOutput(step *entity.PaperOutputTopicStep) bool {
	if step == nil || len(step.Result) == 0 {
		return false
	}
	m := LoadStepResultMap(step)
	switch strings.TrimSpace(step.StageCode) {
	case "retrieve":
		_, ok := m["literature_hits"]
		return ok
	case "idea_discovery", "topic_discovery":
		if _, ok := m["pipeline"]; ok {
			return true
		}
		if _, ok := m["passed"]; ok {
			return true
		}
		return false
	case "topic_confirm":
		_, ok := m["selected_round"]
		return ok
	default:
		return !isRuntimeOnlyResultMap(m)
	}
}

func isRuntimeOnlyResultMap(m map[string]any) bool {
	if len(m) == 0 {
		return false
	}
	for k := range m {
		if _, rt := topicStepRuntimeResultKeys[k]; !rt {
			return false
		}
	}
	return true
}

func clearTopicStepRuntimeKeys(m map[string]any) {
	for k := range topicStepRuntimeResultKeys {
		delete(m, k)
	}
}

func appendResultArray(meta map[string]any, key string, item any) {
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

func LoadStepFiles(step *entity.PaperOutputTopicStep) []StepFileEntry {
	if step == nil || len(step.Files) == 0 {
		return nil
	}
	var out []StepFileEntry
	_ = json.Unmarshal(step.Files, &out)
	return out
}

func AppendStepFileEntries(existing []StepFileEntry, adds ...StepFileEntry) []StepFileEntry {
	if len(adds) == 0 {
		return existing
	}
	out := make([]StepFileEntry, 0, len(existing)+len(adds))
	out = append(out, existing...)
	for _, a := range adds {
		if a.At == "" {
			a.At = time.Now().UTC().Format(time.RFC3339)
		}
		out = append(out, a)
	}
	return out
}

func stepFilesFromDownloads(downloads []literature.LiteratureDownloadItem) []StepFileEntry {
	var out []StepFileEntry
	for _, d := range downloads {
		path := strings.TrimSpace(d.LocalPath)
		if path == "" {
			continue
		}
		st, err := os.Stat(literature.LiteratureLocalAbsPath(path))
		if err != nil || st.Size() == 0 {
			continue
		}
		out = append(out, StepFileEntry{
			Kind:        stepFileKindLiteraturePDF,
			Path:        path,
			ExternalKey: d.ExternalKey,
			At:          time.Now().UTC().Format(time.RFC3339),
		})
	}
	return out
}

func SyncStepFilesAppend(step *entity.PaperOutputTopicStep, downloads []literature.LiteratureDownloadItem) {
	adds := stepFilesFromDownloads(downloads)
	if len(adds) == 0 {
		return
	}
	existing := LoadStepFiles(step)
	seen := map[string]struct{}{}
	for _, e := range existing {
		seen[e.Path+"|"+e.ExternalKey] = struct{}{}
	}
	var newOnes []StepFileEntry
	for _, a := range adds {
		key := a.Path + "|" + a.ExternalKey
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		newOnes = append(newOnes, a)
	}
	if len(newOnes) == 0 {
		return
	}
	step.Files = marshalStepFiles(AppendStepFileEntries(existing, newOnes...))
}

func marshalStepFiles(entries []StepFileEntry) datatypes.JSON {
	if len(entries) == 0 {
		return nil
	}
	return jsonx.MustJSON(entries)
}
