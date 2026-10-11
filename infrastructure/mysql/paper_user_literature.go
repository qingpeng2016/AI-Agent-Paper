package mysql

import (
	"context"

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

func (r *PaperUserLiteratureImpl) ListActiveByManuscript(ctx context.Context, manuscriptID uint64) ([]entity.PaperUserLiterature, error) {
	var rows []entity.PaperUserLiterature
	err := r.db.WithContext(ctx).
		Where("manuscript_id = ? AND status = ?", manuscriptID, "active").
		Order("created_at ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperUserLiteratureImpl) Create(ctx context.Context, row *entity.PaperUserLiterature) error {
	if row == nil {
		return nil
	}
	return r.db.WithContext(ctx).Create(row).Error
}
