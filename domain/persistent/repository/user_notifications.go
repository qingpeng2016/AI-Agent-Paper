package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

type UserNotificationsRepo interface {
	Create(ctx context.Context, tx *gorm.DB, m *entity.UserNotifications) error
	ListInAppByUserID(ctx context.Context, userID uint, unreadOnly bool, offset, limit int) ([]entity.UserNotifications, error)
	CountInAppByUserID(ctx context.Context, userID uint, unreadOnly bool) (int64, error)
	CountUnreadInAppByUserID(ctx context.Context, userID uint) (int64, error)
	FindInAppByIDForUser(ctx context.Context, userID, id uint) (*entity.UserNotifications, error)
	MarkRead(ctx context.Context, userID, id uint) error
	MarkAllReadInApp(ctx context.Context, userID uint) error
}
