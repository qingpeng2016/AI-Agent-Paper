package paper

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"gorm.io/datatypes"
)

const (
	ExtraKeyLiteraturePlatformRequests = "literature_platform_requests"
)

var extraAppendOnlyKeys = map[string]struct{}{
	ExtraKeyLiteraturePlatformRequests: {},
}

// LiteraturePlatformRequestLog 文献平台一次检索的请求/响应（extra 数组，只追加）。
type LiteraturePlatformRequestLog struct {
	At         string `json:"at"`
	SourceCode string `json:"source_code"`
	Input      any    `json:"input"`
	Response   any    `json:"response,omitempty"`
	Error      string `json:"error,omitempty"`
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

const stepFileKindLiteraturePDF = "literature_pdf"

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

func LoadStepExtra(step *entity.PaperOutputTopicStep) map[string]any {
	if step == nil {
		return map[string]any{}
	}
	return unmarshalJSONMap(step.Extra)
}

func LoadStepFiles(step *entity.PaperOutputTopicStep) []StepFileEntry {
	if step == nil || len(step.Files) == 0 {
		return nil
	}
	var out []StepFileEntry
	_ = json.Unmarshal(step.Files, &out)
	return out
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

// MergeStepExtra 将 patch 合入 base；append-only 键只追加元素，其余键覆盖。
func MergeStepExtra(base, patch map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range patch {
		if _, appendOnly := extraAppendOnlyKeys[k]; appendOnly {
			switch items := v.(type) {
			case []any:
				for _, it := range items {
					appendExtraArray(out, k, it)
				}
			case []LiteraturePlatformRequestLog:
				for _, it := range items {
					appendExtraArray(out, k, it)
				}
			default:
				appendExtraArray(out, k, v)
			}
			continue
		}
		out[k] = v
	}
	return out
}

func AppendLiteraturePlatformRequest(meta map[string]any, entry LiteraturePlatformRequestLog) {
	if entry.At == "" {
		entry.At = time.Now().UTC().Format(time.RFC3339)
	}
	appendExtraArray(meta, ExtraKeyLiteraturePlatformRequests, entry)
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

func stepFilesFromDownloads(downloads []LiteratureDownloadItem) []StepFileEntry {
	var out []StepFileEntry
	for _, d := range downloads {
		path := strings.TrimSpace(d.LocalPath)
		if path == "" {
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

func SyncStepFilesAppend(step *entity.PaperOutputTopicStep, downloads []LiteratureDownloadItem) {
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
	return mustJSON(entries)
}
