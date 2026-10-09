package paper

import (
	"os"
	"path/filepath"
	"strings"
)

const literatureStorageSubdir = "literature"

// LiteratureStorageRoot 文献 PDF 存储目录（默认 storage/literature）。
func LiteratureStorageRoot() string {
	if v := strings.TrimSpace(os.Getenv("PAPER_LITERATURE_STORAGE")); v != "" {
		return v
	}
	return filepath.Join("storage", literatureStorageSubdir)
}

// LiteratureLocalRelPath 库内记录的相对路径（相对项目 storage/）。
func LiteratureLocalRelPath(externalKey string) string {
	return filepath.Join(literatureStorageSubdir, literatureLocalPDFFileName(externalKey))
}

// LiteratureLocalAbsPath 由 relPath（literature/xxx.pdf）得到绝对路径。
func LiteratureLocalAbsPath(relPath string) string {
	return filepath.Join("storage", relPath)
}
