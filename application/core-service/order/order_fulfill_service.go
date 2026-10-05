package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	couponSvc "github.com/qingpeng2016/ai-token-mall/application/core-service/coupon"
	"github.com/qingpeng2016/ai-token-mall/application/core-service/invite_rebate"
	"github.com/qingpeng2016/ai-token-mall/common/apikey"
	"github.com/qingpeng2016/ai-token-mall/common/billing"
	"github.com/qingpeng2016/ai-token-mall/common/constants"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type OrderFulfillService struct {
	tx            repository.Transactor
	orders        repository.UserOrdersRepo
	products      repository.ProductsRepo
	subs          repository.UserSubscriptionsRepo
	users         repository.UsersRepo
	wallets       repository.UserWalletFlowsRepo
	notifications repository.UserNotificationsRepo
	apiKeys       repository.UserAPIKeysRepo
	callbacks     repository.PaymentCallbacksRepo
	invoices      repository.UserInvoicesRepo
	inviteRebate *invite_rebate.Service
	coupons      *couponSvc.Service
}

func NewOrderFulfillService(
	tx repository.Transactor,
	orders repository.UserOrdersRepo,
	products repository.ProductsRepo,
	subs repository.UserSubscriptionsRepo,
	users repository.UsersRepo,
	wallets repository.UserWalletFlowsRepo,
	notifications repository.UserNotificationsRepo,
	apiKeys repository.UserAPIKeysRepo,
	callbacks repository.PaymentCallbacksRepo,
	invoices repository.UserInvoicesRepo,
	inviteRebate *invite_rebate.Service,
	coupons *couponSvc.Service,
) *OrderFulfillService {
	return &OrderFulfillService{
		tx:            tx,
		orders:        orders,
		products:      products,
		subs:          subs,
		users:         users,
		wallets:       wallets,
		notifications: notifications,
		apiKeys:       apiKeys,
		callbacks:     callbacks,
		invoices:      invoices,
		inviteRebate: inviteRebate,
		coupons:      coupons,
	}
}

func (s *OrderFulfillService) ApplyPaymentNotifySuccess(ctx context.Context, in repository.PaymentNotifyInput) error {
	return s.tx.Transaction(ctx, func(tx *gorm.DB) error {
		existed, err := s.callbacks.FindByIdempotency(ctx, tx, in.Channel, in.IdempotencyKey)
		if err != nil {
			return err
		}
		if existed != nil {
			return nil
		}

		order, err := s.orders.FindByOutTradeNoForUpdate(ctx, tx, in.OutTradeNo)
		if err != nil {
			return err
		}
		if order == nil {
			return gorm.ErrRecordNotFound
		}
		if order.Status == "completed" {
			_ = s.insertCallback(ctx, tx, in, 1, "ignored")
			return nil
		}
		if order.Status != "pending_payment" {
			return fmt.Errorf("order %s status %s not payable", order.OrderNo, order.Status)
		}

		product, err := s.products.FindByID(ctx, tx, order.ProductID)
		if err != nil {
			return err
		}
		if product == nil {
			return gorm.ErrRecordNotFound
		}

		now := time.Now()
		notifyRaw := datatypes.JSON(in.PayloadJSON)
		if len(in.PayloadJSON) == 0 {
			notifyRaw = datatypes.JSON([]byte("{}"))
		}

		if err := s.orders.UpdateFields(ctx, tx, order.ID, map[string]interface{}{
			"status":          "completed",
			"paid_at":         now,
			"third_trade_no":  in.ThirdTradeNo,
			"raw_notify_json": notifyRaw,
			"updated_at":      now,
		}); err != nil {
			return err
		}
		order.Status = "completed"
		order.PaidAt = &now
		third := in.ThirdTradeNo
		order.ThirdTradeNo = &third
		order.RawNotifyJSON = notifyRaw
		order.UpdatedAt = now

		if err := s.fulfillPaidOrderByType(ctx, tx, order, product, in.Channel, now); err != nil {
			return err
		}

		if s.coupons != nil {
			if err := s.coupons.MarkUsedFromOrderRawRequest(ctx, tx, order.UserID, order.ID, order.RawRequestJSON, now); err != nil {
				return err
			}
		}

		if err := s.inviteRebate.AccrueInviteRebateForPaidOrder(ctx, tx, order, product, now); err != nil {
			return err
		}

		if order.EnterpriseInvoice != 0 {
			inv := entity.UserInvoices{
				OrderID:     order.ID,
				UserID:      order.UserID,
				InvoiceType: "enterprise_vat",
				Title:       "企业开票信息待补充",
				Amount:      order.TotalAmount,
				Status:      "pending",
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			if err := s.invoices.Create(ctx, tx, &inv); err != nil {
				return err
			}
		}

		return s.insertCallback(ctx, tx, in, 1, "success")
	})
}

// FulfillBalanceRenewalInTx Bot 余额自动续费：订单已 completed，仅履约续费链路。
func (s *OrderFulfillService) FulfillBalanceRenewalInTx(
	ctx context.Context,
	tx *gorm.DB,
	order *entity.UserOrders,
	product *entity.Products,
	now time.Time,
) error {
	if err := s.deductBalanceAndRecordFlow(ctx, tx, order, product, "balance", now); err != nil {
		return err
	}
	sub, err := s.renewSubscriptionForOrder(ctx, tx, order, product, now)
	if err != nil {
		return err
	}
	if err := s.reactivateSubscriptionAPIKeys(ctx, tx, sub.ID, now); err != nil {
		return err
	}
	if err := s.insertSubscriptionOrderNotification(ctx, tx, order, sub, nil, "subscription_renewed", now); err != nil {
		return err
	}
	return s.inviteRebate.AccrueInviteRebateForPaidOrder(ctx, tx, order, product, now)
}

func (s *OrderFulfillService) insertCallback(ctx context.Context, tx *gorm.DB, in repository.PaymentNotifyInput, sigOK int, result string) error {
	payload := in.PayloadJSON
	if len(payload) == 0 {
		payload = []byte("{}")
	}
	row := entity.PaymentCallbacks{
		Channel:        in.Channel,
		IdempotencyKey: in.IdempotencyKey,
		PayloadJSON:    datatypes.JSON(payload),
		SignatureOK:    sigOK,
		ProcessResult:  result,
		ProcessedAt:    time.Now(),
	}
	return s.callbacks.Create(ctx, tx, &row)
}

func (s *OrderFulfillService) fulfillPaidOrderByType(
	ctx context.Context,
	tx *gorm.DB,
	order *entity.UserOrders,
	product *entity.Products,
	payChannel string,
	now time.Time,
) error {
	switch order.OrderType {
	case constants.OrderTypeRenewal:
		return s.fulfillRenewalOrderPaid(ctx, tx, order, product, payChannel, now)
	case constants.OrderTypeUpgrade:
		return s.fulfillUpgradeOrderPaid(ctx, tx, order, product, payChannel, now)
	case constants.OrderTypeQuotaAddon:
		return s.fulfillQuotaAddonOrderPaid(ctx, tx, order, product, payChannel, now)
	case constants.OrderTypePurchase, "":
		return s.fulfillPurchaseOrderPaid(ctx, tx, order, product, payChannel, now)
	default:
		return fmt.Errorf("order %s unknown order_type %q", order.OrderNo, order.OrderType)
	}
}

func (s *OrderFulfillService) fulfillPurchaseOrderPaid(ctx context.Context, tx *gorm.DB, order *entity.UserOrders, product *entity.Products, payChannel string, now time.Time) error {
	sub, err := s.createSubscriptionForPurchase(ctx, tx, order, product, now)
	if err != nil {
		return err
	}
	apiKey, err := s.createAPIKeyForNewSubscription(ctx, tx, order, sub, product, now)
	if err != nil {
		return err
	}
	if err := s.insertSubscriptionOrderNotification(ctx, tx, order, sub, apiKey, "subscription_activated", now); err != nil {
		return err
	}
	return s.deductBalanceAndRecordFlow(ctx, tx, order, product, payChannel, now)
}

func (s *OrderFulfillService) createSubscriptionForPurchase(ctx context.Context, tx *gorm.DB, order *entity.UserOrders, product *entity.Products, now time.Time) (*entity.UserSubscriptions, error) {
	tokenGrant := product.LimitTokens * int64(order.Quantity)
	days := billing.PeriodDays(product.PeriodDays, product.BillingPeriod)
	qty := order.Quantity
	if qty < 1 {
		qty = 1
	}
	periodEnd := billing.AddPeriods(now, days, qty)
	ordersJSON, _ := json.Marshal([]uint{order.ID})

	row := entity.UserSubscriptions{
		UserID:               order.UserID,
		ProductID:            order.ProductID,
		Orders:               datatypes.JSON(ordersJSON),
		ProductsCategoryName: product.ProductsCategoryName,
		SKUProductName:       product.SKUProductName,
		BaseLimitTokens:      tokenGrant,
		LimitTokens:          tokenGrant,
		UsedTokens:           0,
		StartedAt:            now,
		ExpiresAt:            periodEnd,
		PeriodStart:          now,
		PeriodEnd:            periodEnd,
		Status:               "active",
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := s.subs.Create(ctx, tx, &row); err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *OrderFulfillService) createAPIKeyForNewSubscription(ctx context.Context, tx *gorm.DB, order *entity.UserOrders, sub *entity.UserSubscriptions, product *entity.Products, now time.Time) (*entity.UserAPIKeys, error) {
	_, hash, err := apikey.Generate()
	if err != nil {
		return nil, err
	}
	ownerUserID, enterpriseInquiryID, err := MainAPIKeyEnterpriseFields(ctx, tx, order.UserID)
	if err != nil {
		return nil, err
	}
	row := entity.UserAPIKeys{
		UserID:               order.UserID,
		OwnerUserID:          ownerUserID,
		EnterpriseInquiryID:  enterpriseInquiryID,
		UserSubscriptionID:   sub.ID,
		KeyType:              constants.APIKeyTypeMain,
		KeyHash:              hash,
		ProductsCategoryName: product.ProductsCategoryName,
		LimitTokens:          sub.LimitTokens,
		UsedTokens:           0,
		Status:               "active",
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := s.apiKeys.Create(ctx, tx, &row); err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *OrderFulfillService) fulfillRenewalOrderPaid(ctx context.Context, tx *gorm.DB, order *entity.UserOrders, product *entity.Products, payChannel string, now time.Time) error {
	sub, err := s.renewSubscriptionForOrder(ctx, tx, order, product, now)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("renew: subscription not found for order %s", order.OrderNo)
		}
		return err
	}
	if err := s.reactivateSubscriptionAPIKeys(ctx, tx, sub.ID, now); err != nil {
		return err
	}
	if err := s.insertSubscriptionOrderNotification(ctx, tx, order, sub, nil, "subscription_renewed", now); err != nil {
		return err
	}
	return s.deductBalanceAndRecordFlow(ctx, tx, order, product, payChannel, now)
}

func (s *OrderFulfillService) renewSubscriptionForOrder(ctx context.Context, tx *gorm.DB, order *entity.UserOrders, product *entity.Products, now time.Time) (*entity.UserSubscriptions, error) {
	sub, err := s.findSubscriptionForRenewal(ctx, tx, order)
	if err != nil {
		return nil, err
	}
	if sub.ProductID != order.ProductID {
		return nil, fmt.Errorf("renew: subscription %d product mismatch", sub.ID)
	}
	if sub.Status != "active" {
		return nil, fmt.Errorf("renew: subscription %d is not active", sub.ID)
	}

	var orderIDs []uint
	_ = json.Unmarshal(sub.Orders, &orderIDs)
	orderIDs = append(orderIDs, order.ID)
	ordersJSON, _ := json.Marshal(orderIDs)

	days := billing.PeriodDays(product.PeriodDays, product.BillingPeriod)
	anchor := now
	if sub.ExpiresAt.After(now) {
		anchor = sub.ExpiresAt
	}
	qty := order.Quantity
	if qty < 1 {
		qty = 1
	}
	newExpires := billing.AddPeriods(anchor, days, qty)

	if err := s.subs.Update(ctx, tx, map[string]interface{}{"id = ?": sub.ID}, map[string]interface{}{
		"orders":     datatypes.JSON(ordersJSON),
		"expires_at": newExpires,
		"updated_at": now,
	}); err != nil {
		return nil, err
	}
	return s.subs.Reload(ctx, tx, sub.ID)
}

func (s *OrderFulfillService) findSubscriptionForRenewal(ctx context.Context, tx *gorm.DB, order *entity.UserOrders) (*entity.UserSubscriptions, error) {
	if order.UserSubscriptionID == 0 {
		return nil, fmt.Errorf("renew: missing user_subscription_id on order %s", order.OrderNo)
	}
	sub, err := s.subs.FindOne(ctx, tx, map[string]interface{}{
		"id = ?":      order.UserSubscriptionID,
		"user_id = ?": order.UserID,
	}, "")
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (s *OrderFulfillService) fulfillUpgradeOrderPaid(ctx context.Context, tx *gorm.DB, order *entity.UserOrders, product *entity.Products, payChannel string, now time.Time) error {
	sub, wasActive, err := s.upgradeSubscriptionForOrder(ctx, tx, order, product, now)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("upgrade: subscription not found for order %s", order.OrderNo)
		}
		return err
	}
	if err := s.reactivateSubscriptionAPIKeys(ctx, tx, sub.ID, now); err != nil {
		return err
	}
	if err := s.syncAPIKeysAfterUpgrade(ctx, tx, sub.ID, product, sub.LimitTokens, wasActive, now); err != nil {
		return err
	}
	if err := s.insertSubscriptionOrderNotification(ctx, tx, order, sub, nil, "subscription_upgraded", now); err != nil {
		return err
	}
	return s.deductBalanceAndRecordFlow(ctx, tx, order, product, payChannel, now)
}

func (s *OrderFulfillService) upgradeSubscriptionForOrder(ctx context.Context, tx *gorm.DB, order *entity.UserOrders, product *entity.Products, now time.Time) (*entity.UserSubscriptions, bool, error) {
	if order.UserSubscriptionID == 0 {
		return nil, false, fmt.Errorf("upgrade: missing user_subscription_id on order %s", order.OrderNo)
	}

	sub, err := s.subs.FindOne(ctx, tx, map[string]interface{}{
		"id = ?":      order.UserSubscriptionID,
		"user_id = ?": order.UserID,
	}, "")
	if err != nil {
		return nil, false, err
	}

	tokenGrant := product.LimitTokens

	var orderIDs []uint
	_ = json.Unmarshal(sub.Orders, &orderIDs)
	orderIDs = append(orderIDs, order.ID)
	ordersJSON, _ := json.Marshal(orderIDs)

	wasActive := sub.Status == "active"

	updates := map[string]interface{}{
		"product_id":             product.ID,
		"orders":                 datatypes.JSON(ordersJSON),
		"products_category_name": product.ProductsCategoryName,
		"sku_product_name":       product.SKUProductName,
		"base_limit_tokens":      tokenGrant,
		"limit_tokens":           tokenGrant,
		"updated_at":             now,
	}

	if wasActive {
		// period / expires 不变
	} else {
		days := billing.PeriodDays(product.PeriodDays, product.BillingPeriod)
		qty := order.Quantity
		if qty < 1 {
			qty = 1
		}
		periodEnd := billing.AddPeriods(now, days, qty)
		updates["used_tokens"] = 0
		updates["started_at"] = now
		updates["expires_at"] = periodEnd
		updates["period_start"] = now
		updates["period_end"] = periodEnd
		updates["status"] = "active"
	}

	if err := s.subs.Update(ctx, tx, map[string]interface{}{"id = ?": sub.ID}, updates); err != nil {
		return nil, false, err
	}
	reloaded, err := s.subs.Reload(ctx, tx, sub.ID)
	if err != nil {
		return nil, false, err
	}
	return reloaded, wasActive, nil
}

func (s *OrderFulfillService) syncAPIKeysAfterUpgrade(ctx context.Context, tx *gorm.DB, subscriptionID uint, product *entity.Products, limitTokens int64, wasActiveBeforeUpgrade bool, now time.Time) error {
	keyUpdates := map[string]interface{}{
		"products_category_name": product.ProductsCategoryName,
		"limit_tokens":           limitTokens,
		"updated_at":             now,
	}
	if !wasActiveBeforeUpgrade {
		keyUpdates["used_tokens"] = 0
	}
	return s.apiKeys.UpdateWhere(ctx, tx, map[string]interface{}{
		"user_subscription_id = ?": subscriptionID,
		"status <> ?":              "rotated",
	}, keyUpdates)
}

func (s *OrderFulfillService) fulfillQuotaAddonOrderPaid(ctx context.Context, tx *gorm.DB, order *entity.UserOrders, product *entity.Products, payChannel string, now time.Time) error {
	sub, err := s.loadSubscriptionForQuotaAddon(ctx, tx, order)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("quota_addon: subscription not found")
		}
		return err
	}
	if sub.Status != "active" {
		return fmt.Errorf("quota_addon: subscription %d is not active", sub.ID)
	}

	addon := product.LimitTokens
	if err := s.subs.Update(ctx, tx, map[string]interface{}{"id = ?": sub.ID}, map[string]interface{}{
		"limit_tokens": gorm.Expr("limit_tokens + ?", addon),
		"updated_at":   now,
	}); err != nil {
		return err
	}
	reloaded, err := s.subs.Reload(ctx, tx, sub.ID)
	if err != nil {
		return err
	}
	if err := s.apiKeys.UpdateWhere(ctx, tx, map[string]interface{}{
		"user_subscription_id = ?": sub.ID,
		"status <> ?":              "rotated",
	}, map[string]interface{}{
		"limit_tokens": gorm.Expr("limit_tokens + ?", addon),
		"updated_at":   now,
	}); err != nil {
		return err
	}
	if err := s.insertSubscriptionOrderNotification(ctx, tx, order, reloaded, nil, "subscription_quota_added", now); err != nil {
		return err
	}
	return s.deductBalanceAndRecordFlow(ctx, tx, order, product, payChannel, now)
}

func (s *OrderFulfillService) loadSubscriptionForQuotaAddon(ctx context.Context, tx *gorm.DB, order *entity.UserOrders) (*entity.UserSubscriptions, error) {
	if order.UserSubscriptionID == 0 {
		return nil, fmt.Errorf("quota_addon: missing user_subscription_id on order %s", order.OrderNo)
	}
	sub, err := s.subs.FindOne(ctx, tx, map[string]interface{}{
		"id = ?":      order.UserSubscriptionID,
		"user_id = ?": order.UserID,
	}, "")
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (s *OrderFulfillService) reactivateSubscriptionAPIKeys(ctx context.Context, tx *gorm.DB, subscriptionID uint, now time.Time) error {
	return s.apiKeys.UpdateWhere(ctx, tx, map[string]interface{}{
		"user_subscription_id = ?": subscriptionID,
		"status <> ?":              "rotated",
	}, map[string]interface{}{
		"status":     "active",
		"updated_at": now,
	})
}

func (s *OrderFulfillService) insertSubscriptionOrderNotification(ctx context.Context, tx *gorm.DB, order *entity.UserOrders, sub *entity.UserSubscriptions, apiKey *entity.UserAPIKeys, templateCode string, now time.Time) error {
	subID := sub.ID
	sentAt := now
	row := entity.UserNotifications{
		UserID:             order.UserID,
		UserSubscriptionID: &subID,
		Channel:            "in_app",
		TemplateCode:       templateCode,
		Status:             "sent",
		SentAt:             &sentAt,
		CreatedAt:          now,
	}
	if apiKey != nil {
		keyID := apiKey.ID
		row.APIKeyID = &keyID
	}
	return s.notifications.Create(ctx, tx, &row)
}

// deductBalanceAndRecordFlow 记买家支付流水；余额支付同时扣 wallet_balance，三方支付 balance_after 为 NULL。
func (s *OrderFulfillService) deductBalanceAndRecordFlow(ctx context.Context, tx *gorm.DB, order *entity.UserOrders, product *entity.Products, channel string, now time.Time) error {
	negAmount := order.TotalAmount.Neg()
	var balanceAfter *decimal.Decimal
	if channel == "balance" {
		bal, err := s.users.ApplyWalletDelta(ctx, tx, order.UserID, negAmount)
		if err != nil {
			return err
		}
		balanceAfter = &bal
	}

	refType := "order"
	remark := fmt.Sprintf("订单 %s · %s · %s", order.OrderNo, product.CardTitle, payChannelFlowLabel(channel))
	row := entity.UserWalletFlows{
		UserID:       order.UserID,
		Type:         "pay",
		Amount:       negAmount,
		BalanceAfter: balanceAfter,
		Currency:     order.Currency,
		RefType:      &refType,
		RefID:        &order.ID,
		Remark:       &remark,
		CreatedAt:    now,
	}
	return s.wallets.CreateFlow(ctx, tx, &row)
}

func payChannelFlowLabel(channel string) string {
	switch channel {
	case "balance":
		return "余额"
	case "alipay":
		return "支付宝"
	case "wechat":
		return "微信支付"
	default:
		if channel == "" {
			return "在线支付"
		}
		return channel
	}
}
