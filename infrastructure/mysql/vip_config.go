package mysql

import (
	"context"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

var vipConfigBaseSelect = []string{
	"id", "level_label", "min_valid_invites", "min_invitee_paid_amount", "rate_percent",
	"sort_order", "enabled", "is_default", "created_at", "updated_at",
}

type VipConfigImpl struct {
	db *gorm.DB
}

func NewVipConfigImpl(db *gorm.DB) repository.VipConfigRepo {
	return &VipConfigImpl{db: db}
}

func (r *VipConfigImpl) FindByID(ctx context.Context, id uint) (*entity.VipConfig, error) {
	var row entity.VipConfig
	err := r.db.WithContext(ctx).Select(vipConfigBaseSelect).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *VipConfigImpl) FindDefault(ctx context.Context) (*entity.VipConfig, error) {
	var id uint
	err := r.db.WithContext(ctx).Table(entity.VipConfig{}.TableName()).
		Where("is_default = ?", true).Limit(1).Pluck("id", &id).Error
	if err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1054 {
			return r.FindByID(ctx, 1)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	if id == 0 {
		return r.FindByID(ctx, 1)
	}
	return r.FindByID(ctx, id)
}

func (r *VipConfigImpl) ListEnabled(ctx context.Context) ([]entity.VipConfig, error) {
	var rows []entity.VipConfig
	err := r.db.WithContext(ctx).
		Where("enabled = ?", true).
		Order("sort_order ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *VipConfigImpl) ListAll(ctx context.Context) ([]entity.VipConfig, error) {
	var rows []entity.VipConfig
	err := r.db.WithContext(ctx).Order("sort_order ASC, id ASC").Find(&rows).Error
	return rows, err
}
