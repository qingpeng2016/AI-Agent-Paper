package mysql

import (
	"context"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/common/errorx"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

type PaperUserLiteratureImpl struct {
	db *gorm.DB
}

func NewPaperUserLiteratureImpl(db *gorm.DB) repository.PaperUserLiteratureRepo {
	return &PaperUserLiteratureImpl{db: db}
}

func (r *PaperUserLiteratureImpl) ListActiveByUser(ctx context.Context, userID uint64) ([]entity.PaperUserLiterature, error) {
	var rows []entity.PaperUserLiterature
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, "active").
		Order("created_at ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperUserLiteratureImpl) userLiteratureListBase(ctx context.Context, userID uint64, filter repository.UserLiteratureListFilter) *gorm.DB {
	q := r.db.WithContext(ctx).
		Table("paper_user_literature AS l").
		Where("l.user_id = ? AND l.status = ?", userID, "active")
	kw := strings.TrimSpace(filter.Keyword)
	if kw != "" {
	 like := "%" + kw + "%"
		q = q.Where("(l.literature_title LIKE ? OR l.file_name LIKE ? OR IFNULL(l.note, '') LIKE ?)", like, like, like)
	}
	if filter.CategoryID != nil {
		if *filter.CategoryID == 0 {
			q = q.Where(`NOT EXISTS (
				SELECT 1 FROM paper_user_literature_category_rel r
				INNER JOIN paper_user_literature_category c ON c.id = r.category_id AND c.status = 'active'
				WHERE r.user_literature_id = l.id
			)`)
		} else {
			q = q.Where(`EXISTS (
				SELECT 1 FROM paper_user_literature_category_rel r
				INNER JOIN paper_user_literature_category c ON c.id = r.category_id AND c.status = 'active'
				WHERE r.user_literature_id = l.id AND r.category_id = ?
			)`, *filter.CategoryID)
		}
	}
	if filter.ManuscriptID != nil {
		if *filter.ManuscriptID == 0 {
			q = q.Where(`NOT EXISTS (
				SELECT 1 FROM paper_user_literature_manuscript_rel r
				INNER JOIN paper_manuscript m ON m.id = r.manuscript_id AND m.user_id = ? AND m.status = 'active'
				WHERE r.user_literature_id = l.id
			)`, userID)
		} else {
			q = q.Where(`EXISTS (
				SELECT 1 FROM paper_user_literature_manuscript_rel r
				INNER JOIN paper_manuscript m ON m.id = r.manuscript_id AND m.user_id = ? AND m.status = 'active'
				WHERE r.user_literature_id = l.id AND r.manuscript_id = ?
			)`, userID, *filter.ManuscriptID)
		}
	}
	return q
}

func (r *PaperUserLiteratureImpl) ListActiveByUserPaged(
	ctx context.Context,
	userID uint64,
	filter repository.UserLiteratureListFilter,
	offset, limit int,
) ([]repository.UserLiteratureListRow, error) {
	var rows []repository.UserLiteratureListRow
	q := r.userLiteratureListBase(ctx, userID, filter).
		Select("l.*").
		Order("l.created_at DESC, l.id DESC")
	if limit > 0 {
		q = q.Offset(offset).Limit(limit)
	}
	err := q.Find(&rows).Error
	return rows, err
}

func (r *PaperUserLiteratureImpl) CountActiveByUser(ctx context.Context, userID uint64, filter repository.UserLiteratureListFilter) (int64, error) {
	var n int64
	err := r.userLiteratureListBase(ctx, userID, filter).
		Table("paper_user_literature AS l").
		Count(&n).Error
	return n, err
}

func (r *PaperUserLiteratureImpl) FindActiveByUserStorageURI(ctx context.Context, userID uint64, storageURI string) (*entity.PaperUserLiterature, error) {
	storageURI = strings.TrimSpace(storageURI)
	if userID == 0 || storageURI == "" {
		return nil, nil
	}
	var row entity.PaperUserLiterature
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ? AND storage_uri = ?", userID, "active", storageURI).
		First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperUserLiteratureImpl) Create(ctx context.Context, row *entity.PaperUserLiterature) error {
	if row == nil {
		return nil
	}
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *PaperUserLiteratureImpl) ListActiveCategoryBindingsByLiteratureIDs(
	ctx context.Context,
	userID uint64,
	literatureIDs []uint64,
) ([]repository.UserLiteratureCategoryBinding, error) {
	if userID == 0 || len(literatureIDs) == 0 {
		return nil, nil
	}
	var rows []repository.UserLiteratureCategoryBinding
	err := r.db.WithContext(ctx).
		Table("paper_user_literature_category_rel AS r").
		Select("r.user_literature_id, r.category_id, c.name AS category_name").
		Joins("INNER JOIN paper_user_literature l ON l.id = r.user_literature_id AND l.user_id = ? AND l.status = 'active'", userID).
		Joins("INNER JOIN paper_user_literature_category c ON c.id = r.category_id AND c.status = 'active'").
		Where("r.user_literature_id IN ?", literatureIDs).
		Order("r.user_literature_id ASC, c.id DESC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperUserLiteratureImpl) ListActiveLiteratureIDsByCategory(
	ctx context.Context,
	userID, categoryID uint64,
) ([]uint64, error) {
	if userID == 0 || categoryID == 0 {
		return nil, nil
	}
	var ids []uint64
	err := r.db.WithContext(ctx).
		Table("paper_user_literature_category_rel AS r").
		Select("r.user_literature_id").
		Joins("INNER JOIN paper_user_literature l ON l.id = r.user_literature_id AND l.user_id = ? AND l.status = 'active'", userID).
		Where("r.category_id = ?", categoryID).
		Order("r.user_literature_id ASC").
		Pluck("r.user_literature_id", &ids).Error
	return ids, err
}

func (r *PaperUserLiteratureImpl) SyncCategoryLiteratureMembership(
	ctx context.Context,
	userID, categoryID uint64,
	literatureIDs []uint64,
) error {
	if userID == 0 || categoryID == 0 {
		return nil
	}
	desired := dedupeUint64s(literatureIDs)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing []uint64
		if err := tx.Table("paper_user_literature_category_rel AS r").
			Select("r.user_literature_id").
			Joins("INNER JOIN paper_user_literature l ON l.id = r.user_literature_id AND l.user_id = ? AND l.status = 'active'", userID).
			Where("r.category_id = ?", categoryID).
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
			if err := tx.Where("category_id = ? AND user_literature_id IN ?", categoryID, toRemove).
				Delete(&entity.PaperUserLiteratureCategoryRel{}).Error; err != nil {
				return err
			}
		}
		toAdd := diffUint64s(desired, existing)
		for _, litID := range toAdd {
			rel := entity.PaperUserLiteratureCategoryRel{
				UserLiteratureID: litID,
				CategoryID:       categoryID,
			}
			if err := tx.Create(&rel).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func dedupeUint64s(in []uint64) []uint64 {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[uint64]struct{}, len(in))
	out := make([]uint64, 0, len(in))
	for _, id := range in {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func diffUint64s(from, subtract []uint64) []uint64 {
	sub := make(map[uint64]struct{}, len(subtract))
	for _, id := range subtract {
		sub[id] = struct{}{}
	}
	var out []uint64
	for _, id := range from {
		if _, ok := sub[id]; !ok {
			out = append(out, id)
		}
	}
	return out
}
