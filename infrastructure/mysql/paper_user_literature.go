package mysql

import (
	"context"
	"strings"

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

func (r *PaperUserLiteratureImpl) ListActiveByUserPaged(ctx context.Context, userID uint64, offset, limit int) ([]entity.PaperUserLiterature, error) {
	var rows []entity.PaperUserLiterature
	q := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, "active").
		Order("created_at DESC, id DESC")
	if limit > 0 {
		q = q.Offset(offset).Limit(limit)
	}
	err := q.Find(&rows).Error
	return rows, err
}

func (r *PaperUserLiteratureImpl) CountActiveByUser(ctx context.Context, userID uint64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&entity.PaperUserLiterature{}).
		Where("user_id = ? AND status = ?", userID, "active").
		Count(&n).Error
	return n, err
}

func (r *PaperUserLiteratureImpl) FindActiveByUserStorageURL(ctx context.Context, userID uint64, storageURL string) (*entity.PaperUserLiterature, error) {
	storageURL = strings.TrimSpace(storageURL)
	if userID == 0 || storageURL == "" {
		return nil, nil
	}
	var row entity.PaperUserLiterature
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ? AND storage_url = ?", userID, "active", storageURL).
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
