package notification

import (
	"context"
	"encoding/json"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strconv"
	"sync"
)

var (
	globalNotificationManager *NotificationManager
	once                      sync.Once
)

// GetGlobalNotificationManager 获取全局通知管理器单例
func GetGlobalNotificationManager() *NotificationManager {
	once.Do(func() {
		globalNotificationManager = NewNotificationManager()
	})
	return globalNotificationManager
}

func SendErrorLog(ctx context.Context, msg string, fields ...zap.Field) {
	// 构建有序的报警字段列表
	var alertFieldList []FieldPair

	// 从 zap.Field 中提取信息，保持顺序
	for _, field := range fields {
		var value string

		switch field.Key {
		case "error":
			if err, ok := field.Interface.(error); ok {
				alertFieldList = append(alertFieldList, FieldPair{"错误详情", err.Error()})
			}
		default:
			// 简单处理所有其他字段
			switch field.Type {
			case zapcore.StringType:
				value = field.String
			case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type:
				value = strconv.FormatInt(field.Integer, 10)
			case zapcore.Uint64Type, zapcore.Uint32Type, zapcore.Uint16Type, zapcore.Uint8Type:
				value = strconv.FormatUint(uint64(field.Integer), 10)
			case zapcore.ReflectType:
				// 处理 zap.Any 类型
				if field.Interface != nil {
					if jsonBytes, err := json.Marshal(field.Interface); err == nil {
						value = string(jsonBytes)
					} else {
						value = field.String
						if value == "" {
							value = "序列化失败"
						}
					}
				} else {
					value = "null"
				}
			default:
				// 特殊处理 type: 1 (可能是数组类型)
				if field.Type == 1 {
					// 处理数组类型
					if field.Interface != nil {
						if jsonBytes, err := json.Marshal(field.Interface); err == nil {
							value = string(jsonBytes)
						} else {
							value = field.String
							if value == "" {
								value = "序列化失败"
							}
						}
					} else {
						value = "null"
					}
				} else {
					// 其他类型直接用 String 方法
					value = field.String
				}
			}

			if value != "" {
				alertFieldList = append(alertFieldList, FieldPair{field.Key, value})
			}
		}
	}

	// 发送错误报警（直接传递有序字段列表）
	nm := GetGlobalNotificationManager()
	nm.SendErrorAlertWithOrder(ctx, msg, alertFieldList)

	// 然后输出日志
	logger.ErrorZ(ctx, msg, fields...)
}
