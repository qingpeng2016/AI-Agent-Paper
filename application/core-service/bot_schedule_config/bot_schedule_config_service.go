package botscheduleconfig

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

// BotScheduleConfigService 脚本调度配置
type BotScheduleConfigService struct {
	repo repository.BotScheduleConfigRepo
	db   *gorm.DB
}

func NewBotScheduleConfigService(repo repository.BotScheduleConfigRepo, db *gorm.DB) *BotScheduleConfigService {
	return &BotScheduleConfigService{repo: repo, db: db}
}

func (s *BotScheduleConfigService) GetAllConfigs(ctx context.Context, where map[string]interface{}) (int64, []entity.BotScheduleConfig, error) {
	return s.repo.List(ctx, s.db, where, "", "")
}

func (s *BotScheduleConfigService) GetConfigByModuleAndTask(ctx context.Context, module, taskName string) (*entity.BotScheduleConfig, error) {
	c, err := s.repo.FindOne(ctx, s.db, map[string]interface{}{
		"module = ?":    module,
		"task_name = ?": taskName,
	}, "")
	if err != nil {
		return nil, err
	}
	return &c, nil
}
