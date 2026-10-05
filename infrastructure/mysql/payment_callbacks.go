package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type PaymentCallbacksImpl struct {
	db *gorm.DB
}

func NewPaymentCallbacksImpl(db *gorm.DB) repository.PaymentCallbacksRepo {
	return &PaymentCallbacksImpl{db: db}
}

func (r *PaymentCallbacksImpl) FindByIdempotency(ctx context.Context, tx *gorm.DB, channel, idempotencyKey string) (*entity.PaymentCallbacks, error) {
	var row entity.PaymentCallbacks
	err := repository.GormDB(ctx, r.db, tx).
		Where("channel = ? AND idempotency_key = ?", channel, idempotencyKey).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaymentCallbacksImpl) Create(ctx context.Context, tx *gorm.DB, m *entity.PaymentCallbacks) error {
	return repository.GormDB(ctx, r.db, tx).Create(m).Error
}
