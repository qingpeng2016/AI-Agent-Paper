package subscriptionlifecycle

import (
	"context"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/constants"
	"gorm.io/gorm"
)

func mainAndSubKeyTypes() []string {
	return []string{constants.APIKeyTypeMain, constants.APIKeyTypeSub}
}

// expireMainAndSubKeysForSubscription 套餐过期：主 Key 与同订阅下全部分配的子 Key 一并置为 expired。
func (j *SubscriptionLifecycleJob) expireMainAndSubKeysForSubscription(ctx context.Context, tx *gorm.DB, subscriptionID uint, now time.Time) error {
	return j.apiKeys.UpdateWhere(ctx, tx, map[string]interface{}{
		"user_subscription_id = ?": subscriptionID,
		"key_type IN ?":            mainAndSubKeyTypes(),
		"status <> ?":              "rotated",
	}, map[string]interface{}{
		"status":     "expired",
		"updated_at": now,
	})
}

// periodResetMainAndSubKeysForSubscription 套餐进入新周期：主 Key 同步套餐额度并清零用量；子 Key 仅清零 used_tokens（保留各自 limit_tokens）。
func (j *SubscriptionLifecycleJob) periodResetMainAndSubKeysForSubscription(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uint,
	mainLimitTokens int64,
	now time.Time,
) error {
	if err := j.apiKeys.UpdateWhere(ctx, tx, map[string]interface{}{
		"user_subscription_id = ?": subscriptionID,
		"key_type = ?":           constants.APIKeyTypeMain,
		"status <> ?":            "rotated",
	}, map[string]interface{}{
		"limit_tokens": mainLimitTokens,
		"used_tokens":  0,
		"updated_at":   now,
	}); err != nil {
		return err
	}
	return j.apiKeys.UpdateWhere(ctx, tx, map[string]interface{}{
		"user_subscription_id = ?": subscriptionID,
		"key_type = ?":           constants.APIKeyTypeSub,
		"status <> ?":            "rotated",
	}, map[string]interface{}{
		"used_tokens": 0,
		"updated_at":  now,
	})
}
