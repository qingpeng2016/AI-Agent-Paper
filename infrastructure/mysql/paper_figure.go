package mysql

import (
	"context"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

type PaperFigureImpl struct {
	db *gorm.DB
}

func NewPaperFigureImpl(db *gorm.DB) repository.PaperFigureRepo {
	return &PaperFigureImpl{db: db}
}

func (r *PaperFigureImpl) ListActiveByManuscript(ctx context.Context, manuscriptID uint64) ([]entity.PaperFigure, error) {
	var rows []entity.PaperFigure
	err := r.db.WithContext(ctx).
		Where("manuscript_id = ? AND status = ?", manuscriptID, "active").
		Order("created_at ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperFigureImpl) Create(ctx context.Context, row *entity.PaperFigure) error {
	if row == nil {
		return nil
	}
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *PaperFigureImpl) figureListBase(ctx context.Context, userID uint64, filter repository.FigureListFilter) *gorm.DB {
	q := r.db.WithContext(ctx).
		Table("paper_user_figure AS f").
		Where("f.user_id = ? AND f.status = ?", userID, "active")
	kw := strings.TrimSpace(filter.Keyword)
	if kw != "" {
		like := "%" + kw + "%"
		q = q.Where("(f.title LIKE ? OR f.file_name LIKE ? OR IFNULL(f.note, '') LIKE ?)", like, like, like)
	}
	if filter.CategoryID != nil {
		if *filter.CategoryID == 0 {
			q = q.Where(`NOT EXISTS (
				SELECT 1 FROM paper_user_figure_category_rel r
				INNER JOIN paper_user_figure_category c ON c.id = r.category_id AND c.status = 'active'
				WHERE r.figure_id = f.id
			)`)
		} else {
			q = q.Where(`EXISTS (
				SELECT 1 FROM paper_user_figure_category_rel r
				INNER JOIN paper_user_figure_category c ON c.id = r.category_id AND c.status = 'active'
				WHERE r.figure_id = f.id AND r.category_id = ?
			)`, *filter.CategoryID)
		}
	}
	if filter.ManuscriptID != nil {
		if *filter.ManuscriptID == 0 {
			q = q.Where(`NOT EXISTS (
				SELECT 1 FROM paper_user_figure_manuscript_rel r
				INNER JOIN paper_manuscript m ON m.id = r.manuscript_id AND m.user_id = ? AND m.status = 'active'
				WHERE r.figure_id = f.id
			)`, userID)
		} else {
			q = q.Where(`EXISTS (
				SELECT 1 FROM paper_user_figure_manuscript_rel r
				INNER JOIN paper_manuscript m ON m.id = r.manuscript_id AND m.user_id = ? AND m.status = 'active'
				WHERE r.figure_id = f.id AND r.manuscript_id = ?
			)`, userID, *filter.ManuscriptID)
		}
	}
	return q
}

func (r *PaperFigureImpl) ListActiveByUserPaged(
	ctx context.Context,
	userID uint64,
	filter repository.FigureListFilter,
	offset, limit int,
) ([]repository.FigureListRow, error) {
	var rows []repository.FigureListRow
	q := r.figureListBase(ctx, userID, filter).
		Select("f.*").
		Order("f.created_at DESC, f.id DESC")
	if limit > 0 {
		q = q.Offset(offset).Limit(limit)
	}
	err := q.Find(&rows).Error
	return rows, err
}

func (r *PaperFigureImpl) CountActiveByUser(ctx context.Context, userID uint64, filter repository.FigureListFilter) (int64, error) {
	var n int64
	err := r.figureListBase(ctx, userID, filter).
		Table("paper_user_figure AS f").
		Count(&n).Error
	return n, err
}

func (r *PaperFigureImpl) ListActiveCategoryBindingsByFigureIDs(
	ctx context.Context,
	userID uint64,
	figureIDs []uint64,
) ([]repository.FigureCategoryBinding, error) {
	if userID == 0 || len(figureIDs) == 0 {
		return nil, nil
	}
	var rows []repository.FigureCategoryBinding
	err := r.db.WithContext(ctx).
		Table("paper_user_figure_category_rel AS r").
		Select("r.figure_id, r.category_id, c.name AS category_name").
		Joins("INNER JOIN paper_user_figure f ON f.id = r.figure_id AND f.user_id = ? AND f.status = 'active'", userID).
		Joins("INNER JOIN paper_user_figure_category c ON c.id = r.category_id AND c.status = 'active'").
		Where("r.figure_id IN ?", figureIDs).
		Order("r.figure_id ASC, c.id DESC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperFigureImpl) ListActiveFigureIDsByCategory(
	ctx context.Context,
	userID, categoryID uint64,
) ([]uint64, error) {
	if userID == 0 || categoryID == 0 {
		return nil, nil
	}
	var ids []uint64
	err := r.db.WithContext(ctx).
		Table("paper_user_figure_category_rel AS r").
		Select("r.figure_id").
		Joins("INNER JOIN paper_user_figure f ON f.id = r.figure_id AND f.user_id = ? AND f.status = 'active'", userID).
		Where("r.category_id = ?", categoryID).
		Order("r.figure_id ASC").
		Pluck("r.figure_id", &ids).Error
	return ids, err
}

func (r *PaperFigureImpl) SyncCategoryFigureMembership(
	ctx context.Context,
	userID, categoryID uint64,
	figureIDs []uint64,
) error {
	if userID == 0 || categoryID == 0 {
		return nil
	}
	desired := dedupeUint64s(figureIDs)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing []uint64
		if err := tx.Table("paper_user_figure_category_rel AS r").
			Select("r.figure_id").
			Joins("INNER JOIN paper_user_figure f ON f.id = r.figure_id AND f.user_id = ? AND f.status = 'active'", userID).
			Where("r.category_id = ?", categoryID).
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
			if err := tx.Where("category_id = ? AND figure_id IN ?", categoryID, toRemove).
				Delete(&entity.PaperFigureCategoryRel{}).Error; err != nil {
				return err
			}
		}
		toAdd := diffUint64s(desired, existing)
		for _, figID := range toAdd {
			rel := entity.PaperFigureCategoryRel{
				FigureID:     figID,
				CategoryID:   categoryID,
			}
			if err := tx.Create(&rel).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
