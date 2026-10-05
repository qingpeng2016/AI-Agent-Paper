package invite_rebate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func (s *Service) CreateCommissionWithdrawal(ctx context.Context, userID uint, amountStr, channel string) (*response.CreateCommissionWithdrawalResp, error) {
	amount, err := parseCommissionAmountYuan(amountStr)
	if err != nil {
		return nil, err
	}
	channel = strings.TrimSpace(strings.ToLower(channel))
	if channel != "alipay" && channel != "wechat" {
		return nil, errorx.ErrParamsError
	}

	payout, err := s.payouts.FindByUserIDAndChannel(ctx, userID, channel)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	if payout == nil || !payout.HasQR() {
		return nil, errorx.ErrParamsError
	}

	var withdrawalRow entity.UserCommissionWithdrawals
	var commissionAfter decimal.Decimal

	err = s.tx.Transaction(ctx, func(tx *gorm.DB) error {
		now := time.Now()
		neg := amount.Neg()
		var txErr error
		commissionAfter, txErr = s.users.ApplyCommissionDelta(ctx, tx, userID, neg)
		if txErr != nil {
			if errors.Is(txErr, gorm.ErrRecordNotFound) {
				return errorx.ErrUserNotFound
			}
			if txErr.Error() == "insufficient commission balance" {
				return errorx.ErrParamsError
			}
			return errorx.ErrDbError
		}

		withdrawalRow = entity.UserCommissionWithdrawals{
			UserID:      userID,
			Amount:      amount,
			Channel:     channel,
			PayoutQrURL: fmt.Sprintf("internal:qr_image:%s", channel),
			Status:      "pending",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := s.withdraws.Create(ctx, tx, &withdrawalRow); err != nil {
			return errorx.ErrDbError
		}

		refType := "commission_withdrawal"
		refID := withdrawalRow.ID
		negAmount := amount.Neg()
		commBal := commissionAfter
		chLabel := "支付宝"
		if channel == "wechat" {
			chLabel = "微信"
		}
		remark := fmt.Sprintf("佣金提现申请 ¥%s · %s", amount.StringFixed(2), chLabel)
		if err := s.wallets.CreateFlow(ctx, tx, &entity.UserWalletFlows{
			UserID:       userID,
			Type:         "withdraw",
			Amount:       negAmount,
			BalanceAfter: &commBal,
			Currency:     "CNY",
			RefType:      &refType,
			RefID:        &refID,
			Remark:       &remark,
			CreatedAt:    now,
		}); err != nil {
			return errorx.ErrDbError
		}

		sentAt := now
		if err := s.notifications.Create(ctx, tx, &entity.UserNotifications{
			UserID:       userID,
			Channel:      "in_app",
			TemplateCode: "commission_withdraw_submitted",
			Status:       entity.NotificationInAppUnread,
			SentAt:       &sentAt,
			CreatedAt:    now,
		}); err != nil {
			return errorx.ErrDbError
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &response.CreateCommissionWithdrawalResp{
		Withdrawal: response.InviteWithdrawalItem{
			ID:        withdrawalRow.ID,
			Amount:    response.MoneyFrom(withdrawalRow.Amount),
			Channel:   withdrawalRow.Channel,
			Status:    withdrawalRow.Status,
			CreatedAt: formatDateTime(withdrawalRow.CreatedAt),
		},
		CommissionBalance: response.MoneyFrom(commissionAfter),
	}, nil
}

func parseCommissionAmountYuan(amountStr string) (decimal.Decimal, error) {
	amountStr = strings.TrimSpace(amountStr)
	if amountStr == "" {
		return decimal.Zero, errorx.ErrParamsError
	}
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return decimal.Zero, errorx.ErrParamsError
	}
	amount = amount.Round(2)
	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, errorx.ErrParamsError
	}
	const maxAmount = 1_000_000
	if amount.GreaterThan(decimal.NewFromInt(maxAmount)) {
		return decimal.Zero, errorx.ErrParamsError
	}
	return amount, nil
}
