package literature

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

	"github.com/qingpeng2016/ai-agent-paper/common/notification"
	httpx "github.com/qingpeng2016/ai-agent-paper/infrastructure/http"
	"go.uber.org/zap"
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
	var statusCode int
	err = httpx.WithLiteratureOutbound(ctx, func() error {
		resp, doErr := client.Do(req)
		if doErr != nil {
			return doErr
		}
		defer resp.Body.Close()
		statusCode = resp.StatusCode
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("download http %d", resp.StatusCode)
		}
		tmp := absPath + ".part"
		f, createErr := os.Create(tmp)
		if createErr != nil {
			return createErr
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
	})
	if err != nil && statusCode != 0 && statusCode != http.StatusOK {
		notification.SendErrorLog(ctx, "literature-pdf-download-http",
			zap.String("pdf_url", pdfURL),
			zap.Int("http_status", statusCode))
	}
	return err
}

// LiteratureDownloadItem result.literature_downloads 单项。
type LiteratureDownloadItem struct {
	ExternalKey   string `json:"external_key"`
	PdfURL        string `json:"pdf_url"`
	LocalPath     string `json:"local_path"`
	DownloadError string `json:"download_error,omitempty"`
}

// ParseLiteratureDownloads 从 step.result JSON 解析 literature_downloads。
func ParseLiteratureDownloads(resultJSON []byte) []LiteratureDownloadItem {
	if len(resultJSON) == 0 {
		return nil
	}
	var wrap map[string]any
	if json.Unmarshal(resultJSON, &wrap) != nil {
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

// LiteratureDownloadFailureReason 汇总未就绪原因；空表示已全部落盘或无需下载。
func LiteratureDownloadFailureReason(items []LiteratureDownloadItem) string {
	var parts []string
	for _, it := range items {
		pdfURL := strings.TrimSpace(it.PdfURL)
		if pdfURL == "" {
			continue
		}
		rel := strings.TrimSpace(it.LocalPath)
		if rel == "" {
			parts = append(parts, it.ExternalKey+": 缺少 local_path")
			continue
		}
		abs := LiteratureLocalAbsPath(rel)
		st, err := os.Stat(abs)
		if err != nil || st.Size() == 0 {
			msg := strings.TrimSpace(it.DownloadError)
			if msg == "" {
				msg = "PDF 未落盘"
			}
			parts = append(parts, it.ExternalKey+": "+msg)
		}
	}
	return strings.Join(parts, "; ")
}

// LiteratureDownloadsNeedWork 是否存在需要 PDF 的条目。
func LiteratureDownloadsNeedWork(items []LiteratureDownloadItem) bool {
	for _, it := range items {
		if strings.TrimSpace(it.PdfURL) != "" {
			return true
		}
	}
	return false
}

// LiteratureAllDownloadsLackPDFURL 有下载项但全部无 pdf_url（如 PubMed/Crossref 元数据）。
func LiteratureAllDownloadsLackPDFURL(items []LiteratureDownloadItem) bool {
	if len(items) == 0 {
		return false
	}
	return !LiteratureDownloadsNeedWork(items)
}

// CountLiteraturePDFsOnDisk 已落盘且非空的 PDF 数量。
func CountLiteraturePDFsOnDisk(items []LiteratureDownloadItem) int {
	n := 0
	for _, it := range items {
		rel := strings.TrimSpace(it.LocalPath)
		if rel == "" {
			continue
		}
		st, err := os.Stat(LiteratureLocalAbsPath(rel))
		if err == nil && st.Size() > 0 {
			n++
		}
	}
	return n
}

// LiteratureRetrieveCorpusReady 至少有一个 PDF 已落盘即可进入后续流程（允许部分下载失败）。
func LiteratureRetrieveCorpusReady(items []LiteratureDownloadItem) bool {
	return CountLiteraturePDFsOnDisk(items) > 0
}

// LiteraturePDFDownloadAllFailed 凡有 pdf_url 的项均未落盘（或无任何 pdf_url）。
func LiteraturePDFDownloadAllFailed(items []LiteratureDownloadItem) bool {
	return CountLiteraturePDFsOnDisk(items) == 0
}

// FormatRetrievePDFSummary 检索步成功摘要文案（PDF 数量见 extra.pdf_on_disk）。
func FormatRetrievePDFSummary(hitCount, pdfOnDisk int) string {
	_ = pdfOnDisk
	return fmt.Sprintf("检索完成，共 %d 篇文献", hitCount)
}
