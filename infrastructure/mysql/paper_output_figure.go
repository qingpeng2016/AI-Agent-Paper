package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

type PaperOutputFigureImpl struct {
	db *gorm.DB
}

func NewPaperOutputFigureImpl(db *gorm.DB) repository.PaperOutputFigureRepo {
	return &PaperOutputFigureImpl{db: db}
}

func (r *PaperOutputFigureImpl) ListActiveByManuscript(ctx context.Context, manuscriptID uint64) ([]entity.PaperOutputFigure, error) {
	var rows []entity.PaperOutputFigure
	err := r.db.WithContext(ctx).
		Where("manuscript_id = ? AND status = ?", manuscriptID, "active").
		Order("created_at ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperOutputFigureImpl) Create(ctx context.Context, row *entity.PaperOutputFigure) error {
	if row == nil {
		return nil
	}
	return r.db.WithContext(ctx).Create(row).Error
}
