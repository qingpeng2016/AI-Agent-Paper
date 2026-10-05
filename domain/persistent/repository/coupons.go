package repository

import (
	"context"
	"time"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

type CouponCampaignsRepo interface {
	ListAutoGrantOnRegister(ctx context.Context) ([]entity.CouponCampaigns, error)
	FindByID(ctx context.Context, id uint) (*entity.CouponCampaigns, error)
}

type UserCouponsRepo interface {
	Create(ctx context.Context, tx *gorm.DB, row *entity.UserCoupons) error
	ExistsByUserAndCampaign(ctx context.Context, userID, campaignID uint) (bool, error)
	ListByUserID(ctx context.Context, userID uint) ([]entity.UserCoupons, error)
	FindByIDForUser(ctx context.Context, userID, couponID uint) (*entity.UserCoupons, error)
	MarkUsed(ctx context.Context, tx *gorm.DB, couponID, orderID uint, usedAt time.Time) error
	ExpireAvailableBefore(ctx context.Context, before time.Time) (int64, error)
}
