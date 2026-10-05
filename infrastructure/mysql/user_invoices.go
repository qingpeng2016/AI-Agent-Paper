package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type UserInvoicesImpl struct {
	db *gorm.DB
}

func NewUserInvoicesImpl(db *gorm.DB) repository.UserInvoicesRepo {
	return &UserInvoicesImpl{db: db}
}

func (r *UserInvoicesImpl) Create(ctx context.Context, tx *gorm.DB, m *entity.UserInvoices) error {
	return repository.GormDB(ctx, r.db, tx).Create(m).Error
}

func (r *UserInvoicesImpl) ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]repository.UserInvoicesListRow, error) {
	if limit <= 0 || limit > 100 {
		limit = 9
	}
	if offset < 0 {
		offset = 0
	}
	var rows []repository.UserInvoicesListRow
	err := r.db.WithContext(ctx).
		Table("user_invoices AS i").
		Select("i.*, o.order_no AS order_no").
		Joins("LEFT JOIN user_orders o ON o.id = i.order_id").
		Where("i.user_id = ?", userID).
		Order("i.id DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *UserInvoicesImpl) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.UserInvoices{}).Where("user_id = ?", userID).Count(&n).Error
	return n, err
}
