package paper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const (
	paperFilesStorageSubdir     = "paper-files"
	experimentDataStorageSubdir = "experiment-data"
)

// StorageRoot 项目本地 storage 根目录（相对工作目录）。
func StorageRoot() string {
	if v := strings.TrimSpace(os.Getenv("PAPER_STORAGE_ROOT")); v != "" {
		return v
	}
	return "storage"
}

// StorageLocalAbsPath rel 为相对 storage/ 的路径（如 paper-files/x.pdf）。
func StorageLocalAbsPath(relPath string) string {
	relPath = filepath.ToSlash(strings.TrimPrefix(relPath, "/"))
	return filepath.Join(StorageRoot(), filepath.FromSlash(relPath))
}

// PaperFilesStorageRoot 论文相关文件（原文献 PDF 等）：storage/paper-files
func PaperFilesStorageRoot() string {
	if v := strings.TrimSpace(os.Getenv("PAPER_LITERATURE_STORAGE")); v != "" {
		return v
	}
	return filepath.Join(StorageRoot(), paperFilesStorageSubdir)
}

// ExperimentDataStorageRoot 实验数据：storage/experiment-data
func ExperimentDataStorageRoot() string {
	if v := strings.TrimSpace(os.Getenv("PAPER_EXPERIMENT_DATA_STORAGE")); v != "" {
		return v
	}
	return filepath.Join(StorageRoot(), experimentDataStorageSubdir)
}

// ExperimentDataLocalRelPath 相对 storage/ 的实验数据路径。
func ExperimentDataLocalRelPath(manuscriptID, planID uint64, originalFileName string) string {
	name := sanitizeStorageFileName(originalFileName)
	return filepath.ToSlash(filepath.Join(
		experimentDataStorageSubdir,
		fmt.Sprintf("manuscript_%d", manuscriptID),
		fmt.Sprintf("plan_%d", planID),
		name,
	))
}

func sanitizeStorageFileName(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "" || base == "." || base == ".." {
		return "upload.bin"
	}
	var b strings.Builder
	for _, r := range base {
		if r == '/' || r == '\\' || unicode.IsControl(r) {
			continue
		}
		b.WriteRune(r)
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "upload.bin"
	}
	return out
}
