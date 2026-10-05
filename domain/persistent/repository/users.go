package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type UserVipConfigRow struct {
	ID          uint
	VipConfigID uint
}

type UsersRepo interface {
	Create(ctx context.Context, tx *gorm.DB, u *entity.Users) error
	// CreateRegister withInvite=false 时不写入邀请/VIP 列，兼容未跑 invite 迁移的 users 表。
	CreateRegister(ctx context.Context, tx *gorm.DB, u *entity.Users, withInviteFields bool) error
	FindByEmail(ctx context.Context, email string) (*entity.Users, error)
	FindByPhone(ctx context.Context, phone string) (*entity.Users, error)
	FindByID(ctx context.Context, id uint) (*entity.Users, error)
	FindByVipDomain(ctx context.Context, vipDomain string) (*entity.Users, error)
	ListByParentUserID(ctx context.Context, parentUserID uint, offset, limit int) ([]entity.Users, error)
	CountByParentUserID(ctx context.Context, parentUserID uint) (int64, error)
	// CountInviteesGroupByInviter 统计每位上级名下的直接下级人数（parent_user_id>0）。
	CountInviteesGroupByInviter(ctx context.Context) (map[uint]int64, error)
	// SumInviteeCompletedOrderAmountByInviter 直属下级已完成订单实付累计（元）。
	SumInviteeCompletedOrderAmountByInviter(ctx context.Context, inviterUserID uint) (decimal.Decimal, error)
	// SumInviteeCompletedOrderAmountGroupByInviter 按上级汇总直属下级已完成订单实付累计。
	SumInviteeCompletedOrderAmountGroupByInviter(ctx context.Context) (map[uint]decimal.Decimal, error)
	ListIDAndVipConfigID(ctx context.Context, offset, limit int) ([]UserVipConfigRow, error)
	UpdateVipConfigID(ctx context.Context, userID, vipConfigID uint) error
	FindByIDForUpdate(ctx context.Context, tx *gorm.DB, id uint) (*entity.Users, error)
	ApplyWalletDelta(ctx context.Context, tx *gorm.DB, userID uint, delta decimal.Decimal) (balanceAfter decimal.Decimal, err error)
	// TransferCommissionToWallet 扣减 commission_balance、增加 wallet_balance（同一行锁内完成）。
	TransferCommissionToWallet(ctx context.Context, tx *gorm.DB, userID uint, amount decimal.Decimal) (commissionAfter, walletAfter decimal.Decimal, err error)
	ApplyCommissionDelta(ctx context.Context, tx *gorm.DB, userID uint, delta decimal.Decimal) (balanceAfter decimal.Decimal, err error)
	UpdateLastLogin(ctx context.Context, id uint) error
	UpdatePassword(ctx context.Context, id uint, passwordHash, passwordPlain string) error
	Count(ctx context.Context) (int64, error)
}

type StatsRepo interface {
	CountUsers(ctx context.Context) (int64, error)
	CountAccessLogs(ctx context.Context) (int64, error)
}
