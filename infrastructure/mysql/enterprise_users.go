package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type EnterpriseUsersImpl struct {
	db *gorm.DB
}

func NewEnterpriseUsersImpl(db *gorm.DB) repository.EnterpriseUsersRepo {
	return &EnterpriseUsersImpl{db: db}
}

func (r *EnterpriseUsersImpl) Create(ctx context.Context, tx *gorm.DB, row *entity.EnterpriseUsers) error {
	return repository.GormDB(ctx, r.db, tx).Create(row).Error
}

func (r *EnterpriseUsersImpl) Update(ctx context.Context, tx *gorm.DB, row *entity.EnterpriseUsers) error {
	return repository.GormDB(ctx, r.db, tx).Save(row).Error
}

func (r *EnterpriseUsersImpl) FindByID(ctx context.Context, id uint) (*entity.EnterpriseUsers, error) {
	var row entity.EnterpriseUsers
	err := r.db.WithContext(ctx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *EnterpriseUsersImpl) FindByIDForOwner(ctx context.Context, ownerUserID, id uint) (*entity.EnterpriseUsers, error) {
	var row entity.EnterpriseUsers
	err := r.db.WithContext(ctx).
		Where("owner_user_id = ? AND id = ?", ownerUserID, id).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *EnterpriseUsersImpl) ListByOwnerUserID(ctx context.Context, ownerUserID uint, offset, limit int) ([]entity.EnterpriseUsers, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var rows []entity.EnterpriseUsers
	err := r.db.WithContext(ctx).
		Where("owner_user_id = ?", ownerUserID).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *EnterpriseUsersImpl) CountByOwnerUserID(ctx context.Context, ownerUserID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.EnterpriseUsers{}).
		Where("owner_user_id = ?", ownerUserID).Count(&n).Error
	return n, err
}

func (r *EnterpriseUsersImpl) FindActiveByLinkedUserID(ctx context.Context, userID uint) (*entity.EnterpriseUsers, error) {
	var row entity.EnterpriseUsers
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, "active").
		Order("id DESC").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *EnterpriseUsersImpl) LinkedUserIDsByOwner(ctx context.Context, ownerUserID uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&entity.EnterpriseUsers{}).
		Where("owner_user_id = ? AND user_id IS NOT NULL AND status = ?", ownerUserID, "active").
		Pluck("user_id", &ids).Error
	return ids, err
}
