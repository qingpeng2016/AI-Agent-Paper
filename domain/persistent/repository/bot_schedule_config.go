package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

// BotScheduleConfigRepo 脚本调度配置表仓库
type BotScheduleConfigRepo interface {
	FindOne(ctx context.Context, tx *gorm.DB, where map[string]interface{}, orderBy string) (entity.BotScheduleConfig, error)
	List(ctx context.Context, tx *gorm.DB, where map[string]interface{}, orderBy, groupBy string) (total int64, list []entity.BotScheduleConfig, err error)
	Update(ctx context.Context, tx *gorm.DB, where map[string]interface{}, updateData map[string]interface{}) error
	Create(ctx context.Context, tx *gorm.DB, m *entity.BotScheduleConfig) (id uint, err error)
}
