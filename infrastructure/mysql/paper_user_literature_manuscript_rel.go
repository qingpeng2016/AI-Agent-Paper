package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

type PaperUserLiteratureManuscriptRelImpl struct {
	db *gorm.DB
}

func NewPaperUserLiteratureManuscriptRelImpl(db *gorm.DB) repository.PaperUserLiteratureManuscriptRelRepo {
	return &PaperUserLiteratureManuscriptRelImpl{db: db}
}

func (r *PaperUserLiteratureManuscriptRelImpl) ListActiveManuscriptBindingsByLiteratureIDs(
	ctx context.Context,
	userID uint,
	literatureIDs []uint64,
) ([]repository.LiteratureManuscriptBinding, error) {
	if userID == 0 || len(literatureIDs) == 0 {
		return nil, nil
	}
	var rows []repository.LiteratureManuscriptBinding
	err := r.db.WithContext(ctx).
		Table("paper_user_literature_manuscript_rel AS r").
		Select("r.user_literature_id, r.manuscript_id, m.title AS manuscript_title").
		Joins("INNER JOIN paper_user_literature l ON l.id = r.user_literature_id AND l.user_id = ? AND l.status = 'active'", userID).
		Joins("INNER JOIN paper_manuscript m ON m.id = r.manuscript_id AND m.user_id = ? AND m.status = 'active'", userID).
		Where("r.user_literature_id IN ?", literatureIDs).
		Order("r.user_literature_id ASC, m.id DESC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperUserLiteratureManuscriptRelImpl) ListActiveLiteratureBindingsByManuscriptIDs(
	ctx context.Context,
	userID uint,
	manuscriptIDs []uint64,
) ([]repository.ManuscriptLiteratureBinding, error) {
	if userID == 0 || len(manuscriptIDs) == 0 {
		return nil, nil
	}
	var rows []repository.ManuscriptLiteratureBinding
	err := r.db.WithContext(ctx).
		Table("paper_user_literature_manuscript_rel AS r").
		Select("r.manuscript_id, r.user_literature_id, l.literature_title").
		Joins("INNER JOIN paper_manuscript m ON m.id = r.manuscript_id AND m.user_id = ? AND m.status = 'active'", userID).
		Joins("INNER JOIN paper_user_literature l ON l.id = r.user_literature_id AND l.user_id = ? AND l.status = 'active'", userID).
		Where("r.manuscript_id IN ?", manuscriptIDs).
		Order("r.manuscript_id ASC, l.literature_title ASC, l.id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperUserLiteratureManuscriptRelImpl) ListActiveLiteratureIDsByManuscript(
	ctx context.Context,
	userID uint,
	manuscriptID uint64,
) ([]uint64, error) {
	if userID == 0 || manuscriptID == 0 {
		return nil, nil
	}
	var ids []uint64
	err := r.db.WithContext(ctx).
		Table("paper_user_literature_manuscript_rel AS r").
		Select("r.user_literature_id").
		Joins("INNER JOIN paper_manuscript m ON m.id = r.manuscript_id AND m.user_id = ? AND m.status = 'active'", userID).
		Joins("INNER JOIN paper_user_literature l ON l.id = r.user_literature_id AND l.user_id = ? AND l.status = 'active'", userID).
		Where("r.manuscript_id = ?", manuscriptID).
		Order("r.user_literature_id ASC").
		Pluck("r.user_literature_id", &ids).Error
	return ids, err
}

func (r *PaperUserLiteratureManuscriptRelImpl) SyncManuscriptLiteratureMembership(
	ctx context.Context,
	userID uint,
	manuscriptID uint64,
	literatureIDs []uint64,
) error {
	if userID == 0 || manuscriptID == 0 {
		return nil
	}
	desired := dedupeUint64s(literatureIDs)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing []uint64
		if err := tx.Table("paper_user_literature_manuscript_rel AS r").
			Select("r.user_literature_id").
			Joins("INNER JOIN paper_manuscript m ON m.id = r.manuscript_id AND m.user_id = ? AND m.status = 'active'", userID).
			Where("r.manuscript_id = ?", manuscriptID).
			Pluck("r.user_literature_id", &existing).Error; err != nil {
			return err
		}
		if len(desired) > 0 {
			var n int64
			if err := tx.Model(&entity.PaperUserLiterature{}).
				Where("user_id = ? AND status = ? AND id IN ?", userID, "active", desired).
				Count(&n).Error; err != nil {
				return err
			}
			if n != int64(len(desired)) {
				return errorx.ErrParamsError.WithDetail("部分文献不存在或无权操作")
			}
		}
		toRemove := diffUint64s(existing, desired)
		if len(toRemove) > 0 {
			if err := tx.Where("manuscript_id = ? AND user_literature_id IN ?", manuscriptID, toRemove).
				Delete(&entity.PaperUserLiteratureManuscriptRel{}).Error; err != nil {
				return err
			}
		}
		toAdd := diffUint64s(desired, existing)
		for _, litID := range toAdd {
			rel := entity.PaperUserLiteratureManuscriptRel{
				UserLiteratureID: litID,
				ManuscriptID:     manuscriptID,
			}
			if err := tx.Create(&rel).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
