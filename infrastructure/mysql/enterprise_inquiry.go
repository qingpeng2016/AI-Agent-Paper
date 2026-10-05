package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type EnterpriseInquiryImpl struct {
	db *gorm.DB
}

func NewEnterpriseInquiryImpl(db *gorm.DB) repository.EnterpriseInquiryRepo {
	return &EnterpriseInquiryImpl{db: db}
}

func (r *EnterpriseInquiryImpl) Create(ctx context.Context, row *entity.EnterpriseInquiry) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *EnterpriseInquiryImpl) FindLatestByOwnerUserID(ctx context.Context, ownerUserID uint) (*entity.EnterpriseInquiry, error) {
	var row entity.EnterpriseInquiry
	err := r.db.WithContext(ctx).
		Where("owner_user_id = ?", ownerUserID).
		Order("id DESC").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
