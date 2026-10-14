package mysql

import (
	"context"
	"errors"
	"strings"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

type PaperStrategyImpl struct {
	db *gorm.DB
}

func NewPaperStrategyImpl(db *gorm.DB) repository.PaperStrategyRepo {
	return &PaperStrategyImpl{db: db}
}

func (r *PaperStrategyImpl) FindActiveRouteByBehaviorKey(ctx context.Context, behaviorKey string) (*entity.PaperStrategyRoute, error) {
	key := strings.TrimSpace(behaviorKey)
	if key == "" {
		return nil, nil
	}
	var row entity.PaperStrategyRoute
	err := r.db.WithContext(ctx).
		Preload("Package").
		Where("behavior_key = ? AND status = ?", key, "active").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperStrategyImpl) FindPackageByID(ctx context.Context, id uint64) (*entity.PaperStrategyPackage, error) {
	if id == 0 {
		return nil, nil
	}
	var row entity.PaperStrategyPackage
	err := r.db.WithContext(ctx).Where("id = ? AND status = ?", id, "active").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
