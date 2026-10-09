package boot

import (
	bot "github.com/qingpeng2016/ai-agent-paper/application/bot"
	couponexpire "github.com/qingpeng2016/ai-agent-paper/application/bot/scripts/coupon_expire"
	topicaudit "github.com/qingpeng2016/ai-agent-paper/application/bot/scripts/topic_audit"
	topicgenerateideas "github.com/qingpeng2016/ai-agent-paper/application/bot/scripts/topic_generate_ideas"
	topicliteraturedownload "github.com/qingpeng2016/ai-agent-paper/application/bot/scripts/topic_literature_download"
	viplevelsync "github.com/qingpeng2016/ai-agent-paper/application/bot/scripts/vip_level_sync"
	botscheduleconfig "github.com/qingpeng2016/ai-agent-paper/application/core-service/bot_schedule_config"
	inviteRebateSvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/invite_rebate"
	papersvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper"
	couponSvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/coupon"
	trackingSvc "github.com/qingpeng2016/ai-agent-paper/application/core-service/tracking"
	"github.com/qingpeng2016/ai-agent-paper/application/core-service/user"
	log2 "github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/common/notification"
	"github.com/qingpeng2016/ai-agent-paper/conf"
	"github.com/qingpeng2016/ai-agent-paper/infrastructure/http"
	anthropicinfra "github.com/qingpeng2016/ai-agent-paper/infrastructure/http/anthropic"
	arxivinfra "github.com/qingpeng2016/ai-agent-paper/infrastructure/http/arxiv"
	googleinfra "github.com/qingpeng2016/ai-agent-paper/infrastructure/http/google"
	openalexinfra "github.com/qingpeng2016/ai-agent-paper/infrastructure/http/openalex"
	openaichatinfra "github.com/qingpeng2016/ai-agent-paper/infrastructure/http/openaichat"
	semanticscholarinfra "github.com/qingpeng2016/ai-agent-paper/infrastructure/http/semanticscholar"
	"github.com/qingpeng2016/ai-agent-paper/infrastructure/mysql"
	"github.com/qingpeng2016/ai-agent-paper/infrastructure/redis"
	"github.com/qingpeng2016/ai-agent-paper/interfaces/handler"
	"github.com/qingpeng2016/ai-agent-paper/interfaces/rest"

	"go.uber.org/dig"
	"go.uber.org/zap/zapcore"
)

func init() {
	log2.NewLogger("ai-agent-paper", "./log", zapcore.DebugLevel)
}

func BuildContainer() *dig.Container {
	c := dig.New()

	cfg := conf.NewCfg()
	_ = c.Provide(func() *conf.Config { return cfg })

	// HTTP
	_ = c.Provide(rest.NewRouter)
	_ = c.Provide(handler.NewUserHandler)
	_ = c.Provide(handler.NewInviteRebateHandler)
	_ = c.Provide(handler.NewCouponHandler)
	_ = c.Provide(handler.NewTrackingHandler)
	_ = c.Provide(handler.NewPaperLiteratureHandler)
	_ = c.Provide(handler.NewPaperTopicDiscoveryHandler)
	_ = c.Provide(handler.NewPaperManuscriptHandler)
	_ = c.Provide(handler.NewPaperLiteratureReviewHandler)
	_ = c.Provide(trackingSvc.NewService)
	_ = c.Provide(papersvc.NewLiteratureSearchService)
	_ = c.Provide(papersvc.NewTopicDiscoveryOptionsService)
	_ = c.Provide(papersvc.NewTopicDiscoveryRunService)
	_ = c.Provide(papersvc.NewManuscriptService)
	_ = c.Provide(papersvc.NewLiteratureReviewService)
	_ = c.Provide(papersvc.NewLLMChatService)
	_ = c.Provide(couponSvc.NewService)
	_ = c.Provide(inviteRebateSvc.NewService)
	_ = c.Provide(user.NewUserService)
	_ = c.Provide(botscheduleconfig.NewBotScheduleConfigService)

	// Bot
	_ = c.Provide(viplevelsync.NewVipLevelSyncJob)
	_ = c.Provide(couponexpire.NewCouponExpireJob)
	_ = c.Provide(topicliteraturedownload.NewTopicLiteratureDownloadJob)
	_ = c.Provide(topicgenerateideas.NewTopicGenerateIdeasJob)
	_ = c.Provide(topicaudit.NewTopicAuditJob)
	_ = c.Provide(bot.NewScheduler)
	_ = c.Provide(bot.NewEntry)

	// Infra
	_ = c.Provide(NewDBClient)
	_ = c.Provide(mysql.NewUsersImpl)
	_ = c.Provide(mysql.NewUserWalletFlowsImpl)
	_ = c.Provide(mysql.NewVipConfigImpl)
	_ = c.Provide(mysql.NewVipDomainConfigImpl)
	_ = c.Provide(mysql.NewUserCommissionPayoutConfigImpl)
	_ = c.Provide(mysql.NewUserCommissionRecordsImpl)
	_ = c.Provide(mysql.NewUserCommissionWithdrawalsImpl)
	_ = c.Provide(mysql.NewTransactorImpl)
	_ = c.Provide(mysql.NewBotScheduleConfigImpl)
	_ = c.Provide(mysql.NewCouponCampaignsImpl)
	_ = c.Provide(mysql.NewUserCouponsImpl)
	_ = c.Provide(mysql.NewUserTrackEventsImpl)
	_ = c.Provide(mysql.NewPaperRefLiteratureSourceImpl)
	_ = c.Provide(mysql.NewPaperRefCatalogImpl)
	_ = c.Provide(mysql.NewPaperManuscriptImpl)
	_ = c.Provide(mysql.NewPaperOutputTopicStepImpl)
	_ = c.Provide(mysql.NewPaperOutputLiteratureReviewImpl)
	_ = c.Provide(mysql.NewPaperLLMImpl)
	_ = c.Provide(redis.NewClient)
	_ = c.Provide(http.NewHTTPClient)
	_ = c.Provide(arxivinfra.NewClient)
	_ = c.Provide(openalexinfra.NewClient)
	_ = c.Provide(semanticscholarinfra.NewClient)
	_ = c.Provide(anthropicinfra.NewClient)
	_ = c.Provide(googleinfra.NewClient)
	_ = c.Provide(openaichatinfra.NewClient)

	_ = c.Provide(notification.NewNotificationManager)

	return c
}
