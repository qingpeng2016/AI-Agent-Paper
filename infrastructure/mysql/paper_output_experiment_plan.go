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
