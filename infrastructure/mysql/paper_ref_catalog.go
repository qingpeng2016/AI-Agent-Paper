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

func (r *PaperRefCatalogImpl) FindDisciplineByCode(ctx context.Context, code string) (*entity.PaperRefDiscipline, error) {
	var row entity.PaperRefDiscipline
	err := r.db.WithContext(ctx).Where("code = ? AND status = ?", code, "active").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperRefCatalogImpl) FindDisciplineByID(ctx context.Context, id uint64) (*entity.PaperRefDiscipline, error) {
	if id == 0 {
		return nil, nil
	}
	var row entity.PaperRefDiscipline
	err := r.db.WithContext(ctx).Where("id = ? AND status = ?", id, "active").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperRefCatalogImpl) ListActiveDisciplines(ctx context.Context) ([]entity.PaperRefDiscipline, error) {
	var rows []entity.PaperRefDiscipline
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort DESC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperRefCatalogImpl) ListExecutionIntensities(ctx context.Context) ([]entity.PaperLLMConstraintExecIntensity, error) {
	var rows []entity.PaperLLMConstraintExecIntensity
	err := r.db.WithContext(ctx).
		Order("FIELD(code, 'fast', 'balanced', 'deep'), code ASC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperRefCatalogImpl) ListAuditLevels(ctx context.Context) ([]entity.PaperLLMConstraintGate, error) {
	var rows []entity.PaperLLMConstraintGate
	err := r.db.WithContext(ctx).
		Order("FIELD(code, 'standard', 'polished', 'strict'), code ASC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperRefCatalogImpl) FindExecutionIntensityByCode(ctx context.Context, code string) (*entity.PaperLLMConstraintExecIntensity, error) {
	var row entity.PaperLLMConstraintExecIntensity
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperRefCatalogImpl) FindAuditLevelByCode(ctx context.Context, code string) (*entity.PaperLLMConstraintGate, error) {
	var row entity.PaperLLMConstraintGate
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperRefCatalogImpl) FindLLMDescriptionByCode(ctx context.Context, code uint64) (*entity.PaperLLMDescription, error) {
	if code == 0 {
		return nil, nil
	}
	var row entity.PaperLLMDescription
	err := r.db.WithContext(ctx).Where("description_code = ? AND status = ?", code, "active").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperRefCatalogImpl) ListActiveVenues(ctx context.Context) ([]entity.PaperRefVenue, error) {
	var rows []entity.PaperRefVenue
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("name ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperRefCatalogImpl) FindVenueByCode(ctx context.Context, code string) (*entity.PaperRefVenue, error) {
	if code == "" {
		return nil, nil
	}
	var row entity.PaperRefVenue
	err := r.db.WithContext(ctx).Where("code = ? AND status = ?", code, "active").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperRefCatalogImpl) FindVenueByID(ctx context.Context, id uint64) (*entity.PaperRefVenue, error) {
	if id == 0 {
		return nil, nil
	}
	var row entity.PaperRefVenue
	err := r.db.WithContext(ctx).Where("id = ? AND status = ?", id, "active").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
