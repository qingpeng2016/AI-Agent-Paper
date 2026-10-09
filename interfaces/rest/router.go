package rest

import (
	"context"
	"errors"
	"net/http"
	"strings"

	ginMiddleware "github.com/qingpeng2016/ai-agent-paper/common/dederi/gin/middleware"
	"github.com/qingpeng2016/ai-agent-paper/common/notification"
	conf2 "github.com/qingpeng2016/ai-agent-paper/conf"
	"github.com/qingpeng2016/ai-agent-paper/interfaces/handler"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Router struct {
	httpServer             *http.Server
	setting                *conf2.Config
	userHandler            *handler.UserHandler
	inviteRebateHandler    *handler.InviteRebateHandler
	couponHandler          *handler.CouponHandler
	trackingHandler        *handler.TrackingHandler
	paperLiteratureHandler    *handler.PaperLiteratureHandler
	paperTopicDiscoveryHandler *handler.PaperTopicDiscoveryHandler
	paperManuscriptHandler       *handler.PaperManuscriptHandler
	paperLiteratureReviewHandler *handler.PaperLiteratureReviewHandler
}

func NewRouter(
	setting *conf2.Config,
	userHandler *handler.UserHandler,
	inviteRebateHandler *handler.InviteRebateHandler,
	couponHandler *handler.CouponHandler,
	trackingHandler *handler.TrackingHandler,
	paperLiteratureHandler *handler.PaperLiteratureHandler,
	paperTopicDiscoveryHandler *handler.PaperTopicDiscoveryHandler,
	paperManuscriptHandler *handler.PaperManuscriptHandler,
	paperLiteratureReviewHandler *handler.PaperLiteratureReviewHandler,
) *Router {
	return &Router{
		setting:                      setting,
		userHandler:                  userHandler,
		inviteRebateHandler:          inviteRebateHandler,
		couponHandler:                couponHandler,
		trackingHandler:              trackingHandler,
		paperLiteratureHandler:       paperLiteratureHandler,
		paperTopicDiscoveryHandler:   paperTopicDiscoveryHandler,
		paperManuscriptHandler:       paperManuscriptHandler,
		paperLiteratureReviewHandler: paperLiteratureReviewHandler,
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
		v1.GET("/coupon-campaigns/register-promo", r.couponHandler.RegisterPromo)
		v1.POST("/tracking/events", r.trackingHandler.ReportEvents)

		paper := v1.Group("/paper")
		{
			paper.GET("/topic-discovery/form-options", r.paperTopicDiscoveryHandler.GetFormOptions)
		}
	}

	paperAuth := engine.Group("/api/v1/paper", ginMiddleware.RequireAuth)
	{
		paperAuth.POST("/topic-discovery/run", r.paperTopicDiscoveryHandler.PostRun)
		paperAuth.GET("/topic-discovery/run/current", r.paperTopicDiscoveryHandler.GetCurrentRun)
		paperAuth.POST("/topic-discovery/run/cancel", r.paperTopicDiscoveryHandler.PostCancelRun)
		paperAuth.POST("/topic-discovery/commit-manuscript", r.paperTopicDiscoveryHandler.PostCommitManuscript)
		paperAuth.GET("/manuscripts", r.paperManuscriptHandler.GetManuscripts)
		paperAuth.POST("/manuscripts", r.paperManuscriptHandler.PostCreateManuscript)
		paperAuth.POST("/manuscripts/set-current", r.paperManuscriptHandler.PostSetCurrentManuscript)
		paperAuth.GET("/literature-reviews", r.paperLiteratureReviewHandler.GetLiteratureReviews)
		paperAuth.POST("/literature-reviews/soft-delete", r.paperLiteratureReviewHandler.PostSoftDeleteLiteratureReview)
	}

	userAuth := engine.Group("/api/v1", ginMiddleware.RequireAuth)
	{
		userAuth.GET("/users/me", r.userHandler.Me)
		userAuth.POST("/users/me/password", r.userHandler.ChangePassword)
		userAuth.GET("/users/coupons", r.couponHandler.ListMine)
		userAuth.GET("/users/wallet-flows", r.userHandler.ListWalletFlows)
		userAuth.GET("/users/invite-rebate/overview", r.inviteRebateHandler.Overview)
		userAuth.GET("/users/invite-rebate/members", r.inviteRebateHandler.ListMembers)
		userAuth.GET("/users/invite-rebate/commission-records", r.inviteRebateHandler.ListCommissionRecords)
		userAuth.GET("/users/invite-rebate/withdrawals", r.inviteRebateHandler.ListWithdrawals)
		userAuth.POST("/users/invite-rebate/withdrawals", r.inviteRebateHandler.CreateWithdrawal)
		userAuth.GET("/users/invite-rebate/payout-config", r.inviteRebateHandler.GetPayoutConfig)
		userAuth.POST("/users/invite-rebate/payout-config/upload", r.inviteRebateHandler.UploadPayoutQR)
		userAuth.POST("/users/invite-rebate/commission/transfer-to-balance", r.inviteRebateHandler.TransferCommissionToBalance)
	}

	thirdParty := engine.Group("/api/v1/third-party")
	{
		literature := thirdParty.Group("/literature")
		{
			literature.GET("/arxiv", r.paperLiteratureHandler.SearchArxiv)
			literature.GET("/openalex", r.paperLiteratureHandler.SearchOpenAlex)
			literature.GET("/semantic-scholar", r.paperLiteratureHandler.SearchSemanticScholar)
		}
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
