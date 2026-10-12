package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

type PaperLiteraturePlatformImpl struct {
	db *gorm.DB
}

func NewPaperLiteraturePlatformImpl(db *gorm.DB) repository.PaperLiteraturePlatformRepo {
	return &PaperLiteraturePlatformImpl{db: db}
}

func (r *PaperLiteraturePlatformImpl) FindActiveByCode(ctx context.Context, code string) (*entity.PaperLiteraturePlatform, error) {
	var row entity.PaperLiteraturePlatform
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

func (r *PaperLiteraturePlatformImpl) ListActiveOrdered(ctx context.Context) ([]entity.PaperLiteraturePlatform, error) {
	var rows []entity.PaperLiteraturePlatform
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort ASC, id ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
