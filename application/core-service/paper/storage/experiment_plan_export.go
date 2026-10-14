package storage

import "time"

const ExperimentPlanStorageSubdir = "experiment-plan"

// ExperimentPlanStorageRoot storage/experiment-plan
func ExperimentPlanStorageRoot() string {
	return storageSubdirRoot(ExperimentPlanStorageSubdir, "PAPER_EXPERIMENT_PLAN_STORAGE")
}

// WriteExperimentPlanExportFiles 写入 PDF 与 Word，返回相对 storage/ 的路径。
func WriteExperimentPlanExportFiles(userID uint, title, body string, at time.Time) (pdfRel, docxRel string, err error) {
	return WriteTopicDocumentExportFiles(
		ExperimentPlanStorageSubdir,
		"PAPER_EXPERIMENT_PLAN_STORAGE",
		userID,
		title,
		body,
		at,
	)
}
