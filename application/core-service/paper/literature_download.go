package paper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DownloadLiteraturePDF 下载 PDF 到绝对路径（已存在且非空则跳过）。
func DownloadLiteraturePDF(ctx context.Context, pdfURL, absPath string) error {
	pdfURL = strings.TrimSpace(pdfURL)
	if pdfURL == "" {
		return fmt.Errorf("empty pdf url")
	}
	if st, err := os.Stat(absPath); err == nil && st.Size() > 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pdfURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "AI-Agent-Paper/1.0")
	if key := strings.TrimSpace(os.Getenv("OPENALEX_API_KEY")); key != "" && strings.Contains(pdfURL, "content.openalex.org") {
		q := req.URL.Query()
		q.Set("api_key", key)
		req.URL.RawQuery = q.Encode()
	}
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download http %d", resp.StatusCode)
	}
	tmp := absPath + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(f, resp.Body)
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(tmp)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return closeErr
	}
	return os.Rename(tmp, absPath)
}

// LiteratureDownloadItem extra.literature_downloads 单项。
type LiteratureDownloadItem struct {
	ExternalKey string `json:"external_key"`
	PdfURL      string `json:"pdf_url"`
	LocalPath   string `json:"local_path"`
}

// ParseLiteratureDownloads 从 step.extra 解析 literature_downloads。
func ParseLiteratureDownloads(extra []byte) []LiteratureDownloadItem {
	if len(extra) == 0 {
		return nil
	}
	var wrap map[string]any
	if json.Unmarshal(extra, &wrap) != nil {
		return nil
	}
	raw, ok := wrap["literature_downloads"]
	if !ok {
		return nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var items []LiteratureDownloadItem
	if json.Unmarshal(b, &items) != nil {
		return nil
	}
	return items
}

// LiteratureDownloadsReady 是否全部可下载项已落盘（无 pdf_url 的项忽略）。
func LiteratureDownloadsReady(items []LiteratureDownloadItem) bool {
	if len(items) == 0 {
		return true
	}
	for _, it := range items {
		pdfURL := strings.TrimSpace(it.PdfURL)
		if pdfURL == "" {
			continue
		}
		rel := strings.TrimSpace(it.LocalPath)
		if rel == "" {
			return false
		}
		abs := LiteratureLocalAbsPath(rel)
		st, err := os.Stat(abs)
		if err != nil || st.Size() == 0 {
			return false
		}
	}
	return true
}
