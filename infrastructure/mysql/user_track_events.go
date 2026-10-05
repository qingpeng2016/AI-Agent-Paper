package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type UserTrackEventsImpl struct {
	db *gorm.DB
}

func NewUserTrackEventsImpl(db *gorm.DB) repository.UserTrackEventsRepo {
	return &UserTrackEventsImpl{db: db}
}

func (r *UserTrackEventsImpl) CreateBatch(ctx context.Context, tx *gorm.DB, rows []entity.UserTrackEvents) error {
	if len(rows) == 0 {
		return nil
	}
	return repository.GormDB(ctx, r.db, tx).CreateInBatches(rows, 50).Error
}
