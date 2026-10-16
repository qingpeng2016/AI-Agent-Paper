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

func (r *PaperLLMImpl) ResolveActiveModel(ctx context.Context, configID uint) (*entity.PaperLLMModelConfig, error) {
	if configID > 0 {
		row, err := r.GetActiveModelByID(ctx, configID)
		if err != nil {
			return nil, err
		}
		if row != nil {
			return row, nil
		}
	}
	var row entity.PaperLLMModelConfig
	err := r.db.WithContext(ctx).
		Where("status = ? AND provider_code = ?", "active", "anthropic").
		Order("id DESC").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = r.db.WithContext(ctx).
			Where("status = ?", "active").
			Order("id DESC").
			First(&row).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperLLMImpl) InsertCallLog(ctx context.Context, row *entity.PaperLLMCallLog) error {
	if row == nil {
		return nil
	}
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *PaperLLMImpl) AddTokensUsed(ctx context.Context, modelConfigID uint, promptDelta, completionDelta int) error {
	if modelConfigID == 0 {
		return nil
	}
	if promptDelta <= 0 && completionDelta <= 0 {
		return nil
	}
	updates := map[string]any{}
	if promptDelta > 0 {
		updates["use_in_tokens"] = gorm.Expr("use_in_tokens + ?", promptDelta)
	}
	if completionDelta > 0 {
		updates["use_out_tokens"] = gorm.Expr("use_out_tokens + ?", completionDelta)
	}
	return r.db.WithContext(ctx).
		Model(&entity.PaperLLMModelConfig{}).
		Where("id = ?", modelConfigID).
		Updates(updates).Error
}
