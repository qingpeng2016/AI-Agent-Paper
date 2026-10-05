package bot

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"github.com/qingpeng2016/ai-token-mall/common/notification"
	"github.com/qingpeng2016/ai-token-mall/conf"
	"go.uber.org/zap"
)

// Entry Bot 入口（定时驱动 Scheduler.Handle）
type Entry struct {
	scheduler *Scheduler
	ns        *gocron.Scheduler
	startMu   sync.Mutex
}

func catch() {
	if e := recover(); e != nil {
		notification.SendErrorLog(context.Background(), "bot panic", zap.Any("e", e), zap.Any("stack", debug.Stack()))
	}
}

// NewEntry 创建 Bot 入口
func NewEntry(scheduler *Scheduler) *Entry {
	return &Entry{scheduler: scheduler}
}

// Start 启动 Bot
func (e *Entry) Start() error {
	_ = notification.GetGlobalNotificationManager().SendInfoAlert(context.Background(), "AI-Token-Mall bot 启动", []notification.FieldPair{
		{Key: "机器", Value: conf.Get_MACHINE_NAME()},
	})

	e.startMu.Lock()
	defer e.startMu.Unlock()
	if e.ns != nil {
		return nil
	}

	e.ns = gocron.NewScheduler(time.Local)
	_, err := e.ns.Every(4).Seconds().Name("Scheduler").WaitForSchedule().SingletonMode().Do(func() {
		defer catch()
		e.scheduler.Handle()
	})
	if err != nil {
		notification.SendErrorLog(context.Background(), "bot Scheduler-job-create-failed", zap.Error(err))
		return err
	}
	e.ns.SetMaxConcurrentJobs(1, gocron.WaitMode)
	e.ns.StartAsync()

	logger.InfoZ(context.Background(), "bot Start", zap.String("message", "AI-Token-Mall bot 已启动"))
	return nil
}

func (e *Entry) Stop() error {
	e.startMu.Lock()
	defer e.startMu.Unlock()
	if e.ns != nil {
		e.ns.Stop()
		e.ns = nil
	}
	_ = e.scheduler.Stop()
	logger.InfoZ(context.Background(), "bot-stopped")
	return nil
}
