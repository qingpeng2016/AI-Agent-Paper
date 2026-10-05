package invite_rebate

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func (s *Service) TransferCommissionToBalance(ctx context.Context, userID uint, amountStr string) (*response.CommissionTransferToBalanceResp, error) {
	amount, err := parseCommissionAmountYuan(amountStr)
	if err != nil {
		return nil, err
	}

	var commissionAfter, walletAfter decimal.Decimal
	err = s.tx.Transaction(ctx, func(tx *gorm.DB) error {
		var txErr error
		commissionAfter, walletAfter, txErr = s.users.TransferCommissionToWallet(ctx, tx, userID, amount)
		if txErr != nil {
			if errors.Is(txErr, gorm.ErrRecordNotFound) {
				return errorx.ErrUserNotFound
			}
			if txErr.Error() == "insufficient commission balance" || txErr.Error() == "invalid transfer amount" {
				return errorx.ErrParamsError
			}
			return errorx.ErrDbError
		}

		now := time.Now()
		refType := "commission_transfer"
		refID := uint(now.UnixNano() % 0x7fffffff)
		neg := amount.Neg()
		commBal := commissionAfter
		remarkOut := fmt.Sprintf("佣金划转至余额 ¥%s", amount.StringFixed(2))
		if err := s.wallets.CreateFlow(ctx, tx, &entity.UserWalletFlows{
			UserID:       userID,
			Type:         "commission",
			Amount:       neg,
			BalanceAfter: &commBal,
			Currency:     "CNY",
			RefType:      &refType,
			RefID:        &refID,
			Remark:       &remarkOut,
			CreatedAt:    now,
		}); err != nil {
			return errorx.ErrDbError
		}

		walBal := walletAfter
		remarkIn := fmt.Sprintf("佣金划转入账 ¥%s", amount.StringFixed(2))
		if err := s.wallets.CreateFlow(ctx, tx, &entity.UserWalletFlows{
			UserID:       userID,
			Type:         "recharge",
			Amount:       amount,
			BalanceAfter: &walBal,
			Currency:     "CNY",
			RefType:      &refType,
			RefID:        &refID,
			Remark:       &remarkIn,
			CreatedAt:    now,
		}); err != nil {
			return errorx.ErrDbError
		}

		sentAt := now
		if err := s.notifications.Create(ctx, tx, &entity.UserNotifications{
			UserID:       userID,
			Channel:      "in_app",
			TemplateCode: "commission_transferred_to_wallet",
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

	return &response.CommissionTransferToBalanceResp{
		CommissionBalance: response.MoneyFrom(commissionAfter),
		WalletBalance:     response.MoneyFrom(walletAfter),
		TransferredAmount: response.MoneyFrom(amount),
	}, nil
}
