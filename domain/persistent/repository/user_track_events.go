package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

type UserTrackEventsRepo interface {
	CreateBatch(ctx context.Context, tx *gorm.DB, rows []entity.UserTrackEvents) error
}
