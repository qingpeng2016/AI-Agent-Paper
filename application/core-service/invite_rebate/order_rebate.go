package invite_rebate

import (
	"context"
	"fmt"
	"time"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// AccrueInviteRebateForPaidOrder 支付成功后在同一事务内为上级结算邀请返利（幂等：每订单一条）。
func (s *Service) AccrueInviteRebateForPaidOrder(
	ctx context.Context,
	tx *gorm.DB,
	order *entity.UserOrders,
	product *entity.Products,
	now time.Time,
) error {
	if order == nil || product == nil {
		return nil
	}
	existing, err := s.records.FindByOrderID(ctx, tx, order.ID)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	buyer, err := s.users.FindByIDForUpdate(ctx, tx, order.UserID)
	if err != nil {
		return err
	}
	if buyer == nil || buyer.ParentUserID == 0 {
		return nil
	}
	if buyer.ParentUserID == buyer.ID {
		return nil
	}

	inviterID := buyer.ParentUserID
	inviter, err := s.users.FindByIDForUpdate(ctx, tx, inviterID)
	if err != nil {
		return err
	}
	if inviter == nil {
		return nil
	}

	rate, err := s.rebateRateForInviter(ctx, inviter)
	if err != nil {
		return err
	}
	if rate.LessThanOrEqual(decimal.Zero) {
		return nil
	}

	orderAmount := order.TotalAmount
	rebate := orderAmount.Mul(rate).Div(decimal.NewFromInt(100)).Round(2)
	if rebate.LessThanOrEqual(decimal.Zero) {
		return nil
	}

	productName := product.CardTitle
	if productName == "" {
		productName = product.SKUProductName
	}

	record := entity.UserCommissionRecords{
		InviterUserID: inviterID,
		InviteeUserID: order.UserID,
		OrderID:       order.ID,
		OrderNo:       order.OrderNo,
		ProductName:   productName,
		OrderAmount:   orderAmount,
		RatePercent:   rate,
		RebateAmount:  rebate,
		Status:        "settled",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.records.Create(ctx, tx, &record); err != nil {
		return err
	}

	commissionAfter, err := s.users.ApplyCommissionDelta(ctx, tx, inviterID, rebate)
	if err != nil {
		return err
	}

	refType := "order_commission"
	refID := record.ID
	commBal := commissionAfter
	remark := fmt.Sprintf(
		"邀请返利 ¥%s · %s · 订单 %s",
		rebate.StringFixed(2),
		productName,
		order.OrderNo,
	)
	if err := s.wallets.CreateFlow(ctx, tx, &entity.UserWalletFlows{
		UserID:       inviterID,
		Type:         "commission",
		Amount:       rebate,
		BalanceAfter: &commBal,
		Currency:     "CNY",
		RefType:      &refType,
		RefID:        &refID,
		Remark:       &remark,
		CreatedAt:    now,
	}); err != nil {
		return err
	}

	sentAt := now
	return s.notifications.Create(ctx, tx, &entity.UserNotifications{
		UserID:       inviterID,
		Channel:      "in_app",
		TemplateCode: "commission_rebate_earned",
		Status:       entity.NotificationInAppUnread,
		SentAt:       &sentAt,
		CreatedAt:    now,
	})
}

func (s *Service) rebateRateForInviter(ctx context.Context, inviter *entity.Users) (decimal.Decimal, error) {
	if inviter == nil {
		return decimal.Zero, nil
	}
	tier, err := s.vipConfigs.FindByID(ctx, inviter.VipConfigID)
	if err != nil {
		return decimal.Zero, err
	}
	if tier != nil && tier.Enabled {
		return tier.RatePercent, nil
	}
	def, err := s.vipConfigs.FindDefault(ctx)
	if err != nil {
		return decimal.Zero, err
	}
	if def == nil {
		return decimal.Zero, nil
	}
	return def.RatePercent, nil
}
