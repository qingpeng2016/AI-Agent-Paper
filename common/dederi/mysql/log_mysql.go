package mysqlclient

import (
	"context"
	"fmt"
	logger2 "github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

type mysqlLogger struct {
}

// 实现 GORM 的 Log 方法
func (l *mysqlLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

// 实现 GORM 的 Info 方法
func (l *mysqlLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	logger2.InfoZ(ctx, msg, zap.Any("data", data))
}

// 实现 GORM 的 Warn 方法
func (l *mysqlLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	logger2.WarnZ(ctx, msg, zap.Any("data", data))
}

// 实现 GORM 的 Error 方法
func (l *mysqlLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	logger2.ErrorZ(ctx, msg, zap.Any("data", data))
}

// 实现 GORM 的 Trace 方法
func (l *mysqlLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err != nil {
		logger2.ErrorZ(ctx, fmt.Sprintf("[sql_file:%s][%.3fms] [rows:%v] %s, %s", utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql, err.Error()))
	} else {
		logger2.InfoZ(ctx, fmt.Sprintf("[sql_file:%s][%.3fms] [rows:%v] %s", utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql))
	}
	if float64(elapsed.Nanoseconds())/1e9 > 1 {
		logger2.ErrorZ(ctx, fmt.Sprintf("[sql_file:%s][%.3fms] [rows:%v] slow sql : %s", utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql))
	}
}
