package notification

import (
	"context"
	"encoding/json"
	"github.com/qingpeng2016/ai-agent-paper/common/dederi/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strconv"
	"sync"
)

var (
	globalNotificationManager *NotificationManager
	once                      sync.Once
)

type skipErrorAlertKey struct{}

// WithSkipErrorAlert 仅写日志、不推送 Lark（用于 notification 内部或告警链路自身失败，避免递归）。
func WithSkipErrorAlert(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipErrorAlertKey{}, true)
}

func skipErrorAlert(ctx context.Context) bool {
	v, _ := ctx.Value(skipErrorAlertKey{}).(bool)
	return v
}

// GetGlobalNotificationManager 获取全局通知管理器单例
func GetGlobalNotificationManager() *NotificationManager {
	once.Do(func() {
		globalNotificationManager = NewNotificationManager()
	})
	return globalNotificationManager
}

// SendNotifyLog 写 Info 日志并推送 Lark 业务通知（notify_webhook_url）。
func SendNotifyLog(ctx context.Context, msg string, fields ...zap.Field) {
	logger.InfoZ(ctx, msg, fields...)
	alertFieldList := zapFieldsToFieldPairs(fields...)
	nm := GetGlobalNotificationManager()
	_ = nm.SendInfoAlert(ctx, msg, alertFieldList)
}

func zapFieldsToFieldPairs(fields ...zap.Field) []FieldPair {
	var alertFieldList []FieldPair
	for _, field := range fields {
		switch field.Key {
		case "error":
			if err, ok := field.Interface.(error); ok {
				alertFieldList = append(alertFieldList, FieldPair{"错误详情", err.Error()})
			}
		default:
			value := zapFieldStringValue(field)
			if value != "" {
				alertFieldList = append(alertFieldList, FieldPair{field.Key, value})
			}
		}
	}
	return alertFieldList
}

func zapFieldStringValue(field zap.Field) string {
	switch field.Type {
	case zapcore.StringType:
		return field.String
	case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type:
		return strconv.FormatInt(field.Integer, 10)
	case zapcore.Uint64Type, zapcore.Uint32Type, zapcore.Uint16Type, zapcore.Uint8Type:
		return strconv.FormatUint(uint64(field.Integer), 10)
	case zapcore.ReflectType:
		if field.Interface != nil {
			if jsonBytes, err := json.Marshal(field.Interface); err == nil {
				return string(jsonBytes)
			}
		}
		return field.String
	default:
		if field.Type == 1 && field.Interface != nil {
			if jsonBytes, err := json.Marshal(field.Interface); err == nil {
				return string(jsonBytes)
			}
		}
		return field.String
	}
}

// SendErrorLog 写 Error 日志并推送 Lark（error_webhook_url，未配置则用 alert_webhook_url）。
func SendErrorLog(ctx context.Context, msg string, fields ...zap.Field) {
	if !skipErrorAlert(ctx) {
		sendErrorLogAlert(ctx, msg, fields...)
	}
	logger.ErrorZ(ctx, msg, fields...)
}

func sendErrorLogAlert(ctx context.Context, msg string, fields ...zap.Field) {
	nm := GetGlobalNotificationManager()
	nm.SendErrorAlertWithOrder(ctx, msg, zapFieldsToFieldPairs(fields...))
}
