package couponexpire

import (
	"context"
	"time"

	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"go.uber.org/zap"
)

const (
	ModuleAIAgentPaper = "ai_agent_paper"
	TaskCouponExpire  = "coupon_expire"
)

type CouponExpireJob struct {
	coupons repository.UserCouponsRepo
}

func NewCouponExpireJob(coupons repository.UserCouponsRepo) *CouponExpireJob {
	return &CouponExpireJob{coupons: coupons}
}

// Run 将 status=available 且 valid_until 已过的用户券标记为 expired。
func (j *CouponExpireJob) Run(ctx context.Context) {
	now := time.Now()
	n, err := j.coupons.ExpireAvailableBefore(ctx, now)
	if err != nil {
		logger.ErrorZ(ctx, "coupon-expire-run-failed", zap.Error(err))
		return
	}
	logger.InfoZ(ctx, "coupon-expire-run-finished", zap.Int64("expired_count", n))
}
