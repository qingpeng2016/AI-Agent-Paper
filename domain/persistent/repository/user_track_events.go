package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"gorm.io/gorm"
)

type UserTrackEventsRepo interface {
	CreateBatch(ctx context.Context, tx *gorm.DB, rows []entity.UserTrackEvents) error
}
