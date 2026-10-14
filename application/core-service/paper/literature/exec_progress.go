package literature

import (
	"os"
	"strings"
)

// ExecProgressPct 将 completed/planned 映射为 0–100（与 topic_discovery 多轮 LLM 步一致）。
func ExecProgressPct(completedUnits, plannedTotal int, beforeFirstInvoke bool) uint8 {
	if plannedTotal <= 0 {
		return 0
	}
	if beforeFirstInvoke && completedUnits == 0 {
		return 1
	}
	if completedUnits >= plannedTotal {
		return 100
	}
	pct := completedUnits * 100 / plannedTotal
	if pct < 1 && completedUnits > 0 {
		return 1
	}
	if pct > 99 && completedUnits < plannedTotal {
		return 99
	}
	return uint8(pct)
}

// CountPlannedPDFDownloads 需要尝试落盘的条目数（有 pdf_url）。
func CountPlannedPDFDownloads(items []LiteratureDownloadItem) int {
	n := 0
	for _, it := range items {
		if strings.TrimSpace(it.PdfURL) != "" {
			n++
		}
	}
	return n
}

// LiteraturePDFOnDiskForKey 平台 storage 中该 external_key 对应 PDF 是否已存在且非空。
func LiteraturePDFOnDiskForKey(externalKey string) bool {
	rel := LiteratureLocalRelPath(externalKey)
	st, err := os.Stat(LiteratureLocalAbsPath(rel))
	return err == nil && st.Size() > 0
}
