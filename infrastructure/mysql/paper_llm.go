package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

type PaperLLMImpl struct {
	db *gorm.DB
}

func NewPaperLLMImpl(db *gorm.DB) repository.PaperLLMRepo {
	return &PaperLLMImpl{db: db}
}

func (r *PaperLLMImpl) GetActiveModelByID(ctx context.Context, id uint) (*entity.PaperLLMModelConfig, error) {
	var row entity.PaperLLMModelConfig
	err := r.db.WithContext(ctx).
		Where("id = ? AND status = ?", id, "active").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperLLMImpl) GetActiveBindingByStage(ctx context.Context, stageCode string) (*entity.PaperLLMWorkflowBinding, error) {
	var row entity.PaperLLMWorkflowBinding
	err := r.db.WithContext(ctx).
		Where("stage_code = ? AND status = ?", stageCode, "active").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperLLMImpl) GetActivePromptByStage(ctx context.Context, stageCode string) (*entity.PaperLLMPromptTemplate, error) {
	var row entity.PaperLLMPromptTemplate
	err := r.db.WithContext(ctx).
		Where("stage_code = ? AND status = ?", stageCode, "active").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
