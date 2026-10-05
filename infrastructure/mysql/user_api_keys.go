package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type UserAPIKeysImpl struct {
	db *gorm.DB
}

func NewUserAPIKeysImpl(db *gorm.DB) repository.UserAPIKeysRepo {
	return &UserAPIKeysImpl{db: db}
}

func (r *UserAPIKeysImpl) Create(ctx context.Context, tx *gorm.DB, m *entity.UserAPIKeys) error {
	return repository.GormDB(ctx, r.db, tx).Create(m).Error
}

func (r *UserAPIKeysImpl) UpdateWhere(ctx context.Context, tx *gorm.DB, where map[string]interface{}, updateData map[string]interface{}) error {
	db := repository.GormDB(ctx, r.db, tx).Model(entity.UserAPIKeys{})
	for condition, value := range where {
		db = db.Where(condition, value)
	}
	return db.Updates(updateData).Error
}

func (r *UserAPIKeysImpl) FindByID(ctx context.Context, id uint) (*entity.UserAPIKeys, error) {
	var row entity.UserAPIKeys
	err := r.db.WithContext(ctx).First(&row, id).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ListMainByUserID 当前用户持有的 Key（含自建主 Key + 他人分配的团队子 Key）。
func (r *UserAPIKeysImpl) ListMainByUserID(ctx context.Context, userID uint) ([]entity.UserAPIKeys, error) {
	var rows []entity.UserAPIKeys
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status <> ?", userID, "rotated").
		Order("key_type ASC, id DESC").
		Find(&rows).Error
	return rows, err
}

func (r *UserAPIKeysImpl) ListSubByOwnerUserID(ctx context.Context, ownerUserID uint) ([]entity.UserAPIKeys, error) {
	var rows []entity.UserAPIKeys
	err := r.db.WithContext(ctx).
		Where("owner_user_id = ? AND key_type = ? AND status <> ?", ownerUserID, "sub", "rotated").
		Order("id DESC").
		Find(&rows).Error
	return rows, err
}

func (r *UserAPIKeysImpl) SumActiveLimitTokensBySubscription(ctx context.Context, subscriptionID uint, excludeKeyID uint) (int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&entity.UserAPIKeys{}).
		Where("user_subscription_id = ? AND status <> ?", subscriptionID, "rotated")
	if excludeKeyID > 0 {
		q = q.Where("id <> ?", excludeKeyID)
	}
	err := q.Select("COALESCE(SUM(limit_tokens), 0)").Scan(&total).Error
	return total, err
}

func (r *UserAPIKeysImpl) SumActiveSubKeyLimitTokensBySubscription(ctx context.Context, subscriptionID uint, excludeKeyID uint) (int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&entity.UserAPIKeys{}).
		Where("user_subscription_id = ? AND key_type = ? AND status <> ?", subscriptionID, "sub", "rotated")
	if excludeKeyID > 0 {
		q = q.Where("id <> ?", excludeKeyID)
	}
	err := q.Select("COALESCE(SUM(limit_tokens), 0)").Scan(&total).Error
	return total, err
}

func (r *UserAPIKeysImpl) ExistsActiveSubKey(ctx context.Context, subscriptionID, memberUserID uint) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.UserAPIKeys{}).
		Where("user_subscription_id = ? AND user_id = ? AND key_type = ? AND status = ?",
			subscriptionID, memberUserID, "sub", "active").
		Count(&n).Error
	return n > 0, err
}
