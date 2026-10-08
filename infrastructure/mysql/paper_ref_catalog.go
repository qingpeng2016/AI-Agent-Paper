package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

type PaperRefCatalogImpl struct {
	db *gorm.DB
}

func NewPaperRefCatalogImpl(db *gorm.DB) repository.PaperRefCatalogRepo {
	return &PaperRefCatalogImpl{db: db}
}

func (r *PaperRefCatalogImpl) ListActiveDisciplines(ctx context.Context) ([]entity.PaperRefDiscipline, error) {
	var rows []entity.PaperRefDiscipline
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort DESC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperRefCatalogImpl) ListExecutionIntensities(ctx context.Context) ([]entity.PaperRefExecutionIntensity, error) {
	var rows []entity.PaperRefExecutionIntensity
	err := r.db.WithContext(ctx).
		Order("FIELD(code, 'fast', 'balanced', 'deep'), code ASC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperRefCatalogImpl) ListAuditLevels(ctx context.Context) ([]entity.PaperRefAuditLevel, error) {
	var rows []entity.PaperRefAuditLevel
	err := r.db.WithContext(ctx).
		Order("FIELD(code, 'standard', 'polished', 'strict'), code ASC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperRefCatalogImpl) FindExecutionIntensityByCode(ctx context.Context, code string) (*entity.PaperRefExecutionIntensity, error) {
	var row entity.PaperRefExecutionIntensity
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperRefCatalogImpl) FindAuditLevelByCode(ctx context.Context, code string) (*entity.PaperRefAuditLevel, error) {
	var row entity.PaperRefAuditLevel
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
