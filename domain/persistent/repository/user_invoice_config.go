package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
)

type UserInvoiceConfigRepo interface {
	ListByUserID(ctx context.Context, userID uint) ([]entity.UserInvoiceConfig, error)
	FindByIDForUser(ctx context.Context, userID, id uint) (*entity.UserInvoiceConfig, error)
	Create(ctx context.Context, m *entity.UserInvoiceConfig) error
	Update(ctx context.Context, id uint, fields map[string]interface{}) error
	ClearDefaultForUser(ctx context.Context, userID uint, exceptID uint) error
}
