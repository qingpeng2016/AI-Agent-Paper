package order

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/common/billing"
	"github.com/qingpeng2016/ai-token-mall/common/constants"
	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
)

func resolveOrderUserSubscriptionID(
	ctx context.Context,
	orders repository.UserOrdersRepo,
	userID uint,
	orderType string,
	targetProduct *entity.Products,
	reqSubID uint,
) (uint, error) {
	switch orderType {
	case constants.OrderTypePurchase:
		return 0, nil
	case constants.OrderTypeRenewal:
		return resolveRenewalActiveSubscriptionID(ctx, orders, userID, targetProduct.ID, reqSubID)
	case constants.OrderTypeUpgrade:
		return resolveUpgradeActiveSubscriptionID(ctx, orders, userID, targetProduct, reqSubID)
	case constants.OrderTypeQuotaAddon:
		return resolveQuotaAddonActiveSubscriptionID(ctx, orders, userID, targetProduct.ID, reqSubID)
	default:
		return 0, errorx.ErrParamsError
	}
}

func subscriptionIsActive(sub *entity.UserSubscriptions) bool {
	return sub != nil && sub.Status == "active"
}

func upgradeOrderQuantity(
	ctx context.Context,
	orders repository.UserOrdersRepo,
	userID, userSubID uint,
	targetPeriodDays int,
) (int, error) {
	sub, err := orders.FindSubscriptionForUser(ctx, userID, userSubID)
	if err != nil {
		return 0, errorx.ErrDbError
	}
	if sub == nil || !subscriptionIsActive(sub) {
		return 0, errorx.ErrRenewNoSubscription
	}
	return billing.CountUpgradeBillingCycles(sub.PeriodEnd, sub.ExpiresAt, targetPeriodDays), nil
}

// resolveRenewalActiveSubscriptionID 续费：必传 user_subscription_id，订阅 active 且 product_id 与订单一致。
func resolveRenewalActiveSubscriptionID(ctx context.Context, orders repository.UserOrdersRepo, userID, productID, reqSubID uint) (uint, error) {
	if reqSubID == 0 {
		return 0, errorx.ErrParamsError
	}
	sub, err := orders.FindSubscriptionForUser(ctx, userID, reqSubID)
	if err != nil {
		return 0, errorx.ErrDbError
	}
	if sub == nil || !subscriptionIsActive(sub) {
		return 0, errorx.ErrRenewNoSubscription
	}
	if sub.ProductID != productID {
		return 0, errorx.ErrParamsError
	}
	return sub.ID, nil
}

// resolveUpgradeActiveSubscriptionID 升档：active 订阅，目标 SKU 须更高额度且与当前不同。
func resolveUpgradeActiveSubscriptionID(
	ctx context.Context,
	orders repository.UserOrdersRepo,
	userID uint,
	targetProduct *entity.Products,
	reqSubID uint,
) (uint, error) {
	sub, err := loadActiveSubscriptionForUser(ctx, orders, userID, reqSubID)
	if err != nil {
		return 0, err
	}
	if targetProduct == nil {
		return 0, errorx.ErrParamsError
	}
	if sub.ProductID == targetProduct.ID {
		return 0, errorx.ErrParamsError
	}
	if targetProduct.LimitTokens <= sub.LimitTokens {
		return 0, errorx.ErrParamsError
	}
	if targetProduct.ProductsCategoryName != "" &&
		sub.ProductsCategoryName != "" &&
		targetProduct.ProductsCategoryName != sub.ProductsCategoryName {
		return 0, errorx.ErrParamsError
	}
	return sub.ID, nil
}

// resolveQuotaAddonActiveSubscriptionID 加购额度：active 订阅，商品须与当前订阅 SKU 一致（与续费相同 product_id）。
func resolveQuotaAddonActiveSubscriptionID(
	ctx context.Context,
	orders repository.UserOrdersRepo,
	userID, productID, reqSubID uint,
) (uint, error) {
	sub, err := loadActiveSubscriptionForUser(ctx, orders, userID, reqSubID)
	if err != nil {
		return 0, err
	}
	if sub.ProductID != productID {
		return 0, errorx.ErrParamsError
	}
	return sub.ID, nil
}

func loadActiveSubscriptionForUser(
	ctx context.Context,
	orders repository.UserOrdersRepo,
	userID, reqSubID uint,
) (*entity.UserSubscriptions, error) {
	if reqSubID == 0 {
		return nil, errorx.ErrParamsError
	}
	sub, err := orders.FindSubscriptionForUser(ctx, userID, reqSubID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	if !subscriptionIsActive(sub) {
		return nil, errorx.ErrRenewNoSubscription
	}
	return sub, nil
}
