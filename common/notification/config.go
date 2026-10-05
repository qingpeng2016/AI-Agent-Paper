package notification

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"github.com/qingpeng2016/ai-token-mall/conf"
	"go.uber.org/zap"
)

// NotificationManager 通知管理器（三路：警告 / 通知 / 错误）
type NotificationManager struct {
	alertClient  *LarkClient // alert_webhook_url：警告
	notifyClient *LarkClient // notify_webhook_url：业务通知
	errorClient  *LarkClient // error_webhook_url：错误
}

// NewNotificationManager 创建通知管理器
func NewNotificationManager() *NotificationManager {
	larkConf := conf.GetLarkNotificationConf()
	if larkConf == nil {
		return &NotificationManager{}
	}

	var alertClient, notifyClient, errorClient *LarkClient
	if larkConf.AlertWebhookURL != "" {
		alertClient = NewLarkClient(larkConf.AlertWebhookURL)
	}
	if larkConf.NotifyWebhookURL != "" {
		notifyClient = NewLarkClient(larkConf.NotifyWebhookURL)
	}
	if larkConf.ErrorWebhookURL != "" {
		errorClient = NewLarkClient(larkConf.ErrorWebhookURL)
	}

	return &NotificationManager{
		alertClient:  alertClient,
		notifyClient: notifyClient,
		errorClient:  errorClient,
	}
}

// FieldPair 字段对
type FieldPair struct {
	Key   string
	Value string
}

// SendWarningAlert 发送警告通知（alert_webhook_url）
func (nm *NotificationManager) SendWarningAlert(ctx context.Context, warningMsg string, fields []FieldPair) error {
	if nm.alertClient == nil {
		logger.InfoZ(ctx, "alert-notification-disabled", zap.String("type", "warning"))
		return nil
	}
	larkConf := conf.GetLarkNotificationConf()
	if larkConf == nil {
		return nil
	}
	return nm.alertClient.SendWarningAlertWithOrder(ctx, larkConf.ServiceName, warningMsg, fields)
}

// SendErrorAlertWithOrder 发送错误报警（保持字段顺序，error_webhook_url）
func (nm *NotificationManager) SendErrorAlertWithOrder(ctx context.Context, errorMsg string, fields []FieldPair) error {
	if nm.errorClient == nil {
		logger.InfoZ(ctx, "error-notification-disabled", zap.String("type", "error"))
		return nil
	}
	larkConf := conf.GetLarkNotificationConf()
	if larkConf == nil {
		return nil
	}
	return nm.errorClient.SendErrorAlertWithOrder(ctx, larkConf.ServiceName, errorMsg, fields)
}

// SendInfoAlert 发送信息通知（notify_webhook_url）
func (nm *NotificationManager) SendInfoAlert(ctx context.Context, infoMsg string, fields []FieldPair) error {
	if nm.notifyClient == nil {
		logger.InfoZ(ctx, "notify-notification-disabled", zap.String("type", "info"))
		return nil
	}
	larkConf := conf.GetLarkNotificationConf()
	if larkConf == nil {
		return nil
	}
	return nm.notifyClient.SendInfoAlert(ctx, larkConf.ServiceName, infoMsg, fields)
}

// IsAlertEnabled 警告群是否启用
func (nm *NotificationManager) IsAlertEnabled() bool {
	return nm.alertClient != nil
}

// IsNotifyEnabled 通知群是否启用
func (nm *NotificationManager) IsNotifyEnabled() bool {
	return nm.notifyClient != nil
}

// IsErrorEnabled 错误群是否启用
func (nm *NotificationManager) IsErrorEnabled() bool {
	return nm.errorClient != nil
}
