package order

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	couponSvc "github.com/qingpeng2016/ai-token-mall/application/core-service/coupon"
	"github.com/qingpeng2016/ai-token-mall/common/billing"
	"github.com/qingpeng2016/ai-token-mall/common/constants"
	"github.com/qingpeng2016/ai-token-mall/common/money"
	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
	"gorm.io/datatypes"
)

type OrderService struct {
	orders   repository.UserOrdersRepo
	products repository.ProductsRepo
	fulfill  *OrderFulfillService
	coupons  *couponSvc.Service
}

func NewOrderService(
	orders repository.UserOrdersRepo,
	products repository.ProductsRepo,
	fulfill *OrderFulfillService,
	coupons *couponSvc.Service,
) *OrderService {
	return &OrderService{orders: orders, products: products, fulfill: fulfill, coupons: coupons}
}

func (s *OrderService) ListMine(ctx context.Context, userID uint, q *request.ListOrdersQuery) (*response.UserOrderListPageResp, error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	total, err := s.orders.CountByUserID(ctx, userID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	rows, err := s.orders.ListByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	items := make([]response.UserOrderListItem, 0, len(rows))
	for _, row := range rows {
		name := row.ProductCardTitle
		if name == "" {
			name = row.ProductSKUName
		}
		items = append(items, response.UserOrderListItem{
			ID:                row.ID,
			OrderNo:           row.OrderNo,
			OrderType:         row.OrderType,
			ProductName:       name,
			Quantity:          row.Quantity,
			TotalAmount:       response.MoneyFrom(row.TotalAmount),
			Status:            row.Status,
			EnterpriseInvoice: row.EnterpriseInvoice == 1,
			OutTradeNo:        row.OutTradeNo,
			CreatedAt:         formatOrderDateTime(row.CreatedAt),
		})
	}
	return &response.UserOrderListPageResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func formatOrderDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}

func (s *OrderService) CreateMockOrder(ctx context.Context, userID uint, req *request.CreateOrderReq) (*response.CreateOrderResp, error) {
	channel := normalizePayChannel(req.Channel)
	if channel == "" {
		return nil, errorx.ErrParamsError
	}
	orderType := normalizeOrderType(req.OrderType)
	if orderType == "" {
		return nil, errorx.ErrParamsError
	}

	product, err := s.products.FindOnSaleByID(ctx, req.ProductID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	if product == nil {
		return nil, errorx.ErrProductNotFound
	}

	userSubID, err := resolveOrderUserSubscriptionID(ctx, s.orders, userID, orderType, product, req.UserSubscriptionID)
	if err != nil {
		return nil, err
	}

	qty := req.Quantity
	if orderType == constants.OrderTypeUpgrade {
		computed, err := upgradeOrderQuantity(ctx, s.orders, userID, userSubID, billing.PeriodDays(product.PeriodDays, product.BillingPeriod))
		if err != nil {
			return nil, err
		}
		qty = computed
	}
	subtotal := money.MulQty(product.Price, qty)
	now := time.Now()
	couponDiscount := money.Zero
	var appliedCouponID uint
	if req.UserCouponID > 0 && s.coupons != nil {
		_, off, err := s.coupons.ResolveForCheckout(ctx, userID, req.UserCouponID, subtotal, now)
		if err != nil {
			return nil, err
		}
		couponDiscount = off
		appliedCouponID = req.UserCouponID
	}
	afterCoupon := subtotal.Sub(couponDiscount).Round(money.Scale)
	if afterCoupon.LessThan(money.Zero) {
		afterCoupon = money.Zero
	}
	total := afterCoupon
	if req.EnterpriseInvoice {
		total = money.AddInvoiceSurcharge(afterCoupon, 6)
	}

	orderNo := fmt.Sprintf("AP%d%04d", now.Unix(), userID%10000)
	outTradeNo := fmt.Sprintf("PAY%d%d", now.UnixNano()/1e6, userID)

	order := &entity.UserOrders{
		OrderNo:           orderNo,
		UserID:            userID,
		ProductID:         product.ID,
		OrderType:          orderType,
		UserSubscriptionID: userSubID,
		Quantity:           qty,
		UnitPrice:         product.Price,
		Status:            "pending_payment",
		TotalAmount:       total,
		Currency:          product.Currency,
		EnterpriseInvoice: boolToInt(req.EnterpriseInvoice),
		ExpireAt:          now.Add(30 * time.Minute),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	rawReq, _ := json.Marshal(map[string]interface{}{
		"product_id":           product.ID,
		"order_type":           orderType,
		"user_subscription_id": userSubID,
		"quantity":             qty,
		"channel":              channel,
		"enterprise_invoice":   req.EnterpriseInvoice,
		"user_coupon_id":       appliedCouponID,
		"coupon_discount":      couponDiscount.StringFixed(money.Scale),
		"subtotal":             subtotal.StringFixed(money.Scale),
	})
	order.PayChannel = channel
	order.OutTradeNo = outTradeNo
	order.RawRequestJSON = datatypes.JSON(rawReq)

	var createErr error
	switch orderType {
	case constants.OrderTypeRenewal:
		createErr = s.orders.CreateRenewalOrderReplacingPending(ctx, order)
	case constants.OrderTypeUpgrade:
		createErr = s.orders.CreateUpgradeOrderReplacingPending(ctx, order)
	case constants.OrderTypeQuotaAddon:
		createErr = s.orders.CreateQuotaAddonOrderReplacingPending(ctx, order)
	default:
		createErr = s.orders.CreateOrder(ctx, order)
	}
	if createErr != nil {
		return nil, errorx.ErrDbError
	}

	return createOrderRespFromEntity(order), nil
}

// MockCheckout 创建订单并模拟支付成功（开发/mock 收银台一步完成，避免只落单未回调）。
func (s *OrderService) MockCheckout(ctx context.Context, userID uint, req *request.CreateOrderReq) (*response.CreateOrderResp, error) {
	created, err := s.CreateMockOrder(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	notifyReq := &request.PaymentNotifyReq{
		OutTradeNo:  created.OutTradeNo,
		TradeStatus: "TRADE_SUCCESS",
	}
	if err := s.HandlePaymentNotify(ctx, created.Channel, notifyReq); err != nil {
		return nil, err
	}
	created.Status = "completed"
	return created, nil
}

func createOrderRespFromEntity(order *entity.UserOrders) *response.CreateOrderResp {
	return &response.CreateOrderResp{
		OrderNo:            order.OrderNo,
		OutTradeNo:         order.OutTradeNo,
		OrderID:            order.ID,
		OrderType:          order.OrderType,
		UserSubscriptionID: order.UserSubscriptionID,
		Channel:            order.PayChannel,
		Status:             order.Status,
		TotalAmount:        response.MoneyFrom(order.TotalAmount),
		Currency:           order.Currency,
	}
}

func (s *OrderService) HandlePaymentNotify(ctx context.Context, channel string, req *request.PaymentNotifyReq) error {
	channel = normalizePayChannel(channel)
	if channel == "" || strings.TrimSpace(req.OutTradeNo) == "" {
		return errorx.ErrParamsError
	}
	status := strings.ToUpper(strings.TrimSpace(req.TradeStatus))
	if status != "" && status != "TRADE_SUCCESS" && status != "SUCCESS" {
		return errorx.ErrParamsError
	}

	order, err := s.orders.FindOrderByOutTradeNo(ctx, req.OutTradeNo)
	if err != nil {
		return errorx.ErrDbError
	}
	if order == nil {
		return errorx.ErrPaymentNotFound
	}
	if order.PayChannel != channel {
		return errorx.ErrParamsError
	}

	third := strings.TrimSpace(req.TradeNo)
	if third == "" {
		third = fmt.Sprintf("MOCK_%s_%d", channel, time.Now().UnixNano())
	}
	payload, _ := json.Marshal(req)
	idem := fmt.Sprintf("%s:%s", channel, req.OutTradeNo)

	in := repository.PaymentNotifyInput{
		Channel:        channel,
		OutTradeNo:     req.OutTradeNo,
		ThirdTradeNo:   third,
		IdempotencyKey: idem,
		PayloadJSON:    payload,
	}
	if err := s.fulfill.ApplyPaymentNotifySuccess(ctx, in); err != nil {
		if strings.Contains(err.Error(), "not payable") {
			return errorx.ErrOrderNotPayable
		}
		if strings.Contains(err.Error(), "renew: subscription") && strings.Contains(err.Error(), "is not active") ||
			strings.Contains(err.Error(), "renew: subscription not found") {
			return errorx.ErrRenewNoSubscription
		}
		if strings.Contains(err.Error(), "quota_addon: subscription") && strings.Contains(err.Error(), "is not active") ||
			strings.Contains(err.Error(), "quota_addon: subscription not found") ||
			strings.Contains(err.Error(), "quota_addon: missing user_subscription_id") {
			return errorx.ErrRenewNoSubscription
		}
		if strings.Contains(err.Error(), "upgrade: subscription not found") ||
			strings.Contains(err.Error(), "upgrade: missing user_subscription_id") {
			return errorx.ErrRenewNoSubscription
		}
		if strings.Contains(err.Error(), "unknown order_type") {
			return errorx.ErrParamsError
		}
		return errorx.ErrPaymentFulfillFailed.WithDetail(err.Error())
	}
	return nil
}

func normalizeOrderType(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", constants.OrderTypePurchase:
		return constants.OrderTypePurchase
	case constants.OrderTypeRenewal:
		return constants.OrderTypeRenewal
	case constants.OrderTypeUpgrade:
		return constants.OrderTypeUpgrade
	case constants.OrderTypeQuotaAddon:
		return constants.OrderTypeQuotaAddon
	default:
		return ""
	}
}

func normalizePayChannel(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "alipay":
		return "alipay"
	case "wechat", "wxpay":
		return "wechat"
	case "paypal":
		return "paypal"
	default:
		return ""
	}
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
