package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type UserNotificationsImpl struct {
	db *gorm.DB
}

func NewUserNotificationsImpl(db *gorm.DB) repository.UserNotificationsRepo {
	return &UserNotificationsImpl{db: db}
}

func (r *UserNotificationsImpl) Create(ctx context.Context, tx *gorm.DB, m *entity.UserNotifications) error {
	return repository.GormDB(ctx, r.db, tx).Create(m).Error
}

func inAppInboxBase(ctx context.Context, db *gorm.DB, userID uint) *gorm.DB {
	return db.WithContext(ctx).Model(&entity.UserNotifications{}).
		Where("user_id = ? AND channel = ?", userID, "in_app").
		Where("status IN ?", []string{entity.NotificationInAppUnread, entity.NotificationInAppRead})
}

func (r *UserNotificationsImpl) ListInAppByUserID(ctx context.Context, userID uint, unreadOnly bool, offset, limit int) ([]entity.UserNotifications, error) {
	q := inAppInboxBase(ctx, r.db, userID)
	if unreadOnly {
		q = q.Where("status = ?", entity.NotificationInAppUnread)
	}
	var rows []entity.UserNotifications
	err := q.Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *UserNotificationsImpl) CountInAppByUserID(ctx context.Context, userID uint, unreadOnly bool) (int64, error) {
	q := inAppInboxBase(ctx, r.db, userID)
	if unreadOnly {
		q = q.Where("status = ?", entity.NotificationInAppUnread)
	}
	var n int64
	err := q.Count(&n).Error
	return n, err
}

func (r *UserNotificationsImpl) CountUnreadInAppByUserID(ctx context.Context, userID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.UserNotifications{}).
		Where("user_id = ? AND channel = ? AND status = ?", userID, "in_app", entity.NotificationInAppUnread).
		Count(&n).Error
	return n, err
}

func (r *UserNotificationsImpl) FindInAppByIDForUser(ctx context.Context, userID, id uint) (*entity.UserNotifications, error) {
	var row entity.UserNotifications
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ? AND channel = ?", id, userID, "in_app").
		Where("status IN ?", []string{entity.NotificationInAppUnread, entity.NotificationInAppRead}).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *UserNotificationsImpl) MarkRead(ctx context.Context, userID, id uint) error {
	return r.db.WithContext(ctx).Model(&entity.UserNotifications{}).
		Where("id = ? AND user_id = ? AND channel = ? AND status = ?", id, userID, "in_app", entity.NotificationInAppUnread).
		Update("status", entity.NotificationInAppRead).Error
}

func (r *UserNotificationsImpl) MarkAllReadInApp(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&entity.UserNotifications{}).
		Where("user_id = ? AND channel = ? AND status = ?", userID, "in_app", entity.NotificationInAppUnread).
		Update("status", entity.NotificationInAppRead).Error
}
