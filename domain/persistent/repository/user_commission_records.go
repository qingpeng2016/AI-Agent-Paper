package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

type UserCommissionRecordsRepo interface {
	Create(ctx context.Context, tx *gorm.DB, row *entity.UserCommissionRecords) error
	FindByOrderID(ctx context.Context, tx *gorm.DB, orderID uint) (*entity.UserCommissionRecords, error)
	ListByInviterUserID(ctx context.Context, inviterUserID uint, offset, limit int) ([]entity.UserCommissionRecords, error)
	CountByInviterUserID(ctx context.Context, inviterUserID uint) (int64, error)
}
