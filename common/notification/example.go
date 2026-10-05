package notification

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/common/dederi/logger"

	"go.uber.org/zap"
)

func ExampleUsage() {
	notificationManager := NewNotificationManager()
	ctx := context.Background()

	if err := notificationManager.SendErrorAlertWithOrder(ctx, "示例错误", []FieldPair{
		{Key: "模块", Value: "example"},
	}); err != nil {
		logger.ErrorZ(ctx, "Failed to send error alert", zap.Error(err))
	}

	if err := notificationManager.SendInfoAlert(ctx, "服务启动成功", []FieldPair{
		{Key: "服务名称", Value: "mojo-strategy-liquidation"},
	}); err != nil {
		logger.ErrorZ(ctx, "Failed to send info alert", zap.Error(err))
	}
}
