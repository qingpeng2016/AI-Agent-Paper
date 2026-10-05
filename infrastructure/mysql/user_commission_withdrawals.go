package mysql

import (
	"context"
	"time"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type UserCommissionWithdrawalsImpl struct {
	db *gorm.DB
}

func NewUserCommissionWithdrawalsImpl(db *gorm.DB) repository.UserCommissionWithdrawalsRepo {
	return &UserCommissionWithdrawalsImpl{db: db}
}

func (r *UserCommissionWithdrawalsImpl) Create(ctx context.Context, tx *gorm.DB, row *entity.UserCommissionWithdrawals) error {
	return repository.GormDB(ctx, r.db, tx).Create(row).Error
}

func (r *UserCommissionWithdrawalsImpl) ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]entity.UserCommissionWithdrawals, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	var rows []entity.UserCommissionWithdrawals
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *UserCommissionWithdrawalsImpl) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.UserCommissionWithdrawals{}).
		Where("user_id = ?", userID).Count(&n).Error
	return n, err
}

func (r *UserCommissionWithdrawalsImpl) UpdateStatus(ctx context.Context, tx *gorm.DB, id uint, status string, failReason *string, processedAt *time.Time) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if failReason != nil {
		updates["fail_reason"] = *failReason
	}
	if processedAt != nil {
		updates["processed_at"] = *processedAt
	}
	return repository.GormDB(ctx, r.db, tx).Model(&entity.UserCommissionWithdrawals{}).
		Where("id = ?", id).Updates(updates).Error
}
