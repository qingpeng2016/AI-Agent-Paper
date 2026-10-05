package subscriptionlifecycle

import (
	"context"
	"sync"
	"time"

	"github.com/qingpeng2016/ai-token-mall/application/core-service/order"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"go.uber.org/zap"
)

const (
	ModuleAITokenMall         = "ai_token_mall"
	TaskSubscriptionLifecycle = "subscription_lifecycle"
	lifecycleBatchSize        = 100
	autoRenewLeadDays         = 7
)

type subscriptionLifecycleHandler func(ctx context.Context, subID uint, now time.Time) error

type SubscriptionLifecycleJob struct {
	tx            repository.Transactor
	subs          repository.UserSubscriptionsRepo
	products      repository.ProductsRepo
	users         repository.UsersRepo
	orders        repository.UserOrdersRepo
	apiKeys       repository.UserAPIKeysRepo
	notifications repository.UserNotificationsRepo
	fulfill       *order.OrderFulfillService
}

func NewSubscriptionLifecycleJob(
	tx repository.Transactor,
	subs repository.UserSubscriptionsRepo,
	products repository.ProductsRepo,
	users repository.UsersRepo,
	orders repository.UserOrdersRepo,
	apiKeys repository.UserAPIKeysRepo,
	notifications repository.UserNotificationsRepo,
	fulfill *order.OrderFulfillService,
) *SubscriptionLifecycleJob {
	return &SubscriptionLifecycleJob{
		tx:            tx,
		subs:          subs,
		products:      products,
		users:         users,
		orders:        orders,
		apiKeys:       apiKeys,
		notifications: notifications,
		fulfill:       fulfill,
	}
}

// Run 扫描 active 订阅，依次：自动续费 → 周期重置 → 过期处理。
func (j *SubscriptionLifecycleJob) Run(ctx context.Context) {
	now := time.Now()
	if err := j.runAutoRenewPhase(ctx, now); err != nil {
		logger.ErrorZ(ctx, "subscription-lifecycle-auto-renew-phase-failed", zap.Error(err))
		return
	}
	if err := j.runPeriodResetPhase(ctx, now); err != nil {
		logger.ErrorZ(ctx, "subscription-lifecycle-period-reset-phase-failed", zap.Error(err))
		return
	}
	if err := j.runExpirePhase(ctx, now); err != nil {
		logger.ErrorZ(ctx, "subscription-lifecycle-expire-phase-failed", zap.Error(err))
		return
	}
	logger.InfoZ(ctx, "subscription-lifecycle-run-finished")
}

func (j *SubscriptionLifecycleJob) listActiveSubscriptionIDs(ctx context.Context) ([]uint, error) {
	return j.subs.ListActiveIDs(ctx)
}

func (j *SubscriptionLifecycleJob) runActiveByBatch(
	ctx context.Context,
	now time.Time,
	logKey string,
	handler subscriptionLifecycleHandler,
) error {
	ids, err := j.listActiveSubscriptionIDs(ctx)
	if err != nil {
		return err
	}
	for start := 0; start < len(ids); start += lifecycleBatchSize {
		end := start + lifecycleBatchSize
		if end > len(ids) {
			end = len(ids)
		}
		batch := ids[start:end]
		var wg sync.WaitGroup
		for _, id := range batch {
			wg.Add(1)
			subID := id
			go func() {
				defer wg.Done()
				if err := handler(ctx, subID, now); err != nil {
					logger.ErrorZ(ctx, logKey, zap.Uint("subscription_id", subID), zap.Error(err))
				}
			}()
		}
		wg.Wait()
	}
	return nil
}

func (j *SubscriptionLifecycleJob) runAutoRenewPhase(ctx context.Context, now time.Time) error {
	return j.runActiveByBatch(ctx, now, "subscription-auto-renew-failed", j.autoRenewSubscription)
}

func (j *SubscriptionLifecycleJob) runPeriodResetPhase(ctx context.Context, now time.Time) error {
	return j.runActiveByBatch(ctx, now, "subscription-period-reset-failed", j.periodResetSubscription)
}

func (j *SubscriptionLifecycleJob) runExpirePhase(ctx context.Context, now time.Time) error {
	return j.runActiveByBatch(ctx, now, "subscription-expire-failed", j.expireSubscription)
}

func subscriptionTimeEqual(a, b time.Time) bool {
	return a.Unix() == b.Unix()
}
