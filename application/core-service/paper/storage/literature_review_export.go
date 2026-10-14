package storage

import "time"

const LiteratureReviewStorageSubdir = "literature-review"

// LiteratureReviewStorageRoot storage/literature-review
func LiteratureReviewStorageRoot() string {
	return storageSubdirRoot(LiteratureReviewStorageSubdir, "PAPER_LITERATURE_REVIEW_STORAGE")
}

// WriteLiteratureReviewExportFiles 写入 PDF 与 Word，返回相对 storage/ 的路径。
func WriteLiteratureReviewExportFiles(userID uint, title, body string, at time.Time) (pdfRel, docxRel string, err error) {
	return WriteTopicDocumentExportFiles(
		LiteratureReviewStorageSubdir,
		"PAPER_LITERATURE_REVIEW_STORAGE",
		userID,
		title,
		body,
		at,
	)
}
