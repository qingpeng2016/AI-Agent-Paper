package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

type PaperFigureManuscriptRelImpl struct {
	db *gorm.DB
}

func NewPaperFigureManuscriptRelImpl(db *gorm.DB) repository.PaperFigureManuscriptRelRepo {
	return &PaperFigureManuscriptRelImpl{db: db}
}

func (r *PaperFigureManuscriptRelImpl) ListActiveManuscriptBindingsByFigureIDs(
	ctx context.Context,
	userID uint,
	figureIDs []uint64,
) ([]repository.FigureManuscriptBinding, error) {
	if userID == 0 || len(figureIDs) == 0 {
		return nil, nil
	}
	var rows []repository.FigureManuscriptBinding
	err := r.db.WithContext(ctx).
		Table("paper_user_figure_manuscript_rel AS r").
		Select("r.figure_id, r.manuscript_id, m.title AS manuscript_title").
		Joins("INNER JOIN paper_user_figure f ON f.id = r.figure_id AND f.user_id = ? AND f.status = 'active'", userID).
		Joins("INNER JOIN paper_manuscript m ON m.id = r.manuscript_id AND m.user_id = ? AND m.status = 'active'", userID).
		Where("r.figure_id IN ?", figureIDs).
		Order("r.figure_id ASC, m.id DESC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperFigureManuscriptRelImpl) ListActiveFigureBindingsByManuscriptIDs(
	ctx context.Context,
	userID uint,
	manuscriptIDs []uint64,
) ([]repository.ManuscriptFigureBinding, error) {
	if userID == 0 || len(manuscriptIDs) == 0 {
		return nil, nil
	}
	var rows []repository.ManuscriptFigureBinding
	err := r.db.WithContext(ctx).
		Table("paper_user_figure_manuscript_rel AS r").
		Select("r.manuscript_id, r.figure_id, f.title AS figure_title").
		Joins("INNER JOIN paper_manuscript m ON m.id = r.manuscript_id AND m.user_id = ? AND m.status = 'active'", userID).
		Joins("INNER JOIN paper_user_figure f ON f.id = r.figure_id AND f.user_id = ? AND f.status = 'active'", userID).
		Where("r.manuscript_id IN ?", manuscriptIDs).
		Order("r.manuscript_id ASC, f.title ASC, f.id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperFigureManuscriptRelImpl) ListActiveFigureIDsByManuscript(
	ctx context.Context,
	userID uint,
	manuscriptID uint64,
) ([]uint64, error) {
	if userID == 0 || manuscriptID == 0 {
		return nil, nil
	}
	var ids []uint64
	err := r.db.WithContext(ctx).
		Table("paper_user_figure_manuscript_rel AS r").
		Select("r.figure_id").
		Joins("INNER JOIN paper_manuscript m ON m.id = r.manuscript_id AND m.user_id = ? AND m.status = 'active'", userID).
		Joins("INNER JOIN paper_user_figure f ON f.id = r.figure_id AND f.user_id = ? AND f.status = 'active'", userID).
		Where("r.manuscript_id = ?", manuscriptID).
		Order("r.figure_id ASC").
		Pluck("r.figure_id", &ids).Error
	return ids, err
}

func (r *PaperFigureManuscriptRelImpl) SyncManuscriptFigureMembership(
	ctx context.Context,
	userID uint,
	manuscriptID uint64,
	figureIDs []uint64,
) error {
	if userID == 0 || manuscriptID == 0 {
		return nil
	}
	desired := dedupeUint64s(figureIDs)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing []uint64
		if err := tx.Table("paper_user_figure_manuscript_rel AS r").
			Select("r.figure_id").
			Joins("INNER JOIN paper_manuscript m ON m.id = r.manuscript_id AND m.user_id = ? AND m.status = 'active'", userID).
			Where("r.manuscript_id = ?", manuscriptID).
			Pluck("r.figure_id", &existing).Error; err != nil {
			return err
		}
		if len(desired) > 0 {
			var n int64
			if err := tx.Model(&entity.PaperFigure{}).
				Where("user_id = ? AND status = ? AND id IN ?", userID, "active", desired).
				Count(&n).Error; err != nil {
				return err
			}
			if n != int64(len(desired)) {
				return errorx.ErrParamsError.WithDetail("部分图表不存在或无权操作")
			}
		}
		toRemove := diffUint64s(existing, desired)
		if len(toRemove) > 0 {
			if err := tx.Where("manuscript_id = ? AND figure_id IN ?", manuscriptID, toRemove).
				Delete(&entity.PaperFigureManuscriptRel{}).Error; err != nil {
				return err
			}
		}
		toAdd := diffUint64s(desired, existing)
		for _, figID := range toAdd {
			rel := entity.PaperFigureManuscriptRel{
				FigureID:     figID,
				ManuscriptID: manuscriptID,
			}
			if err := tx.Create(&rel).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
