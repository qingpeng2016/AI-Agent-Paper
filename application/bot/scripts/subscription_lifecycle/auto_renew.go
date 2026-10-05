package subscriptionlifecycle

import (
	"context"
	"fmt"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/constants"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

func (j *SubscriptionLifecycleJob) autoRenewSubscription(ctx context.Context, subID uint, now time.Time) error {
	return j.tx.Transaction(ctx, func(tx *gorm.DB) error {
		sub, err := j.subs.FindActiveForUpdate(ctx, tx, subID)
		if err != nil {
			return err
		}
		if sub == nil {
			return nil
		}

		if !now.Before(sub.ExpiresAt) {
			return nil
		}
		lead := sub.ExpiresAt.Add(-autoRenewLeadDays * 24 * time.Hour)
		if now.Before(lead) {
			return nil
		}
		if !subscriptionTimeEqual(sub.PeriodEnd, sub.ExpiresAt) {
			return nil
		}

		product, err := j.products.FindByID(ctx, tx, sub.ProductID)
		if err != nil {
			return err
		}
		if product == nil {
			return gorm.ErrRecordNotFound
		}
		price := product.Price
		user, err := j.users.FindByIDForUpdate(ctx, tx, sub.UserID)
		if err != nil {
			return err
		}
		if user == nil {
			return gorm.ErrRecordNotFound
		}
		if user.WalletBalance.LessThan(price) {
			return nil
		}

		orderNo := fmt.Sprintf("RN%d%04d", now.Unix(), sub.UserID%10000)
		outTradeNo := fmt.Sprintf("BALRN%d%d", now.UnixNano()/1e6, sub.ID)
		paidAt := now
		order := entity.UserOrders{
			OrderNo:            orderNo,
			UserID:             sub.UserID,
			ProductID:          sub.ProductID,
			OrderType:          constants.OrderTypeRenewal,
			UserSubscriptionID: sub.ID,
			Quantity:           1,
			UnitPrice:          price,
			Status:             "completed",
			TotalAmount:        price,
			Currency:           product.Currency,
			PayChannel:         "balance",
			OutTradeNo:         outTradeNo,
			PaidAt:             &paidAt,
			ExpireAt:           now.Add(30 * time.Minute),
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		if err := j.orders.CreateTx(ctx, tx, &order); err != nil {
			return err
		}

		// 余额扣款、流水、续期通知 + 上级邀请返利（与支付回调 AccrueInviteRebateForPaidOrder 一致）
		return j.fulfill.FulfillBalanceRenewalInTx(ctx, tx, &order, product, now)
	})
}
