package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type CouponCampaignsImpl struct {
	db *gorm.DB
}

func NewCouponCampaignsImpl(db *gorm.DB) repository.CouponCampaignsRepo {
	return &CouponCampaignsImpl{db: db}
}

func (r *CouponCampaignsImpl) ListAutoGrantOnRegister(ctx context.Context) ([]entity.CouponCampaigns, error) {
	var rows []entity.CouponCampaigns
	err := r.db.WithContext(ctx).
		Where("auto_grant_on_register = ? AND enabled = ?", true, true).
		Order("id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *CouponCampaignsImpl) FindByID(ctx context.Context, id uint) (*entity.CouponCampaigns, error) {
	var row entity.CouponCampaigns
	err := r.db.WithContext(ctx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
