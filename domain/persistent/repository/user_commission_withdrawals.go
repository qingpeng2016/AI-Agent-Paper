package repository

import (
	"context"
	"time"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

type UserCommissionWithdrawalsRepo interface {
	Create(ctx context.Context, tx *gorm.DB, row *entity.UserCommissionWithdrawals) error
	ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]entity.UserCommissionWithdrawals, error)
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	UpdateStatus(ctx context.Context, tx *gorm.DB, id uint, status string, failReason *string, processedAt *time.Time) error
}
