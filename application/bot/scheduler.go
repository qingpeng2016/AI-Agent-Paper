package bot

import (
	"context"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	couponexpire "github.com/qingpeng2016/ai-agent-paper/application/bot/scripts/coupon_expire"
	experimentplan "github.com/qingpeng2016/ai-agent-paper/application/bot/scripts/experiment_plan"
	topicaudit "github.com/qingpeng2016/ai-agent-paper/application/bot/scripts/topic_audit"
	topicgenerateideas "github.com/qingpeng2016/ai-agent-paper/application/bot/scripts/topic_generate_ideas"
	topicliteraturedownload "github.com/qingpeng2016/ai-agent-paper/application/bot/scripts/topic_literature_download"
	viplevelsync "github.com/qingpeng2016/ai-agent-paper/application/bot/scripts/vip_level_sync"
	botscheduleconfig "github.com/qingpeng2016/ai-agent-paper/application/core-service/bot_schedule_config"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"github.com/qingpeng2016/ai-agent-paper/common/notification"
	"github.com/qingpeng2016/ai-agent-paper/conf"
	entity2 "github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"

	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
)

// Scheduler Bot 调度器
type Scheduler struct {
	botScheduleConfigService *botscheduleconfig.BotScheduleConfigService
	vipLevelSyncJob              *viplevelsync.VipLevelSyncJob
	couponExpireJob              *couponexpire.CouponExpireJob
	topicLiteratureDownloadJob   *topicliteraturedownload.TopicLiteratureDownloadJob
	topicGenerateIdeasJob        *topicgenerateideas.TopicGenerateIdeasJob
	topicAuditJob                *topicaudit.TopicAuditJob
	experimentPlanJob            *experimentplan.ExperimentPlanJob
	ns                       *gocron.Scheduler
	jobMap                   map[string]*gocron.Job
	configMap                map[string]float64
	started                  bool
	mu                       sync.Mutex
}

// NewScheduler 创建调度器
func NewScheduler(
	botScheduleConfigService *botscheduleconfig.BotScheduleConfigService,
	vipLevelSyncJob *viplevelsync.VipLevelSyncJob,
	couponExpireJob *couponexpire.CouponExpireJob,
	topicLiteratureDownloadJob *topicliteraturedownload.TopicLiteratureDownloadJob,
	topicGenerateIdeasJob *topicgenerateideas.TopicGenerateIdeasJob,
	topicAuditJob *topicaudit.TopicAuditJob,
	experimentPlanJob *experimentplan.ExperimentPlanJob,
) *Scheduler {
	return &Scheduler{
		botScheduleConfigService:   botScheduleConfigService,
		vipLevelSyncJob:            vipLevelSyncJob,
		couponExpireJob:            couponExpireJob,
		topicLiteratureDownloadJob: topicLiteratureDownloadJob,
		topicGenerateIdeasJob:      topicGenerateIdeasJob,
		topicAuditJob:              topicAuditJob,
		experimentPlanJob:          experimentPlanJob,
		ns:                       gocron.NewScheduler(time.Local),
		jobMap:                   make(map[string]*gocron.Job),
		configMap:                make(map[string]float64),
		started:                  false,
	}
}

// Handle 读取配置并动态更新调度任务（非阻塞，可重复调用）
// configIDs: 如果提供且不为空，则只查询指定ID的配置；否则查询所有启用的配置
func (s *Scheduler) Handle(configIDs ...uint) {
	ctx := context.Background()
	s.mu.Lock()
	machineName := conf.Get_MACHINE_NAME()
	logger.DebugZ(ctx, "bot scheduler-LockAcquired", zap.String("machineName", machineName))
	defer s.mu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			notification.SendErrorLog(ctx, "bot scheduler-Panic",
				zap.Any("panic", r),
				zap.String("stack", string(debug.Stack())))
		}
	}()

	if !s.started {
		s.ns.StartAsync()
		s.started = true
		logger.InfoZ(ctx, "bot scheduler-Started")
	}

	// 获取所有启用的调度配置
	where := map[string]interface{}{}
	hasConfigIDs := len(configIDs) > 0
	if hasConfigIDs {
		where["id IN ?"] = configIDs
	} else {
		where["module IN ?"] = []string{couponexpire.ModuleAIAgentPaper}
		where["is_enabled = ?"] = 1
		where["is_strategy_enabled = ?"] = 1
	}
	total, botConfigs, err := s.botScheduleConfigService.GetAllConfigs(ctx, where)
	if err != nil {
		notification.SendErrorLog(ctx, "bot scheduler-FailedToGetConfigs", zap.Error(err))
		return
	}
	configMap := make(map[string]entity2.BotScheduleConfig)
	for _, config := range botConfigs {
		jobName := config.Module + "-" + config.TaskName
		configMap[jobName] = config
	}

	// 移除已禁用的任务
	for jobName := range s.jobMap {
		if _, exists := configMap[jobName]; !exists {
			s.ns.RemoveByTags(jobName)
			delete(s.jobMap, jobName)
			delete(s.configMap, jobName)
			logger.InfoZ(ctx, "bot scheduler-JobRemoved", zap.String("jobName", jobName))
		}
	}
	if total == 0 || len(botConfigs) == 0 {
		logger.InfoZ(ctx, "bot scheduler-NoEnabledConfigs")
		return
	}

	// 为每个配置创建或更新调度任务
	for _, config := range botConfigs {
		handleFunc := s.getHandleFunc(config.Module, config.TaskName)
		if handleFunc == nil {
			notification.SendErrorLog(ctx, "bot scheduler-UnknownTask",
				zap.String("module", config.Module),
				zap.String("taskName", config.TaskName))
			continue
		}

		jobName := config.Module + "-" + config.TaskName
		module := config.Module
		taskName := config.TaskName
		needUpdate := false
		if existingJob, exists := s.jobMap[jobName]; exists {
			if savedInterval, ok := s.configMap[jobName]; ok {
				if savedInterval != config.IntervalSeconds {
					needUpdate = true
				}
			} else {
				needUpdate = true
			}
			_ = existingJob
		}

		if !needUpdate && s.jobMap[jobName] != nil {
			continue
		}

		if needUpdate {
			s.ns.RemoveByTags(jobName)
			delete(s.jobMap, jobName)
			delete(s.configMap, jobName)
		}

		intervalDur := time.Duration(config.IntervalSeconds * float64(time.Second))
		if intervalDur < time.Millisecond {
			intervalDur = time.Millisecond
		}
		job, err := s.ns.Every(intervalDur).
			Tag(jobName).
			Name(jobName).
			WaitForSchedule().
			SingletonMode().
			Do(func() {
				defer catch()
				ctx = context.Background()
				if !hasConfigIDs {
					dbConfig, err := s.botScheduleConfigService.GetConfigByModuleAndTask(ctx, module, taskName)
					if err != nil {
						notification.SendErrorLog(ctx, "bot scheduler-FailedToGetConfig",
							zap.String("module", module),
							zap.String("taskName", taskName),
							zap.Error(err))
						return
					}
					if dbConfig == nil || dbConfig.IsEnabled != 1 {
						logger.InfoZ(ctx, "bot scheduler-TaskDisabled",
							zap.String("module", module),
							zap.String("taskName", taskName))
						return
					}
				}
				handleFunc()
			})
		if err != nil {
			notification.SendErrorLog(ctx, "bot scheduler-FailedToCreateJob",
				zap.String("module", config.Module),
				zap.String("taskName", config.TaskName),
				zap.Error(err))
			continue
		}

		s.jobMap[jobName] = job
		s.configMap[jobName] = config.IntervalSeconds
		if needUpdate {
			logger.InfoZ(ctx, "bot scheduler-JobUpdated",
				zap.String("module", config.Module),
				zap.String("taskName", config.TaskName),
				zap.Float64("interval", config.IntervalSeconds))
		} else {
			logger.InfoZ(ctx, "bot scheduler-JobCreated",
				zap.String("module", config.Module),
				zap.String("taskName", config.TaskName),
				zap.Float64("interval", config.IntervalSeconds))
		}
	}

	jitterMs := rand.Intn(1000)
	time.Sleep(time.Duration(jitterMs) * time.Millisecond)
}

// Stop 停止调度器
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for jobName := range s.jobMap {
		_ = s.ns.RemoveByTags(jobName)
	}
	s.jobMap = make(map[string]*gocron.Job)
	s.configMap = make(map[string]float64)
	s.started = false
	s.ns.Stop()
	return nil
}

func (s *Scheduler) getHandleFunc(module, taskName string) func() {
	switch module {
	case couponexpire.ModuleAIAgentPaper:
		return s.getAIAgentPaperHandleFunc(taskName)
	}
	return nil
}

func (s *Scheduler) getAIAgentPaperHandleFunc(taskName string) func() {
	switch taskName {
	case viplevelsync.TaskVipLevelSync:
		return func() {
			s.vipLevelSyncJob.Run(context.Background())
		}
	case couponexpire.TaskCouponExpire:
		return func() {
			s.couponExpireJob.Run(context.Background())
		}
	case topicliteraturedownload.TaskTopicLiteratureDownload:
		return func() {
			s.topicLiteratureDownloadJob.Run(context.Background())
		}
	case topicgenerateideas.TaskTopicGenerateIdeasLLM:
		return func() {
			s.topicGenerateIdeasJob.Run(context.Background())
		}
	case topicaudit.TaskTopicAuditLLM:
		return func() {
			s.topicAuditJob.Run(context.Background())
		}
	case experimentplan.TaskExperimentPlanLLM:
		return func() {
			s.experimentPlanJob.Run(context.Background())
		}
	default:
		return nil
	}
}
