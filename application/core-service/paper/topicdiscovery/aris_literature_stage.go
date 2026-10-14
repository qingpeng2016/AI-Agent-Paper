package topicdiscovery

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/literature"
)

const arisLiteratureSubdir = "literature"

type arisLiteratureStagingEntry struct {
	ExternalKey     string `json:"external_key"`
	PlatformRelPath string `json:"platform_rel_path"`
	WorkspacePath   string `json:"workspace_path"`
	Bytes           int64  `json:"bytes,omitempty"`
}

// enrichCorpusWithRetrieveExtra 把 retrieve.extra 里的 literature_downloads 并入 corpus，供 invoke 落盘与 staging。
func enrichCorpusWithRetrieveExtra(corpus json.RawMessage, retrieveExtra []byte) json.RawMessage {
	downloads := literature.ParseLiteratureDownloads(retrieveExtra)
	if len(downloads) == 0 {
		return corpus
	}
	var m map[string]any
	if len(corpus) > 0 && json.Unmarshal(corpus, &m) != nil {
		m = nil
	}
	if m == nil {
		m = map[string]any{}
	}
	m["literature_downloads"] = downloads
	b, err := json.Marshal(m)
	if err != nil {
		return corpus
	}
	return json.RawMessage(b)
}

func parseLiteratureDownloadsFromCorpus(corpus json.RawMessage) []literature.LiteratureDownloadItem {
	if len(corpus) == 0 {
		return nil
	}
	var m map[string]any
	if json.Unmarshal(corpus, &m) != nil {
		return nil
	}
	raw, ok := m["literature_downloads"]
	if !ok {
		return nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var items []literature.LiteratureDownloadItem
	if json.Unmarshal(b, &items) != nil {
		return nil
	}
	return items
}

// stageRetrieveLiteraturePDFs 将平台 storage 中的 PDF 放入 ARIS workspace 的 literature/（research-lit 默认识别路径）。
func stageRetrieveLiteraturePDFs(paperRoot string, corpus json.RawMessage) ([]arisLiteratureStagingEntry, error) {
	items := parseLiteratureDownloadsFromCorpus(corpus)
	if len(items) == 0 {
		return nil, nil
	}
	litDir := filepath.Join(paperRoot, arisLiteratureSubdir)
	if err := os.MkdirAll(litDir, 0o755); err != nil {
		return nil, err
	}
	var staged []arisLiteratureStagingEntry
	for _, it := range items {
		rel := strings.TrimSpace(it.LocalPath)
		if rel == "" {
			continue
		}
		src := literature.LiteratureLocalAbsPath(rel)
		st, err := os.Stat(src)
		if err != nil || st.Size() == 0 {
			continue
		}
		name := filepath.Base(rel)
		if name == "" || name == "." {
			name = sanitizePathSegment(it.ExternalKey) + ".pdf"
		}
		dst := filepath.Join(litDir, name)
		if err := linkOrCopyPDF(src, dst); err != nil {
			return staged, fmt.Errorf("stage pdf %s: %w", it.ExternalKey, err)
		}
		wsRel := filepath.ToSlash(filepath.Join(arisLiteratureSubdir, name))
		staged = append(staged, arisLiteratureStagingEntry{
			ExternalKey:     strings.TrimSpace(it.ExternalKey),
			PlatformRelPath: filepath.ToSlash(rel),
			WorkspacePath:   wsRel,
			Bytes:           st.Size(),
		})
	}
	return staged, nil
}

func linkOrCopyPDF(src, dst string) error {
	if st, err := os.Stat(dst); err == nil && st.Size() > 0 {
		return nil
	}
	_ = os.Remove(dst)
	if err := os.Symlink(src, dst); err == nil {
		return nil
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(dst)
		return copyErr
	}
	return closeErr
}

func writeLiteratureStagingManifest(paperRoot string, staged []arisLiteratureStagingEntry) error {
	if len(staged) == 0 {
		return nil
	}
	dir := filepath.Join(paperRoot, ".paper")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	payload := map[string]any{
		"note":    "Platform retrieve PDFs staged for /research-lit local library (literature/*.pdf).",
		"entries": staged,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "literature_staging.json"), b, 0o644)
}
