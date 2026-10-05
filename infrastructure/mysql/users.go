package mysql

import (
	"context"
	"errors"
	"time"

	"fmt"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UsersImpl struct {
	db *gorm.DB
}

func NewUsersImpl(db *gorm.DB) repository.UsersRepo {
	return &UsersImpl{db: db}
}

func (r *UsersImpl) Create(ctx context.Context, tx *gorm.DB, u *entity.Users) error {
	return r.CreateRegister(ctx, tx, u, true)
}

func (r *UsersImpl) CreateRegister(ctx context.Context, tx *gorm.DB, u *entity.Users, withInviteFields bool) error {
	if tx == nil {
		tx = r.db
	}
	q := tx.WithContext(ctx)
	if !withInviteFields {
		q = q.Omit("parent_user_id", "vip_config_id", "vip_domain")
	}
	return q.Create(u).Error
}

func (r *UsersImpl) FindByEmail(ctx context.Context, email string) (*entity.Users, error) {
	var u entity.Users
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UsersImpl) FindByPhone(ctx context.Context, phone string) (*entity.Users, error) {
	var u entity.Users
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UsersImpl) FindByID(ctx context.Context, id uint) (*entity.Users, error) {
	var u entity.Users
	err := r.db.WithContext(ctx).First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UsersImpl) FindByVipDomain(ctx context.Context, vipDomain string) (*entity.Users, error) {
	var u entity.Users
	err := r.db.WithContext(ctx).Where("vip_domain = ?", vipDomain).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UsersImpl) ListByParentUserID(ctx context.Context, parentUserID uint, offset, limit int) ([]entity.Users, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	var rows []entity.Users
	err := r.db.WithContext(ctx).
		Where("parent_user_id = ?", parentUserID).
		Order("id DESC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *UsersImpl) CountByParentUserID(ctx context.Context, parentUserID uint) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.Users{}).
		Where("parent_user_id = ?", parentUserID).Count(&n).Error
	return n, err
}

func (r *UsersImpl) CountInviteesGroupByInviter(ctx context.Context) (map[uint]int64, error) {
	type row struct {
		InviterID uint  `gorm:"column:inviter_id"`
		Cnt       int64 `gorm:"column:cnt"`
	}
	var rows []row
	err := r.db.WithContext(ctx).Model(&entity.Users{}).
		Select("parent_user_id AS inviter_id, COUNT(*) AS cnt").
		Where("parent_user_id > ?", 0).
		Group("parent_user_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[uint]int64, len(rows))
	for _, r := range rows {
		out[r.InviterID] = r.Cnt
	}
	return out, nil
}

func (r *UsersImpl) ListIDAndVipConfigID(ctx context.Context, offset, limit int) ([]repository.UserVipConfigRow, error) {
	if limit <= 0 || limit > 2000 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}
	var rows []repository.UserVipConfigRow
	err := r.db.WithContext(ctx).Model(&entity.Users{}).
		Select("id", "vip_config_id").
		Order("id ASC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *UsersImpl) SumInviteeCompletedOrderAmountByInviter(ctx context.Context, inviterUserID uint) (decimal.Decimal, error) {
	var total decimal.Decimal
	err := r.db.WithContext(ctx).Table("users AS invitee").
		Select("COALESCE(SUM(o.total_amount), 0)").
		Joins("INNER JOIN user_orders o ON o.user_id = invitee.id AND o.status = ?", "completed").
		Where("invitee.parent_user_id = ?", inviterUserID).
		Scan(&total).Error
	return total, err
}

func (r *UsersImpl) SumInviteeCompletedOrderAmountGroupByInviter(ctx context.Context) (map[uint]decimal.Decimal, error) {
	type row struct {
		InviterID uint            `gorm:"column:inviter_id"`
		Total     decimal.Decimal `gorm:"column:total"`
	}
	var rows []row
	err := r.db.WithContext(ctx).Table("users AS invitee").
		Select("invitee.parent_user_id AS inviter_id, COALESCE(SUM(o.total_amount), 0) AS total").
		Joins("INNER JOIN user_orders o ON o.user_id = invitee.id AND o.status = ?", "completed").
		Where("invitee.parent_user_id > ?", 0).
		Group("invitee.parent_user_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[uint]decimal.Decimal, len(rows))
	for _, r := range rows {
		out[r.InviterID] = r.Total
	}
	return out, nil
}

func (r *UsersImpl) UpdateVipConfigID(ctx context.Context, userID, vipConfigID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.Users{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"vip_config_id": vipConfigID,
		"updated_at":    now,
	}).Error
}

func (r *UsersImpl) FindByIDForUpdate(ctx context.Context, tx *gorm.DB, id uint) (*entity.Users, error) {
	var u entity.Users
	err := repository.GormDB(ctx, r.db, tx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UsersImpl) ApplyCommissionDelta(ctx context.Context, tx *gorm.DB, userID uint, delta decimal.Decimal) (decimal.Decimal, error) {
	u, err := r.FindByIDForUpdate(ctx, tx, userID)
	if err != nil {
		return decimal.Zero, err
	}
	if u == nil {
		return decimal.Zero, gorm.ErrRecordNotFound
	}
	newBal := u.CommissionBalance.Add(delta)
	if newBal.IsNegative() {
		return decimal.Zero, fmt.Errorf("insufficient commission balance")
	}
	now := time.Now()
	if err := repository.GormDB(ctx, r.db, tx).Model(&entity.Users{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"commission_balance": newBal,
		"updated_at":         now,
	}).Error; err != nil {
		return decimal.Zero, err
	}
	return newBal, nil
}

func (r *UsersImpl) TransferCommissionToWallet(
	ctx context.Context, tx *gorm.DB, userID uint, amount decimal.Decimal,
) (decimal.Decimal, decimal.Decimal, error) {
	u, err := r.FindByIDForUpdate(ctx, tx, userID)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	if u == nil {
		return decimal.Zero, decimal.Zero, gorm.ErrRecordNotFound
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, decimal.Zero, fmt.Errorf("invalid transfer amount")
	}
	newCommission := u.CommissionBalance.Sub(amount)
	if newCommission.IsNegative() {
		return decimal.Zero, decimal.Zero, fmt.Errorf("insufficient commission balance")
	}
	newWallet := u.WalletBalance.Add(amount)
	now := time.Now()
	if err := repository.GormDB(ctx, r.db, tx).Model(&entity.Users{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"commission_balance": newCommission,
		"wallet_balance":     newWallet,
		"updated_at":         now,
	}).Error; err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	return newCommission, newWallet, nil
}

func (r *UsersImpl) ApplyWalletDelta(ctx context.Context, tx *gorm.DB, userID uint, delta decimal.Decimal) (decimal.Decimal, error) {
	u, err := r.FindByIDForUpdate(ctx, tx, userID)
	if err != nil {
		return decimal.Zero, err
	}
	if u == nil {
		return decimal.Zero, gorm.ErrRecordNotFound
	}
	newBal := u.WalletBalance.Add(delta)
	if newBal.IsNegative() {
		return decimal.Zero, fmt.Errorf("insufficient wallet balance")
	}
	if err := repository.GormDB(ctx, r.db, tx).Model(&entity.Users{}).
		Where("id = ?", userID).
		Update("wallet_balance", newBal).Error; err != nil {
		return decimal.Zero, err
	}
	return newBal, nil
}

func (r *UsersImpl) UpdateLastLogin(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.Users{}).Where("id = ?", id).Update("last_login_at", now).Error
}

func (r *UsersImpl) UpdatePassword(ctx context.Context, id uint, passwordHash, passwordPlain string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.Users{}).Where("id = ?", id).Updates(map[string]interface{}{
		"password_hash":  passwordHash,
		"password_plain": passwordPlain,
		"updated_at":     now,
	}).Error
}

func (r *UsersImpl) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.Users{}).Count(&n).Error
	return n, err
}

type StatsImpl struct {
	db *gorm.DB
}

func NewStatsImpl(db *gorm.DB) repository.StatsRepo {
	return &StatsImpl{db: db}
}

func (r *StatsImpl) CountUsers(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.Users{}).Count(&n).Error
	return n, err
}

func (r *StatsImpl) CountAccessLogs(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&entity.UserAccessLogs{}).Count(&n).Error
	return n, err
}
