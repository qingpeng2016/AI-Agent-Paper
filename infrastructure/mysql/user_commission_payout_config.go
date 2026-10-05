package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserCommissionPayoutConfigImpl struct {
	db *gorm.DB
}

func NewUserCommissionPayoutConfigImpl(db *gorm.DB) repository.UserCommissionPayoutConfigRepo {
	return &UserCommissionPayoutConfigImpl{db: db}
}

func (r *UserCommissionPayoutConfigImpl) ListByUserID(ctx context.Context, userID uint) ([]entity.UserCommissionPayoutConfig, error) {
	var rows []entity.UserCommissionPayoutConfig
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error
	return rows, err
}

func (r *UserCommissionPayoutConfigImpl) FindByUserIDAndChannel(ctx context.Context, userID uint, channel string) (*entity.UserCommissionPayoutConfig, error) {
	var row entity.UserCommissionPayoutConfig
	err := r.db.WithContext(ctx).Where("user_id = ? AND channel = ?", userID, channel).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *UserCommissionPayoutConfigImpl) Save(ctx context.Context, row *entity.UserCommissionPayoutConfig) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "channel"}},
		DoUpdates: clause.AssignmentColumns([]string{"qr_mime", "qr_image", "updated_at"}),
	}).Create(row).Error
}
