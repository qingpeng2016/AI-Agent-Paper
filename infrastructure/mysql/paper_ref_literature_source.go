package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

type PaperRefLiteratureSourceImpl struct {
	db *gorm.DB
}

func NewPaperRefLiteratureSourceImpl(db *gorm.DB) repository.PaperRefLiteratureSourceRepo {
	return &PaperRefLiteratureSourceImpl{db: db}
}

func (r *PaperRefLiteratureSourceImpl) FindActiveByCode(ctx context.Context, code string) (*entity.PaperRefLiteratureSource, error) {
	var row entity.PaperRefLiteratureSource
	err := r.db.WithContext(ctx).
		Where("code = ? AND status = ?", code, "active").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperRefLiteratureSourceImpl) ListActiveOrdered(ctx context.Context) ([]entity.PaperRefLiteratureSource, error) {
	var rows []entity.PaperRefLiteratureSource
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("priority ASC, id ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
