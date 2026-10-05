package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserSubscriptionsImpl struct {
	db *gorm.DB
}

func NewUserSubscriptionsImpl(db *gorm.DB) repository.UserSubscriptionsRepo {
	return &UserSubscriptionsImpl{db: db}
}

func (r *UserSubscriptionsImpl) ListByUserID(ctx context.Context, userID uint) ([]repository.UserSubscriptionsListRow, error) {
	var rows []repository.UserSubscriptionsListRow
	err := r.db.WithContext(ctx).
		Table("user_subscriptions AS s").
		Select("s.*, p.card_title AS product_card_title").
		Joins("LEFT JOIN products p ON p.id = s.product_id").
		Where("s.user_id = ?", userID).
		Order("s.id DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *UserSubscriptionsImpl) ListActiveIDs(ctx context.Context) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&entity.UserSubscriptions{}).
		Where("status = ?", "active").
		Order("id ASC").
		Pluck("id", &ids).Error
	return ids, err
}

func (r *UserSubscriptionsImpl) FindOne(ctx context.Context, tx *gorm.DB, where map[string]interface{}, orderBy string) (entity.UserSubscriptions, error) {
	db := repository.GormDB(ctx, r.db, tx).Model(entity.UserSubscriptions{})
	for condition, value := range where {
		db = db.Where(condition, value)
	}
	if orderBy != "" {
		db = db.Order(orderBy)
	}
	var m entity.UserSubscriptions
	err := db.First(&m).Error
	return m, err
}

func (r *UserSubscriptionsImpl) FindActiveForUpdate(ctx context.Context, tx *gorm.DB, id uint) (*entity.UserSubscriptions, error) {
	var row entity.UserSubscriptions
	err := repository.GormDB(ctx, r.db, tx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND status = ?", id, "active").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *UserSubscriptionsImpl) Create(ctx context.Context, tx *gorm.DB, m *entity.UserSubscriptions) error {
	return repository.GormDB(ctx, r.db, tx).Create(m).Error
}

func (r *UserSubscriptionsImpl) Update(ctx context.Context, tx *gorm.DB, where map[string]interface{}, updateData map[string]interface{}) error {
	db := repository.GormDB(ctx, r.db, tx).Model(entity.UserSubscriptions{})
	for condition, value := range where {
		db = db.Where(condition, value)
	}
	return db.Updates(updateData).Error
}

func (r *UserSubscriptionsImpl) Reload(ctx context.Context, tx *gorm.DB, id uint) (*entity.UserSubscriptions, error) {
	var row entity.UserSubscriptions
	err := repository.GormDB(ctx, r.db, tx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
