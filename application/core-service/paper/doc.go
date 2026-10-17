// Package paper 聚合 core-service 下按领域拆分的子包，供 boot / handler / bot 沿用单一 import 路径。
package paper

import (
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/experimentplan"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/figure"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/internal/jsonx"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/literature"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/literaturereview"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/manuscript"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/submission"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/topicdiscovery"
)

type (
	LiteratureSearchService       = literature.LiteratureSearchService
	TopicDiscoveryOptionsService  = topicdiscovery.TopicDiscoveryOptionsService
	TopicDiscoveryRunService      = topicdiscovery.TopicDiscoveryRunService
	ManuscriptService             = manuscript.ManuscriptService
	LiteratureReviewService       = literaturereview.LiteratureReviewService
	ExperimentPlanService         = experimentplan.ExperimentPlanService
	PaperFigureService            = figure.PaperFigureService
	ManuscriptSubmissionService   = submission.ManuscriptSubmissionService
	LiteratureDownloadItem   = literature.LiteratureDownloadItem
	TopicStepExecNotify      = step.TopicStepExecNotify
)

var (
	NewLiteratureSearchService      = literature.NewLiteratureSearchService
	NewTopicDiscoveryOptionsService = topicdiscovery.NewTopicDiscoveryOptionsService
	NewTopicDiscoveryRunService     = topicdiscovery.NewTopicDiscoveryRunService
	NewManuscriptService            = manuscript.NewManuscriptService
	NewLiteratureReviewService      = literaturereview.NewLiteratureReviewService
	NewExperimentPlanService        = experimentplan.NewExperimentPlanService
	NewPaperFigureService           = figure.NewPaperFigureService
	NewManuscriptSubmissionService  = submission.NewManuscriptSubmissionService

	LoadStepExtra              = step.LoadStepExtra
	LoadStepResultMap          = step.LoadStepResultMap
	FailTopicStep              = step.FailTopicStep
	MergeStepExtra             = step.MergeStepExtra
	TopicStepHasPersistedOutput = step.TopicStepHasPersistedOutput
	SyncStepFilesAppend         = step.SyncStepFilesAppend
	NotifyTopicStepExecutionResult = step.NotifyTopicStepExecutionResult
	TopicStepOutcomeSuccess        = step.TopicStepOutcomeSuccess
	TopicStepOutcomeFailed         = step.TopicStepOutcomeFailed
	TopicStepOutcomeCheckpoint     = step.TopicStepOutcomeCheckpoint
	IdeaDiscoveryCheckpointDetail   = topicdiscovery.IdeaDiscoveryCheckpointDetail
	IdeaDiscoveryNeedsUserConfirm   = topicdiscovery.IdeaDiscoveryNeedsUserConfirm
	MustJSONStep     = jsonx.MustJSONStep

	ParseLiteratureDownloads           = literature.ParseLiteratureDownloads
	LiteratureDownloadsNeedWork       = literature.LiteratureDownloadsNeedWork
	LiteratureLocalRelPath             = literature.LiteratureLocalRelPath
	LiteratureLocalAbsPath             = literature.LiteratureLocalAbsPath
	LiteratureRetrieveCorpusReady      = literature.LiteratureRetrieveCorpusReady
	LiteratureDownloadFailureReason    = literature.LiteratureDownloadFailureReason
	DownloadLiteraturePDF              = literature.DownloadLiteraturePDF
	SyncDownloadedUserLiterature       = literature.SyncDownloadedUserLiterature
	CountLiteraturePDFsOnDisk          = literature.CountLiteraturePDFsOnDisk
	CountPlannedPDFDownloads           = literature.CountPlannedPDFDownloads
	ExecProgressPct                    = literature.ExecProgressPct
	LiteraturePDFOnDiskForKey          = literature.LiteraturePDFOnDiskForKey
	FormatRetrievePDFSummary           = literature.FormatRetrievePDFSummary
	RetrieveSearchPending              = topicdiscovery.RetrieveSearchPending
	ParseFigureListFilter              = figure.ParseFigureListFilter
)

const LitReviewStatusGeneratingExperimentPlan = experimentplan.LitReviewStatusGeneratingExperimentPlan
