package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const (
	PaperConfDirName             = "conf"
	PaperInoutDirName            = "inout"
	PaperFilesStorageSubdir      = "paper-files"
	paperFilesStorageSubdir      = PaperFilesStorageSubdir
	ExperimentDataStorageSubdir  = "experiment-data"
	experimentDataStorageSubdir  = ExperimentDataStorageSubdir
	submissionDataStorageSubdir  = "submission-data"
)

func ensureDirAbs(path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		return
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	_ = os.MkdirAll(abs, 0o755)
}

// PaperConfRoot 项目 conf 根目录（默认仓库下 conf/）。
func PaperConfRoot() string {
	if v := strings.TrimSpace(os.Getenv("PAPER_CONF_ROOT")); v != "" {
		return v
	}
	candidates := []string{PaperConfDirName, filepath.Join("AI-Agent-Paper", PaperConfDirName)}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, PaperConfDirName))
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	abs, _ := filepath.Abs(PaperConfDirName)
	return abs
}

// PaperInoutRoot 项目 inout 根目录（ARIS invoke 等工作区；默认仓库下 inout/）。
func PaperInoutRoot() string {
	var root string
	if v := strings.TrimSpace(os.Getenv("PAPER_ARIS_WORK_ROOT")); v != "" {
		root = v
	} else if v := strings.TrimSpace(os.Getenv("PAPER_INOUT_ROOT")); v != "" {
		root = v
	} else {
		candidates := []string{PaperInoutDirName, filepath.Join("AI-Agent-Paper", PaperInoutDirName)}
		if wd, err := os.Getwd(); err == nil {
			candidates = append(candidates, filepath.Join(wd, PaperInoutDirName))
		}
		for _, c := range candidates {
			if st, err := os.Stat(c); err == nil && st.IsDir() {
				abs, _ := filepath.Abs(c)
				root = abs
				break
			}
		}
		if root == "" {
			abs, _ := filepath.Abs(PaperInoutDirName)
			root = abs
		}
	}
	ensureDirAbs(root)
	abs, _ := filepath.Abs(root)
	if abs != "" {
		return abs
	}
	return root
}

// ArisInvokeWorkRoot 与 PaperInoutRoot 相同（invoke 会话目录直接挂在 inout 下）。
func ArisInvokeWorkRoot() string {
	return PaperInoutRoot()
}

// StorageRoot 项目本地 storage 根目录（默认 storage/，不存在则创建）。
func StorageRoot() string {
	root := "storage"
	if v := strings.TrimSpace(os.Getenv("PAPER_STORAGE_ROOT")); v != "" {
		root = v
	}
	ensureDirAbs(root)
	abs, _ := filepath.Abs(root)
	if abs != "" {
		return abs
	}
	return root
}

// StorageLocalAbsPath rel 为相对 storage/ 的路径（如 paper-files/x.pdf）。
func StorageLocalAbsPath(relPath string) string {
	relPath = filepath.ToSlash(strings.TrimPrefix(relPath, "/"))
	return filepath.Join(StorageRoot(), filepath.FromSlash(relPath))
}

// PaperFilesStorageRoot 论文相关文件（原文献 PDF 等）：storage/paper-files
func PaperFilesStorageRoot() string {
	var root string
	if v := strings.TrimSpace(os.Getenv("PAPER_LITERATURE_STORAGE")); v != "" {
		root = v
	} else {
		root = filepath.Join(StorageRoot(), paperFilesStorageSubdir)
	}
	ensureDirAbs(root)
	abs, _ := filepath.Abs(root)
	if abs != "" {
		return abs
	}
	return root
}

// ExperimentDataStorageRoot 实验数据：storage/experiment-data
func ExperimentDataStorageRoot() string {
	var root string
	if v := strings.TrimSpace(os.Getenv("PAPER_EXPERIMENT_DATA_STORAGE")); v != "" {
		root = v
	} else {
		root = filepath.Join(StorageRoot(), experimentDataStorageSubdir)
	}
	ensureDirAbs(root)
	abs, _ := filepath.Abs(root)
	if abs != "" {
		return abs
	}
	return root
}

// ExperimentDataLocalRelPath 相对 storage/ 的实验数据路径。
// SubmissionMainPDFLocalRelPath 相对 storage/ 的投稿主 PDF 路径。
func SubmissionMainPDFLocalRelPath(manuscriptID, submissionID uint64, originalFileName string) string {
	name := sanitizeStorageFileName(originalFileName)
	if !strings.HasSuffix(strings.ToLower(name), ".pdf") {
		name = name + ".pdf"
	}
	return filepath.ToSlash(filepath.Join(
		submissionDataStorageSubdir,
		fmt.Sprintf("manuscript_%d", manuscriptID),
		fmt.Sprintf("submission_%d", submissionID),
		name,
	))
}

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
