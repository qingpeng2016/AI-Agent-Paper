package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

type PaymentCallbacksRepo interface {
	FindByIdempotency(ctx context.Context, tx *gorm.DB, channel, idempotencyKey string) (*entity.PaymentCallbacks, error)
	Create(ctx context.Context, tx *gorm.DB, m *entity.PaymentCallbacks) error
}
