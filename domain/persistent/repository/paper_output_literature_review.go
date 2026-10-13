package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type PaperOutputLiteratureReviewRepo interface {
	Create(ctx context.Context, row *entity.PaperOutputLiteratureReview) error
	ListByManuscriptForUser(ctx context.Context, userID uint, manuscriptID uint64) ([]entity.PaperOutputLiteratureReview, error)
	ListByManuscriptForUserPaged(
		ctx context.Context,
		userID uint,
		manuscriptID uint64,
		offset, limit int,
	) ([]entity.PaperOutputLiteratureReview, error)
	CountByManuscriptForUser(ctx context.Context, userID uint, manuscriptID uint64) (int64, error)
	// ListByManuscript 按论文查全部版本（调用方须已校验 manuscript 归属）。
	ListByManuscript(ctx context.Context, manuscriptID uint64) ([]entity.PaperOutputLiteratureReview, error)
	UpdateStatusForManuscript(
		ctx context.Context,
		id, manuscriptID uint64,
		userID uint,
		status string,
	) (bool, error)
	GetByIDForUser(
		ctx context.Context,
		id, manuscriptID uint64,
		userID uint,
	) (*entity.PaperOutputLiteratureReview, error)
	ListByStatus(ctx context.Context, status string, limit int) ([]entity.PaperOutputLiteratureReview, error)
	CompleteExperimentPlanLink(
		ctx context.Context,
		reviewID, manuscriptID uint64,
		userID uint,
		planID uint64,
		llmResponseJSON []byte,
	) (bool, error)
	FailExperimentPlanGeneration(
		ctx context.Context,
		reviewID, manuscriptID uint64,
		userID uint,
		llmResponseJSON []byte,
	) (bool, error)
	SaveExperimentPlanLLMRequest(
		ctx context.Context,
		reviewID, manuscriptID uint64,
		userID uint,
		llmRequestJSON []byte,
	) (bool, error)
	TryBeginExperimentPlanGeneration(
		ctx context.Context,
		reviewID, manuscriptID uint64,
		userID uint,
	) (bool, error)
	// GetIDByExperimentPlanID 由 experiment_plan_id 反查文献综述（input_params 缺失时的兜底）。
	GetIDByExperimentPlanID(ctx context.Context, planID, manuscriptID uint64) (uint64, error)
	// MapLiteratureReviewIDsByExperimentPlanIDs 批量反查 experiment_plan_id → literature_review id。
	MapLiteratureReviewIDsByExperimentPlanIDs(
		ctx context.Context,
		manuscriptID uint64,
		planIDs []uint64,
	) (map[uint64]uint64, error)
	LinkExperimentPlan(ctx context.Context, reviewID, planID, userID, manuscriptID uint64) error
}
