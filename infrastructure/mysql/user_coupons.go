package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type UserCouponsImpl struct {
	db *gorm.DB
}

func NewUserCouponsImpl(db *gorm.DB) repository.UserCouponsRepo {
	return &UserCouponsImpl{db: db}
}

func (r *UserCouponsImpl) Create(ctx context.Context, tx *gorm.DB, row *entity.UserCoupons) error {
	return repository.GormDB(ctx, r.db, tx).Create(row).Error
}

func (r *UserCouponsImpl) ExistsByUserAndCampaign(ctx context.Context, userID, campaignID uint) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.UserCoupons{}).
		Where("user_id = ? AND campaign_id = ?", userID, campaignID).
		Count(&n).Error
	return n > 0, err
}

func (r *UserCouponsImpl) ListByUserID(ctx context.Context, userID uint) ([]entity.UserCoupons, error) {
	var rows []entity.UserCoupons
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("id DESC").
		Find(&rows).Error
	return rows, err
}

func (r *UserCouponsImpl) FindByIDForUser(ctx context.Context, userID, couponID uint) (*entity.UserCoupons, error) {
	if couponID == 0 {
		return nil, nil
	}
	var row entity.UserCoupons
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", couponID, userID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *UserCouponsImpl) MarkUsed(ctx context.Context, tx *gorm.DB, couponID, orderID uint, usedAt time.Time) error {
	res := repository.GormDB(ctx, r.db, tx).Model(&entity.UserCoupons{}).
		Where("id = ? AND status = ?", couponID, "available").
		Updates(map[string]interface{}{
			"status":     "used",
			"used_at":    usedAt,
			"order_id":   orderID,
			"updated_at": usedAt,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserCouponsImpl) ExpireAvailableBefore(ctx context.Context, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Model(&entity.UserCoupons{}).
		Where("status = ? AND valid_until < ?", "available", before).
		Updates(map[string]interface{}{
			"status":     "expired",
			"updated_at": before,
		})
	return res.RowsAffected, res.Error
}
