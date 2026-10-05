package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type UserInvoiceConfigImpl struct {
	db *gorm.DB
}

func NewUserInvoiceConfigImpl(db *gorm.DB) repository.UserInvoiceConfigRepo {
	return &UserInvoiceConfigImpl{db: db}
}

func (r *UserInvoiceConfigImpl) ListByUserID(ctx context.Context, userID uint) ([]entity.UserInvoiceConfig, error) {
	var rows []entity.UserInvoiceConfig
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, id DESC").
		Find(&rows).Error
	return rows, err
}

func (r *UserInvoiceConfigImpl) FindByIDForUser(ctx context.Context, userID, id uint) (*entity.UserInvoiceConfig, error) {
	var row entity.UserInvoiceConfig
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *UserInvoiceConfigImpl) Create(ctx context.Context, m *entity.UserInvoiceConfig) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *UserInvoiceConfigImpl) Update(ctx context.Context, id uint, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&entity.UserInvoiceConfig{}).Where("id = ?", id).Updates(fields).Error
}

func (r *UserInvoiceConfigImpl) ClearDefaultForUser(ctx context.Context, userID uint, exceptID uint) error {
	q := r.db.WithContext(ctx).Model(&entity.UserInvoiceConfig{}).
		Where("user_id = ? AND is_default = 1", userID)
	if exceptID > 0 {
		q = q.Where("id <> ?", exceptID)
	}
	return q.Update("is_default", 0).Error
}
