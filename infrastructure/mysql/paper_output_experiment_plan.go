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

func (r *PaperOutputExperimentPlanImpl) ListByManuscript(
	ctx context.Context,
	manuscriptID uint64,
) ([]entity.PaperOutputExperimentPlan, error) {
	var rows []entity.PaperOutputExperimentPlan
	err := r.db.WithContext(ctx).
		Where("manuscript_id = ? AND status <> ?", manuscriptID, "deleted").
		Order("created_at DESC, id DESC").
		Find(&rows).Error
	return rows, err
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
