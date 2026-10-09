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

func (r *PaperOutputLiteratureReviewImpl) CreateAsCurrent(ctx context.Context, row *entity.PaperOutputLiteratureReview) error {
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
		row.IsCurrent = true
		if row.Status == "" {
			row.Status = "completed"
		}
		if row.Format == "" {
			row.Format = "md"
		}
		if err := tx.Model(&entity.PaperOutputLiteratureReview{}).
			Where("manuscript_id = ? AND is_current = ?", row.ManuscriptID, true).
			Updates(map[string]any{
				"is_current": false,
				"status":     "superseded",
			}).Error; err != nil {
			return err
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
		Where("manuscript_id = ? AND user_id = ?", manuscriptID, userID).
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
		Where("manuscript_id = ?", manuscriptID).
		Order("created_at DESC, id DESC").
		Find(&rows).Error
	return rows, err
}
