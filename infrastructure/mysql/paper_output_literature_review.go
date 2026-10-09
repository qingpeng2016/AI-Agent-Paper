package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

type PaperOutputLiteratureReviewImpl struct {
	db *gorm.DB
}

func NewPaperOutputLiteratureReviewImpl(db *gorm.DB) repository.PaperOutputLiteratureReviewRepo {
	return &PaperOutputLiteratureReviewImpl{db: db}
}

func (r *PaperOutputLiteratureReviewImpl) Create(ctx context.Context, row *entity.PaperOutputLiteratureReview) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var maxVer *int
		if err := tx.Model(&entity.PaperOutputLiteratureReview{}).
			Where("manuscript_id = ?", row.ManuscriptID).
			Select("COALESCE(MAX(version), 0)").
			Scan(&maxVer).Error; err != nil {
			return err
		}
		next := 1
		if maxVer != nil {
			next = *maxVer + 1
		}
		row.Version = next
		if row.Status == "" {
			row.Status = "completed"
		}
		if row.Format == "" {
			row.Format = "md"
		}
		return tx.Create(row).Error
	})
}

func (r *PaperOutputLiteratureReviewImpl) ListByManuscriptForUser(
	ctx context.Context,
	userID uint,
	manuscriptID uint64,
) ([]entity.PaperOutputLiteratureReview, error) {
	var rows []entity.PaperOutputLiteratureReview
	err := r.db.WithContext(ctx).
		Where("manuscript_id = ? AND user_id = ? AND status <> ?", manuscriptID, userID, "deleted").
		Order("created_at DESC, id DESC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperOutputLiteratureReviewImpl) ListByManuscript(
	ctx context.Context,
	manuscriptID uint64,
) ([]entity.PaperOutputLiteratureReview, error) {
	var rows []entity.PaperOutputLiteratureReview
	err := r.db.WithContext(ctx).
		Where("manuscript_id = ? AND status <> ?", manuscriptID, "deleted").
		Order("created_at DESC, id DESC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperOutputLiteratureReviewImpl) UpdateStatusForManuscript(
	ctx context.Context,
	id, manuscriptID uint64,
	userID uint,
	status string,
) (bool, error) {
	res := r.db.WithContext(ctx).
		Model(&entity.PaperOutputLiteratureReview{}).
		Where("id = ? AND manuscript_id = ? AND user_id = ?", id, manuscriptID, userID).
		Update("status", status)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *PaperOutputLiteratureReviewImpl) TryBeginExperimentPlanGeneration(
	ctx context.Context,
	reviewID, manuscriptID uint64,
	userID uint,
) (bool, error) {
	res := r.db.WithContext(ctx).
		Model(&entity.PaperOutputLiteratureReview{}).
		Where(
			"id = ? AND manuscript_id = ? AND user_id = ? AND status IN ?",
			reviewID, manuscriptID, userID, []string{"completed", "gen_exp_fail"},
		).
		Updates(map[string]any{
			"status":                        "gen_exp_plan",
			"experiment_plan_llm_request":   nil,
			"experiment_plan_llm_response":  nil,
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *PaperOutputLiteratureReviewImpl) GetByIDForUser(
	ctx context.Context,
	id, manuscriptID uint64,
	userID uint,
) (*entity.PaperOutputLiteratureReview, error) {
	var row entity.PaperOutputLiteratureReview
	err := r.db.WithContext(ctx).
		Where("id = ? AND manuscript_id = ? AND user_id = ?", id, manuscriptID, userID).
		First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *PaperOutputLiteratureReviewImpl) ListByStatus(
	ctx context.Context,
	status string,
	limit int,
) ([]entity.PaperOutputLiteratureReview, error) {
	if limit <= 0 {
		limit = 20
	}
	var rows []entity.PaperOutputLiteratureReview
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("updated_at ASC, id ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *PaperOutputLiteratureReviewImpl) SaveExperimentPlanLLMRequest(
	ctx context.Context,
	reviewID, manuscriptID uint64,
	userID uint,
	llmRequestJSON []byte,
) (bool, error) {
	updates := map[string]any{}
	if len(llmRequestJSON) > 0 {
		updates["experiment_plan_llm_request"] = llmRequestJSON
	}
	if len(updates) == 0 {
		return false, nil
	}
	res := r.db.WithContext(ctx).
		Model(&entity.PaperOutputLiteratureReview{}).
		Where(
			"id = ? AND manuscript_id = ? AND user_id = ? AND status = ?",
			reviewID, manuscriptID, userID, "gen_exp_plan",
		).
		Updates(updates)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *PaperOutputLiteratureReviewImpl) CompleteExperimentPlanLink(
	ctx context.Context,
	reviewID, manuscriptID uint64,
	userID uint,
	planID uint64,
	metaJSON []byte,
	llmResponseJSON []byte,
) (bool, error) {
	updates := map[string]any{
		"status":                           "completed",
		"experiment_plan_id": planID,
	}
	if len(metaJSON) > 0 {
		updates["meta"] = metaJSON
	}
	if len(llmResponseJSON) > 0 {
		updates["experiment_plan_llm_response"] = llmResponseJSON
	}
	res := r.db.WithContext(ctx).
		Model(&entity.PaperOutputLiteratureReview{}).
		Where(
			"id = ? AND manuscript_id = ? AND user_id = ? AND status = ?",
			reviewID, manuscriptID, userID, "gen_exp_plan",
		).
		Updates(updates)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *PaperOutputLiteratureReviewImpl) FailExperimentPlanGeneration(
	ctx context.Context,
	reviewID, manuscriptID uint64,
	userID uint,
	metaJSON []byte,
	llmResponseJSON []byte,
) (bool, error) {
	updates := map[string]any{"status": "gen_exp_fail"}
	if len(metaJSON) > 0 {
		updates["meta"] = metaJSON
	}
	if len(llmResponseJSON) > 0 {
		updates["experiment_plan_llm_response"] = llmResponseJSON
	}
	res := r.db.WithContext(ctx).
		Model(&entity.PaperOutputLiteratureReview{}).
		Where(
			"id = ? AND manuscript_id = ? AND user_id = ? AND status = ?",
			reviewID, manuscriptID, userID, "gen_exp_plan",
		).
		Updates(updates)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *PaperOutputLiteratureReviewImpl) GetIDByExperimentPlanID(
	ctx context.Context,
	planID, manuscriptID uint64,
) (uint64, error) {
	var id uint64
	err := r.db.WithContext(ctx).
		Model(&entity.PaperOutputLiteratureReview{}).
		Select("id").
		Where(
			"experiment_plan_id = ? AND manuscript_id = ? AND status <> ?",
			planID, manuscriptID, "deleted",
		).
		Limit(1).
		Scan(&id).Error
	if err != nil {
		return 0, err
	}
	return id, nil
}
