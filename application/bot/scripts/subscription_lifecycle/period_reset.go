package subscriptionlifecycle

import (
	"context"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/billing"
	"gorm.io/gorm"
)

func (j *SubscriptionLifecycleJob) periodResetSubscription(ctx context.Context, subID uint, now time.Time) error {
	return j.tx.Transaction(ctx, func(tx *gorm.DB) error {
		sub, err := j.subs.FindActiveForUpdate(ctx, tx, subID)
		if err != nil {
			return err
		}
		if sub == nil {
			return nil
		}

		if !now.After(sub.PeriodEnd) {
			return nil
		}
		if !sub.ExpiresAt.After(sub.PeriodEnd) {
			return nil
		}

		product, err := j.products.FindByID(ctx, tx, sub.ProductID)
		if err != nil {
			return err
		}
		if product == nil {
			return gorm.ErrRecordNotFound
		}
		days := billing.PeriodDays(product.PeriodDays, product.BillingPeriod)
		newStart := sub.PeriodEnd
		newEnd := billing.AddPeriod(newStart, days)

		if err := j.subs.Update(ctx, tx, map[string]interface{}{"id = ?": sub.ID}, map[string]interface{}{
			"period_start": newStart,
			"period_end":   newEnd,
			"limit_tokens": sub.BaseLimitTokens,
			"used_tokens":  0,
			"updated_at":   now,
		}); err != nil {
			return err
		}

		return j.periodResetMainAndSubKeysForSubscription(ctx, tx, sub.ID, sub.BaseLimitTokens, now)
	})
}
