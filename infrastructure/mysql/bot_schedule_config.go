package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type BotScheduleConfigImpl struct {
	db *gorm.DB
}

func NewBotScheduleConfigImpl(db *gorm.DB) repository.BotScheduleConfigRepo {
	return &BotScheduleConfigImpl{db: db}
}

func (r *BotScheduleConfigImpl) FindOne(ctx context.Context, tx *gorm.DB, where map[string]interface{}, orderBy string) (entity.BotScheduleConfig, error) {
	if tx == nil {
		tx = r.db
	}
	db := tx.WithContext(ctx).Model(entity.BotScheduleConfig{})
	for condition, value := range where {
		db = db.Where(condition, value)
	}
	if orderBy != "" {
		db = db.Order(orderBy)
	}
	var m entity.BotScheduleConfig
	err := db.First(&m).Error
	return m, err
}

func (r *BotScheduleConfigImpl) List(ctx context.Context, tx *gorm.DB, where map[string]interface{}, orderBy, groupBy string) (total int64, list []entity.BotScheduleConfig, err error) {
	if tx == nil {
		tx = r.db
	}
	db := tx.WithContext(ctx).Model(entity.BotScheduleConfig{})
	for condition, value := range where {
		db = db.Where(condition, value)
	}
	if orderBy != "" {
		db = db.Order(orderBy)
	}
	if groupBy != "" {
		db = db.Group(groupBy)
	}
	err = db.Count(&total).Find(&list).Error
	return total, list, err
}

func (r *BotScheduleConfigImpl) Update(ctx context.Context, tx *gorm.DB, where map[string]interface{}, updateData map[string]interface{}) error {
	if tx == nil {
		tx = r.db
	}
	db := tx.WithContext(ctx).Model(entity.BotScheduleConfig{})
	for condition, value := range where {
		db = db.Where(condition, value)
	}
	return db.Updates(updateData).Error
}

func (r *BotScheduleConfigImpl) Create(ctx context.Context, tx *gorm.DB, m *entity.BotScheduleConfig) (uint, error) {
	if tx == nil {
		tx = r.db
	}
	if err := tx.WithContext(ctx).Model(entity.BotScheduleConfig{}).Create(m).Error; err != nil {
		return 0, err
	}
	return m.ID, nil
}
