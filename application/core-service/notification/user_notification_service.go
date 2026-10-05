package notification

import (
	"context"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
)

const defaultNotificationPageSize = 20

type templateView struct {
	Category string
	Title    string
	Body     string
	LinkTab  string
}

var notificationTemplates = map[string]templateView{
	"subscription_activated": {
		Category: "subscription",
		Title:    "套餐已开通",
		Body:     "您的套餐已激活，可在「我的套餐」查看详情，在「API Key」复制主 Key。",
		LinkTab:  "api-keys",
	},
	"subscription_renewed": {
		Category: "subscription",
		Title:    "套餐续费成功",
		Body:     "套餐已成功续费，有效期已延长，请至「我的套餐」确认。",
		LinkTab:  "plans",
	},
	"subscription_upgraded": {
		Category: "subscription",
		Title:    "套餐已升档",
		Body:     "升档已完成，额度与权益已更新，请在「我的套餐」查看。",
		LinkTab:  "plans",
	},
	"subscription_quota_added": {
		Category: "subscription",
		Title:    "额度已补充",
		Body:     "已为当前套餐增加额度，可在「我的套餐」查看用量。",
		LinkTab:  "plans",
	},
	"subscription_expired": {
		Category: "subscription",
		Title:    "套餐已过期",
		Body:     "您的套餐已到期，相关 Key 可能已停用，请尽快续费。",
		LinkTab:  "plans",
	},
	"renew_reminder": {
		Category: "subscription",
		Title:    "续费提醒",
		Body:     "套餐即将到期，建议提前续费以免影响调用。",
		LinkTab:  "plans",
	},
	"key_issued": {
		Category: "subscription",
		Title:    "API Key 已下发",
		Body:     "新的 API Key 已生成，完整 Key 仅在开通/下发时展示一次，请妥善保存。",
		LinkTab:  "api-keys",
	},
	"commission_transferred_to_wallet": {
		Category: "finance",
		Title:    "佣金已划转到余额",
		Body:     "您已将佣金划转到账户余额，可在「账户余额」查看资金流水。",
		LinkTab:  "account",
	},
	"commission_withdraw_submitted": {
		Category: "finance",
		Title:    "提现申请已提交",
		Body:     "佣金提现申请已受理，打款后将更新状态，可在「邀请返利 → 提现记录」查看。",
		LinkTab:  "sub-accounts",
	},
	"commission_rebate_earned": {
		Category: "finance",
		Title:    "邀请返利到账",
		Body:     "下级订单支付成功，返利已计入您的佣金，可在「邀请返利 → 返利记录」查看。",
		LinkTab:  "sub-accounts",
	},
	"vip_level_upgraded": {
		Category: "finance",
		Title:    "推广等级已提升",
		Body:     "您的邀请推广 VIP 等级已升级，返佣比例已更新，请在「邀请返利」查看当前档位。",
		LinkTab:  "sub-accounts",
	},
}

type UserNotificationService struct {
	notifications repository.UserNotificationsRepo
}

func NewUserNotificationService(notifications repository.UserNotificationsRepo) *UserNotificationService {
	return &UserNotificationService{notifications: notifications}
}

func (s *UserNotificationService) ListMine(ctx context.Context, userID uint, q *request.ListUserNotificationsQuery) (*response.UserNotificationListPageResp, error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize < 1 {
		pageSize = defaultNotificationPageSize
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	total, err := s.notifications.CountInAppByUserID(ctx, userID, q.UnreadOnly)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	rows, err := s.notifications.ListInAppByUserID(ctx, userID, q.UnreadOnly, offset, pageSize)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	items := make([]response.UserNotificationItemResp, 0, len(rows))
	for i := range rows {
		items = append(items, itemFromEntity(&rows[i]))
	}
	return &response.UserNotificationListPageResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *UserNotificationService) UnreadCount(ctx context.Context, userID uint) (*response.UserNotificationUnreadCountResp, error) {
	n, err := s.notifications.CountUnreadInAppByUserID(ctx, userID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	return &response.UserNotificationUnreadCountResp{Count: n}, nil
}

func (s *UserNotificationService) MarkRead(ctx context.Context, userID, id uint) error {
	row, err := s.notifications.FindInAppByIDForUser(ctx, userID, id)
	if err != nil {
		return errorx.ErrDbError
	}
	if row == nil {
		return errorx.ErrParamsError
	}
	if row.Status == entity.NotificationInAppRead {
		return nil
	}
	if err := s.notifications.MarkRead(ctx, userID, id); err != nil {
		return errorx.ErrDbError
	}
	return nil
}

func (s *UserNotificationService) MarkAllRead(ctx context.Context, userID uint) error {
	if err := s.notifications.MarkAllReadInApp(ctx, userID); err != nil {
		return errorx.ErrDbError
	}
	return nil
}

func itemFromEntity(row *entity.UserNotifications) response.UserNotificationItemResp {
	view := notificationTemplates[row.TemplateCode]
	if view.Title == "" {
		view = templateView{
			Category: "system",
			Title:    "系统通知",
			Body:     row.TemplateCode,
			LinkTab:  "messages",
		}
	}
	read := row.Status == entity.NotificationInAppRead
	return response.UserNotificationItemResp{
		ID:           row.ID,
		Category:     view.Category,
		Title:        view.Title,
		Body:         view.Body,
		TemplateCode: row.TemplateCode,
		Read:         read,
		CreatedAt:    formatNotificationTime(row.CreatedAt),
		LinkTab:      view.LinkTab,
	}
}

func formatNotificationTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}
