package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type UserCommissionRecordsImpl struct {
	db *gorm.DB
}

func NewUserCommissionRecordsImpl(db *gorm.DB) repository.UserCommissionRecordsRepo {
	return &UserCommissionRecordsImpl{db: db}
}

func (r *UserCommissionRecordsImpl) Create(ctx context.Context, tx *gorm.DB, row *entity.UserCommissionRecords) error {
	return repository.GormDB(ctx, r.db, tx).Create(row).Error
}

func (r *UserCommissionRecordsImpl) FindByOrderID(ctx context.Context, tx *gorm.DB, orderID uint) (*entity.UserCommissionRecords, error) {
	var row entity.UserCommissionRecords
	err := repository.GormDB(ctx, r.db, tx).Where("order_id = ?", orderID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *UserCommissionRecordsImpl) ListByInviterUserID(ctx context.Context, inviterUserID uint, offset, limit int) ([]entity.UserCommissionRecords, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	var rows []entity.UserCommissionRecords
	err := r.db.WithContext(ctx).
		Where("inviter_user_id = ?", inviterUserID).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *UserCommissionRecordsImpl) CountByInviterUserID(ctx context.Context, inviterUserID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.UserCommissionRecords{}).
		Where("inviter_user_id = ?", inviterUserID).Count(&n).Error
	return n, err
}
