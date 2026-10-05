package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type VipDomainConfigImpl struct {
	db *gorm.DB
}

func NewVipDomainConfigImpl(db *gorm.DB) repository.VipDomainConfigRepo {
	return &VipDomainConfigImpl{db: db}
}

func (r *VipDomainConfigImpl) FindByID(ctx context.Context, id uint) (*entity.VipDomainConfig, error) {
	var row entity.VipDomainConfig
	err := r.db.WithContext(ctx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *VipDomainConfigImpl) FindByDomain(ctx context.Context, domain string) (*entity.VipDomainConfig, error) {
	var row entity.VipDomainConfig
	err := r.db.WithContext(ctx).Where("domain = ?", domain).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *VipDomainConfigImpl) ListAll(ctx context.Context) ([]entity.VipDomainConfig, error) {
	var rows []entity.VipDomainConfig
	err := r.db.WithContext(ctx).
		Order("id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *VipDomainConfigImpl) ListNonOfficial(ctx context.Context) ([]entity.VipDomainConfig, error) {
	var rows []entity.VipDomainConfig
	err := r.db.WithContext(ctx).
		Where("is_official = ?", false).
		Order("id ASC").
		Find(&rows).Error
	return rows, err
}
