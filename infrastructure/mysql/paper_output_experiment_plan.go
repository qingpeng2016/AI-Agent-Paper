package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

type PaperOutputExperimentPlanImpl struct {
	db *gorm.DB
}

func NewPaperOutputExperimentPlanImpl(db *gorm.DB) repository.PaperOutputExperimentPlanRepo {
	return &PaperOutputExperimentPlanImpl{db: db}
}

func (r *PaperOutputExperimentPlanImpl) Create(ctx context.Context, row *entity.PaperOutputExperimentPlan) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.PaperOutputExperimentPlan{}).
			Where("manuscript_id = ?", row.ManuscriptID).
			Update("is_current", false).Error; err != nil {
			return err
		}
		var maxVer *int
		if err := tx.Model(&entity.PaperOutputExperimentPlan{}).
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
		row.IsCurrent = true
		if row.Status == "" {
			row.Status = "completed"
		}
		if row.Format == "" {
			row.Format = "md"
		}
		return tx.Create(row).Error
	})
}

func (r *PaperOutputExperimentPlanImpl) experimentPlanListQuery(ctx context.Context, manuscriptID uint64) *gorm.DB {
	return r.db.WithContext(ctx).
		Model(&entity.PaperOutputExperimentPlan{}).
		Select(
			"id", "manuscript_id", "user_id", "version", "is_current", "status",
			"title", "summary", "format", "input_params", "experiment_data_uri",
			"created_at", "updated_at",
		).
		Where("manuscript_id = ? AND status <> ?", manuscriptID, "deleted").
		Order("created_at DESC, id DESC")
}

func (r *PaperOutputExperimentPlanImpl) ListByManuscript(
	ctx context.Context,
	manuscriptID uint64,
) ([]entity.PaperOutputExperimentPlan, error) {
	return r.ListByManuscriptPaged(ctx, manuscriptID, 0, 0)
}

func (r *PaperOutputExperimentPlanImpl) ListByManuscriptPaged(
	ctx context.Context,
	manuscriptID uint64,
	offset, limit int,
) ([]entity.PaperOutputExperimentPlan, error) {
	var rows []entity.PaperOutputExperimentPlan
	q := r.experimentPlanListQuery(ctx, manuscriptID)
	if limit > 0 {
		q = q.Offset(offset).Limit(limit)
	}
	err := q.Find(&rows).Error
	return rows, err
}

func (r *PaperOutputExperimentPlanImpl) CountByManuscript(ctx context.Context, manuscriptID uint64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&entity.PaperOutputExperimentPlan{}).
		Where("manuscript_id = ? AND status <> ?", manuscriptID, "deleted").
		Count(&n).Error
	return n, err
}

func (r *PaperOutputExperimentPlanImpl) GetByIDForUser(
	ctx context.Context,
	id, manuscriptID uint64,
	userID uint,
) (*entity.PaperOutputExperimentPlan, error) {
	var row entity.PaperOutputExperimentPlan
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

func (r *PaperOutputExperimentPlanImpl) GetByIDForManuscript(
	ctx context.Context,
	id, manuscriptID uint64,
) (*entity.PaperOutputExperimentPlan, error) {
	var row entity.PaperOutputExperimentPlan
	err := r.db.WithContext(ctx).
		Where("id = ? AND manuscript_id = ?", id, manuscriptID).
		First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *PaperOutputExperimentPlanImpl) UpdateExperimentDataURI(
	ctx context.Context,
	id, manuscriptID uint64,
	userID uint,
	uri string,
) (bool, error) {
	res := r.db.WithContext(ctx).
		Model(&entity.PaperOutputExperimentPlan{}).
		Where("id = ? AND manuscript_id = ? AND user_id = ?", id, manuscriptID, userID).
		Update("experiment_data_uri", uri)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *PaperOutputExperimentPlanImpl) ClearExperimentDataURI(
	ctx context.Context,
	id, manuscriptID uint64,
	userID uint,
) (bool, error) {
	res := r.db.WithContext(ctx).
		Model(&entity.PaperOutputExperimentPlan{}).
		Where("id = ? AND manuscript_id = ? AND user_id = ?", id, manuscriptID, userID).
		Update("experiment_data_uri", nil)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
