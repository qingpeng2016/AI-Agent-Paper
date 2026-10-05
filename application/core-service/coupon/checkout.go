package coupon

import (
	"context"
	"encoding/json"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/common/money"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func DiscountAmount(coupon *entity.UserCoupons, subtotal decimal.Decimal) decimal.Decimal {
	if coupon == nil || subtotal.LessThanOrEqual(money.Zero) {
		return money.Zero
	}
	if subtotal.LessThan(coupon.MinOrderAmount) {
		return money.Zero
	}
	var off decimal.Decimal
	switch coupon.DiscountType {
	case "percent":
		off = subtotal.Mul(coupon.DiscountValue).Div(decimal.NewFromInt(100))
	default:
		off = coupon.DiscountValue
	}
	off = off.Round(money.Scale)
	if off.GreaterThan(subtotal) {
		return subtotal.Round(money.Scale)
	}
	if off.LessThan(money.Zero) {
		return money.Zero
	}
	return off
}

func (s *Service) ResolveForCheckout(
	ctx context.Context,
	userID, couponID uint,
	subtotal decimal.Decimal,
	now time.Time,
) (*entity.UserCoupons, decimal.Decimal, error) {
	if couponID == 0 {
		return nil, money.Zero, nil
	}
	row, err := s.coupons.FindByIDForUser(ctx, userID, couponID)
	if err != nil {
		return nil, money.Zero, err
	}
	if row == nil || !isCouponUsable(row, now) {
		return nil, money.Zero, errorx.ErrCouponUnavailable
	}
	off := DiscountAmount(row, subtotal)
	if off.LessThanOrEqual(money.Zero) {
		return nil, money.Zero, errorx.ErrCouponUnavailable
	}
	return row, off, nil
}

func isCouponUsable(row *entity.UserCoupons, now time.Time) bool {
	if row.Status != "available" {
		return false
	}
	if now.Before(row.ValidFrom) {
		return false
	}
	if now.After(row.ValidUntil) {
		return false
	}
	return true
}

type orderRawCoupon struct {
	UserCouponID uint `json:"user_coupon_id"`
}

func (s *Service) MarkUsedFromOrderRawRequest(
	ctx context.Context,
	tx *gorm.DB,
	userID, orderID uint,
	rawRequestJSON []byte,
	usedAt time.Time,
) error {
	if len(rawRequestJSON) == 0 {
		return nil
	}
	var raw orderRawCoupon
	if err := json.Unmarshal(rawRequestJSON, &raw); err != nil || raw.UserCouponID == 0 {
		return nil
	}
	row, err := s.coupons.FindByIDForUser(ctx, userID, raw.UserCouponID)
	if err != nil {
		return err
	}
	if row == nil || row.Status != "available" {
		return nil
	}
	return s.coupons.MarkUsed(ctx, tx, raw.UserCouponID, orderID, usedAt)
}
