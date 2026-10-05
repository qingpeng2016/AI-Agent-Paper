package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserOrdersImpl struct {
	db *gorm.DB
}

func NewUserOrdersImpl(db *gorm.DB) repository.UserOrdersRepo {
	return &UserOrdersImpl{db: db}
}

func (r *UserOrdersImpl) ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]repository.UserOrdersListRow, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	var rows []repository.UserOrdersListRow
	err := r.db.WithContext(ctx).
		Table("user_orders AS o").
		Select("o.*, p.card_title AS product_card_title, p.sku_product_name AS product_sku_name").
		Joins("LEFT JOIN products p ON p.id = o.product_id").
		Where("o.user_id = ?", userID).
		Order("o.id DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *UserOrdersImpl) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.UserOrders{}).Where("user_id = ?", userID).Count(&n).Error
	return n, err
}

func (r *UserOrdersImpl) CreateOrder(ctx context.Context, order *entity.UserOrders) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *UserOrdersImpl) CreateRenewalOrderReplacingPending(ctx context.Context, order *entity.UserOrders) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&entity.UserOrders{}).
			Where(
				"user_id = ? AND user_subscription_id = ? AND order_type = ? AND status = ?",
				order.UserID, order.UserSubscriptionID, "renewal", "pending_payment",
			).
			Updates(map[string]interface{}{
				"status":     "cancelled",
				"closed_at":  now,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}
		return tx.Create(order).Error
	})
}

func (r *UserOrdersImpl) CreateUpgradeOrderReplacingPending(ctx context.Context, order *entity.UserOrders) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&entity.UserOrders{}).
			Where(
				"user_id = ? AND user_subscription_id = ? AND order_type = ? AND status = ?",
				order.UserID, order.UserSubscriptionID, "upgrade", "pending_payment",
			).
			Updates(map[string]interface{}{
				"status":     "cancelled",
				"closed_at":  now,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}
		return tx.Create(order).Error
	})
}

func (r *UserOrdersImpl) CreateQuotaAddonOrderReplacingPending(ctx context.Context, order *entity.UserOrders) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&entity.UserOrders{}).
			Where(
				"user_id = ? AND user_subscription_id = ? AND order_type = ? AND status = ?",
				order.UserID, order.UserSubscriptionID, "quota_addon", "pending_payment",
			).
			Updates(map[string]interface{}{
				"status":     "cancelled",
				"closed_at":  now,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}
		return tx.Create(order).Error
	})
}

func (r *UserOrdersImpl) FindOrderByOutTradeNo(ctx context.Context, outTradeNo string) (*entity.UserOrders, error) {
	var row entity.UserOrders
	err := r.db.WithContext(ctx).Where("out_trade_no = ?", outTradeNo).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *UserOrdersImpl) FindOrderByID(ctx context.Context, id uint) (*entity.UserOrders, error) {
	var row entity.UserOrders
	err := r.db.WithContext(ctx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *UserOrdersImpl) FindActiveSubscriptionByUserProduct(ctx context.Context, userID, productID uint) (*entity.UserSubscriptions, error) {
	var row entity.UserSubscriptions
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ? AND status = ?", userID, productID, "active").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *UserOrdersImpl) FindLatestSubscriptionByUserProduct(ctx context.Context, userID, productID uint) (*entity.UserSubscriptions, error) {
	var row entity.UserSubscriptions
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ? AND status IN ?", userID, productID, []string{"active", "expired"}).
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

func (r *UserOrdersImpl) FindSubscriptionForUser(ctx context.Context, userID, subscriptionID uint) (*entity.UserSubscriptions, error) {
	if subscriptionID == 0 {
		return nil, nil
	}
	var row entity.UserSubscriptions
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", subscriptionID, userID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *UserOrdersImpl) CreateTx(ctx context.Context, tx *gorm.DB, order *entity.UserOrders) error {
	return repository.GormDB(ctx, r.db, tx).Create(order).Error
}

func (r *UserOrdersImpl) FindByOutTradeNoForUpdate(ctx context.Context, tx *gorm.DB, outTradeNo string) (*entity.UserOrders, error) {
	var row entity.UserOrders
	err := repository.GormDB(ctx, r.db, tx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("out_trade_no = ?", outTradeNo).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *UserOrdersImpl) UpdateFields(ctx context.Context, tx *gorm.DB, id uint, fields map[string]interface{}) error {
	return repository.GormDB(ctx, r.db, tx).Model(&entity.UserOrders{}).Where("id = ?", id).Updates(fields).Error
}
