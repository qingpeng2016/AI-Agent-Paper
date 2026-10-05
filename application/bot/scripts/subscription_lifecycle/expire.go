package subscriptionlifecycle

import (
	"context"
	"time"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

func (j *SubscriptionLifecycleJob) expireSubscription(ctx context.Context, subID uint, now time.Time) error {
	return j.tx.Transaction(ctx, func(tx *gorm.DB) error {
		sub, err := j.subs.FindActiveForUpdate(ctx, tx, subID)
		if err != nil {
			return err
		}
		if sub == nil {
			return nil
		}

		if !now.After(sub.ExpiresAt) {
			return nil
		}

		if err := j.subs.Update(ctx, tx, map[string]interface{}{"id = ?": sub.ID}, map[string]interface{}{
			"status":     "expired",
			"updated_at": now,
		}); err != nil {
			return err
		}

		if err := j.insertSubscriptionLifecycleNotification(ctx, tx, sub.UserID, sub, "subscription_expired", now); err != nil {
			return err
		}

		return j.expireMainAndSubKeysForSubscription(ctx, tx, sub.ID, now)
	})
}

func (j *SubscriptionLifecycleJob) insertSubscriptionLifecycleNotification(ctx context.Context, tx *gorm.DB, userID uint, sub *entity.UserSubscriptions, templateCode string, now time.Time) error {
	subID := sub.ID
	sentAt := now
	row := entity.UserNotifications{
		UserID:             userID,
		UserSubscriptionID: &subID,
		Channel:            "in_app",
		TemplateCode:       templateCode,
		Status:             "sent",
		SentAt:             &sentAt,
		CreatedAt:          now,
	}
	return j.notifications.Create(ctx, tx, &row)
}
