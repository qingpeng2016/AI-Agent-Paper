package paper

import (
	"path/filepath"
)

// LiteratureStorageRoot 文献 PDF 存储目录（默认 storage/paper-files）。
func LiteratureStorageRoot() string {
	return PaperFilesStorageRoot()
}

// LiteratureLocalRelPath 库内记录的相对路径（相对项目 storage/，paper-files/xxx.pdf）。
func LiteratureLocalRelPath(externalKey string) string {
	return filepath.ToSlash(filepath.Join(paperFilesStorageSubdir, literatureLocalPDFFileName(externalKey)))
}

// LiteratureLocalAbsPath 由 relPath（paper-files/xxx.pdf）得到绝对路径。
func LiteratureLocalAbsPath(relPath string) string {
	return StorageLocalAbsPath(relPath)
}
