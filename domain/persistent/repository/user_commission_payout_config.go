package repository

import (
	"context"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

type UserCommissionPayoutConfigRepo interface {
	ListByUserID(ctx context.Context, userID uint) ([]entity.UserCommissionPayoutConfig, error)
	FindByUserIDAndChannel(ctx context.Context, userID uint, channel string) (*entity.UserCommissionPayoutConfig, error)
	Save(ctx context.Context, row *entity.UserCommissionPayoutConfig) error
}
