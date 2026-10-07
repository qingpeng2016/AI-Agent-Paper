package rest

import (
	"context"
	"errors"
	"net/http"
	"strings"

	ginMiddleware "github.com/qingpeng2016/ai-token-mall/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-token-mall/common/notification"
	conf2 "github.com/qingpeng2016/ai-token-mall/conf"
	"github.com/qingpeng2016/ai-token-mall/interfaces/handler"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Router struct {
	httpServer       *http.Server
	setting          *conf2.Config
	userHandler      *handler.UserHandler
	productHandler   *handler.ProductHandler
	tutorialHandler    *handler.TutorialHandler
	enterpriseHandler  *handler.EnterpriseHandler
	orderHandler          *handler.OrderHandler
	subscriptionHandler   *handler.SubscriptionHandler
	invoiceConfigHandler       *handler.InvoiceConfigHandler
	userNotificationHandler    *handler.UserNotificationHandler
	inviteRebateHandler        *handler.InviteRebateHandler
	userAPIKeyHandler          *handler.UserAPIKeyHandler
	couponHandler              *handler.CouponHandler
	trackingHandler            *handler.TrackingHandler
	paperLiteratureHandler     *handler.PaperLiteratureHandler
}

func NewRouter(
	setting *conf2.Config,
	userHandler *handler.UserHandler,
	productHandler *handler.ProductHandler,
	tutorialHandler *handler.TutorialHandler,
	enterpriseHandler *handler.EnterpriseHandler,
	orderHandler *handler.OrderHandler,
	subscriptionHandler *handler.SubscriptionHandler,
	invoiceConfigHandler *handler.InvoiceConfigHandler,
	userNotificationHandler *handler.UserNotificationHandler,
	inviteRebateHandler *handler.InviteRebateHandler,
	userAPIKeyHandler *handler.UserAPIKeyHandler,
	couponHandler *handler.CouponHandler,
	trackingHandler *handler.TrackingHandler,
	paperLiteratureHandler *handler.PaperLiteratureHandler,
) *Router {
	return &Router{
		setting:                   setting,
		userHandler:               userHandler,
		productHandler:            productHandler,
		tutorialHandler:           tutorialHandler,
		enterpriseHandler:         enterpriseHandler,
		orderHandler:              orderHandler,
		subscriptionHandler:       subscriptionHandler,
		invoiceConfigHandler:      invoiceConfigHandler,
		userNotificationHandler:   userNotificationHandler,
		inviteRebateHandler:     inviteRebateHandler,
		userAPIKeyHandler:       userAPIKeyHandler,
		couponHandler:           couponHandler,
		trackingHandler:         trackingHandler,
		paperLiteratureHandler:  paperLiteratureHandler,
	}
}

func (r *Router) setupRouters() *gin.Engine {
	engine := gin.Default()
	engine.Use(ginMiddleware.TraceRequestLog, ginMiddleware.CORSMiddleware)

	v1 := engine.Group("/api/v1")
	{
		v1.POST("/users/register", r.userHandler.Register)
		v1.POST("/users/login", r.userHandler.Login)
		v1.POST("/users/logout", r.userHandler.Logout)
		v1.GET("/products", r.productHandler.List)
		v1.GET("/products/nav-menu", r.productHandler.NavMenu)
		v1.GET("/products/slug/:slug", r.productHandler.DetailBySlug)
		v1.GET("/tutorials", r.tutorialHandler.List)
		v1.GET("/tutorials/articles/:slug", r.tutorialHandler.Detail)
		v1.GET("/enterprise/products", r.enterpriseHandler.ListProducts)
		v1.GET("/enterprise/products/:code", r.enterpriseHandler.GetProduct)
		v1.POST("/enterprise/inquiries", r.enterpriseHandler.SubmitInquiry)
		v1.GET("/coupon-campaigns/register-promo", r.couponHandler.RegisterPromo)
		v1.POST("/tracking/events", r.trackingHandler.ReportEvents)
	}

	userAuth := engine.Group("/api/v1", ginMiddleware.RequireAuth)
	{
		userAuth.GET("/users/me", r.userHandler.Me)
		userAuth.POST("/users/me/password", r.userHandler.ChangePassword)
		userAuth.GET("/users/notifications", r.userNotificationHandler.ListMine)
		userAuth.GET("/users/notifications/unread-count", r.userNotificationHandler.UnreadCount)
		userAuth.POST("/users/notifications/read-all", r.userNotificationHandler.MarkAllRead)
		userAuth.POST("/users/notifications/:id/read", r.userNotificationHandler.MarkRead)
		userAuth.GET("/users/coupons", r.couponHandler.ListMine)
		userAuth.GET("/users/wallet-flows", r.userHandler.ListWalletFlows)
		userAuth.GET("/users/invoices", r.userHandler.ListInvoices)
		userAuth.GET("/users/invoice-configs", r.invoiceConfigHandler.ListMine)
		userAuth.GET("/users/invoice-configs/enterprise-lookup", r.invoiceConfigHandler.LookupEnterprise)
		userAuth.POST("/users/invoice-configs", r.invoiceConfigHandler.Create)
		userAuth.PUT("/users/invoice-configs/:id", r.invoiceConfigHandler.Update)
		userAuth.POST("/users/invoice-configs/:id/set-default", r.invoiceConfigHandler.SetDefault)
		userAuth.GET("/users/invite-rebate/overview", r.inviteRebateHandler.Overview)
		userAuth.GET("/users/invite-rebate/members", r.inviteRebateHandler.ListMembers)
		userAuth.GET("/users/invite-rebate/commission-records", r.inviteRebateHandler.ListCommissionRecords)
		userAuth.GET("/users/invite-rebate/withdrawals", r.inviteRebateHandler.ListWithdrawals)
		userAuth.POST("/users/invite-rebate/withdrawals", r.inviteRebateHandler.CreateWithdrawal)
		userAuth.GET("/users/invite-rebate/payout-config", r.inviteRebateHandler.GetPayoutConfig)
		userAuth.POST("/users/invite-rebate/payout-config/upload", r.inviteRebateHandler.UploadPayoutQR)
		userAuth.POST("/users/invite-rebate/commission/transfer-to-balance", r.inviteRebateHandler.TransferCommissionToBalance)
		userAuth.GET("/users/api-keys/main", r.userAPIKeyHandler.ListMain)
		userAuth.GET("/users/api-keys/team", r.userAPIKeyHandler.ListTeam)
		userAuth.POST("/users/api-keys/sub", r.userAPIKeyHandler.CreateSub)
		userAuth.PATCH("/users/api-keys/:id/limit", r.userAPIKeyHandler.UpdateSubLimit)
		userAuth.GET("/users/api-team/members", r.userAPIKeyHandler.ListTeamMembers)
		userAuth.GET("/users/api-team/addable-invitees", r.userAPIKeyHandler.ListAddableInvitees)
		userAuth.POST("/users/api-team/members", r.userAPIKeyHandler.AddTeamMember)
		userAuth.GET("/users/api-team/enterprise-inquiry", r.userAPIKeyHandler.GetEnterpriseInquiry)
		userAuth.POST("/users/api-team/enterprise-inquiry", r.userAPIKeyHandler.SubmitEnterpriseInquiry)
	}

	// 服务端代请求外部 API（文献库等）；联调测试暂不鉴权
	thirdParty := engine.Group("/api/v1/third-party")
	{
		literature := thirdParty.Group("/literature")
		{
			literature.GET("/arxiv", r.paperLiteratureHandler.SearchArxiv)
			literature.GET("/openalex", r.paperLiteratureHandler.SearchOpenAlex)
			literature.GET("/semantic-scholar", r.paperLiteratureHandler.SearchSemanticScholar)
		}
	}

	// 需登录；正式网关回调另开 /api/v1/payments/notify 且无鉴权
	authGroup := engine.Group("/api/v1/mock", ginMiddleware.RequireAuth)
	{
		authGroup.GET("/subscriptions", r.subscriptionHandler.ListMine)
		authGroup.GET("/orders", r.orderHandler.ListMine)
		authGroup.POST("/orders", r.orderHandler.CreateOrder)
		authGroup.POST("/orders/checkout", r.orderHandler.MockCheckout)
		authGroup.POST("/payments/notify/:channel", r.orderHandler.PaymentNotify)
	}

	return engine
}

func (r *Router) Run(setting *conf2.Server) {
	gin.SetMode(strings.ToLower(setting.RunMode))
	go func() {
		r.httpServer = &http.Server{Addr: setting.Port, Handler: r.setupRouters()}
		if err := r.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			notification.SendErrorLog(context.Background(), "run http server failed", zap.String("error", err.Error()))
			panic(err)
		}
	}()
}

func (r *Router) Close() {
	if r.httpServer == nil {
		return
	}
	if err := r.httpServer.Shutdown(context.Background()); err != nil {
		notification.SendErrorLog(context.Background(), "stop http server failed", zap.String("error", err.Error()))
	}
}
