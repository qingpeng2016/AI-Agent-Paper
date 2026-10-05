package logger

import (
	"context"
	"fmt"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"runtime"
)

var logger *zap.Logger

const (
	CallerKey = "caller"
)

// NewLogger 创建分离的info和error日志
// serviceName 服务名称
// logDir 日志目录路径
// loglevel 设置日志级别
//
//	debug 可以打印出 info debug warn
//	info  级别可以打印 warn info
//	warn  只能打印 warn
//	debug->info->warn->error
func NewLogger(serviceName, logDir string, loglevel zapcore.Level) {
	// 创建日志目录
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create log directory: %v", err))
	}

	// Info日志配置
	infoHook := lumberjack.Logger{
		Filename:   logDir + "/info.log", // info日志文件路径
		MaxSize:    300,                  // 进行切割之前，日志文件的最大大小(MB为单位)，300MB
		MaxBackups: 400,                  // 保留300个备份
		MaxAge:     14,                   // 保留14天
		Compress:   false,                // 是否压缩，默认不压缩
	}

	// Error日志配置
	errorHook := lumberjack.Logger{
		Filename:   logDir + "/error.log", // error日志文件路径
		MaxSize:    300,                   // 进行切割之前，日志文件的最大大小(MB为单位)，300MB
		MaxBackups: 400,                   // 保留300个备份
		MaxAge:     14,                    // 保留14天
		Compress:   false,                 // 是否压缩，默认不压缩
	}

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601（进程 time.Local=北京时间）
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}

	// 创建info级别的core (info, debug, warn级别)
	infoCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout), // 同时输出到控制台
			zapcore.AddSync(&infoHook), // 输出到info.log
		),
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.InfoLevel && lvl < zapcore.ErrorLevel
		}),
	)

	// 创建error级别的core (error, panic, fatal级别)
	errorCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&errorHook), // 只输出到error.log，不输出到控制台
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		}),
	)

	// 先对每个 core 进行脱敏包装，再合并
	redactedInfoCore := NewRedactingCore(infoCore)
	redactedErrorCore := NewRedactingCore(errorCore)

	// 合并多个已脱敏的core
	core := zapcore.NewTee(redactedInfoCore, redactedErrorCore)

	// 开启开发模式，堆栈跟踪
	addCaller := zap.AddCaller()
	// 开启文件及行号
	development := zap.Development()
	// 设置初始化字段,如：添加一个服务器名称
	filed := zap.Fields(zap.String("service_name", serviceName))
	// 构造日志
	logger = zap.New(core, addCaller, development, filed)
	logger.Info("Logger init success")
}

func GetLogger() *zap.Logger {
	return logger
}

// ErrorZ error logger with zap new api, this high performance api, strong advise to use
func ErrorZ(ctx context.Context, msg string, fields ...zap.Field) {
	if logger == nil {
		return
	}
	fields = append(fields, zap.String(CallerKey, caller()), zap.String(trace.TraceID, trace.GetTraceIdByCtx(ctx)))
	logger.Error(msg, fields...)
}

// WarnZ warn logger with zap new api, this high performance api, strong advise to use
func WarnZ(ctx context.Context, msg string, fields ...zap.Field) {
	if logger == nil {
		return
	}
	fields = append(fields, zap.String(CallerKey, caller()), zap.String(trace.TraceID, trace.GetTraceIdByCtx(ctx)))
	logger.Warn(msg, fields...)
}

// InfoZ info logger with zap new api, this high performance api, strong advise to use
func InfoZ(ctx context.Context, msg string, fields ...zap.Field) {
	if logger == nil {
		return
	}
	fields = append(fields, zap.String(CallerKey, caller()), zap.String(trace.TraceID, trace.GetTraceIdByCtx(ctx)))
	logger.Info(msg, fields...)
}

// DebugZ debug logger with zap new api, this high performance api, strong advise to use
func DebugZ(ctx context.Context, msg string, fields ...zap.Field) {
	if logger == nil {
		return
	}
	fields = append(fields, zap.String(CallerKey, caller()), zap.String(trace.TraceID, trace.GetTraceIdByCtx(ctx)))
	logger.Debug(msg, fields...)
}

func PanicZ(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(CallerKey, caller()), zap.String(trace.TraceID, trace.GetTraceIdByCtx(ctx)))
	logger.Panic(msg, fields...)
}

func caller() string {
	pc, _, _, _ := runtime.Caller(2)
	f := runtime.FuncForPC(pc)
	file, line := f.FileLine(pc)
	name := f.Name()
	return fmt.Sprintf("%s:%d %s", file, line, name)
}
