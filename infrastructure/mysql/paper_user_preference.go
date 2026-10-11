package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaperUserPreferenceImpl struct {
	db *gorm.DB
}

func NewPaperUserPreferenceImpl(db *gorm.DB) repository.PaperUserPreferenceRepo {
	return &PaperUserPreferenceImpl{db: db}
}

func (r *PaperUserPreferenceImpl) GetByUserID(ctx context.Context, userID uint64) (*entity.PaperUserPreference, error) {
	if userID == 0 {
		return nil, nil
	}
	var row entity.PaperUserPreference
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperUserPreferenceImpl) Save(ctx context.Context, row *entity.PaperUserPreference) error {
	if row == nil {
		return nil
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			UpdateAll: true,
		}).
		Create(row).Error
}
