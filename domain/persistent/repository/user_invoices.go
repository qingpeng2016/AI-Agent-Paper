package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

type UserInvoicesListRow struct {
	entity.UserInvoices
	OrderNo string `gorm:"column:order_no"`
}

type UserInvoicesRepo interface {
	Create(ctx context.Context, tx *gorm.DB, m *entity.UserInvoices) error
	ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]UserInvoicesListRow, error)
	CountByUserID(ctx context.Context, userID uint) (int64, error)
}
