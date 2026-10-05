package invite_rebate

import (
	"context"
	"errors"
	"strings"

	"github.com/shopspring/decimal"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
	"gorm.io/gorm"
)

var policyNotes = []string{
	"返佣比例按「邀请下级人数」与「下级累计消费」自动升级，下级为注册时已绑定到您账号的用户。",
	"受邀用户每笔已支付订单，按实付金额 × 当前返佣比例计算返利，支付成功后即时计入「佣金」。",
	"佣金可划转到余额或者提现。",
}

type Service struct {
	tx           repository.Transactor
	users        repository.UsersRepo
	vipConfigs   repository.VipConfigRepo
	records      repository.UserCommissionRecordsRepo
	withdraws    repository.UserCommissionWithdrawalsRepo
	payouts      repository.UserCommissionPayoutConfigRepo
	wallets      repository.UserWalletFlowsRepo
	notifications repository.UserNotificationsRepo
}

func NewService(
	tx repository.Transactor,
	users repository.UsersRepo,
	vipConfigs repository.VipConfigRepo,
	records repository.UserCommissionRecordsRepo,
	withdraws repository.UserCommissionWithdrawalsRepo,
	payouts repository.UserCommissionPayoutConfigRepo,
	wallets repository.UserWalletFlowsRepo,
	notifications repository.UserNotificationsRepo,
) *Service {
	return &Service{
		tx:            tx,
		users:         users,
		vipConfigs:    vipConfigs,
		records:       records,
		withdraws:     withdraws,
		payouts:       payouts,
		wallets:       wallets,
		notifications: notifications,
	}
}

func (s *Service) GetOverview(ctx context.Context, userID uint) (*response.InviteRebateOverviewResp, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.ErrUserNotFound
		}
		return nil, errorx.ErrDbError
	}

	tiers, err := s.vipConfigs.ListEnabled(ctx)
	if err != nil {
		return nil, errorx.ErrDbError
	}

	var currentTier *entity.VipConfig
	for i := range tiers {
		if tiers[i].ID == u.VipConfigID {
			currentTier = &tiers[i]
			break
		}
	}
	if currentTier == nil {
		currentTier, _ = s.vipConfigs.FindByID(ctx, u.VipConfigID)
	}
	if currentTier == nil && len(tiers) > 0 {
		currentTier = &tiers[0]
	}

	validCount, err := s.users.CountByParentUserID(ctx, userID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	inviteePaidTotal, err := s.users.SumInviteeCompletedOrderAmountByInviter(ctx, userID)
	if err != nil {
		return nil, errorx.ErrDbError
	}

	tierItems := make([]response.InviteRebateTierItem, 0, len(tiers))
	currentLabel := ""
	currentRate := response.MoneyFrom(decimal.Zero)
	if currentTier != nil {
		currentLabel = currentTier.LevelLabel
		currentRate = response.MoneyFrom(currentTier.RatePercent)
	}
	for _, t := range tiers {
		tierItems = append(tierItems, response.InviteRebateTierItem{
			LevelLabel:           t.LevelLabel,
			MinValidInvites:      t.MinValidInvites,
			MinInviteePaidAmount: response.MoneyFrom(t.MinInviteePaidAmount),
			RatePercent:          response.MoneyFrom(t.RatePercent),
			IsCurrent:            currentTier != nil && t.ID == currentTier.ID,
		})
	}

	promoURL := ""
	if u.VipDomain != nil && strings.TrimSpace(*u.VipDomain) != "" {
		promoURL = "https://" + strings.TrimSpace(*u.VipDomain)
	}

	return &response.InviteRebateOverviewResp{
		PromoDomainURL:      promoURL,
		CurrentLevelLabel:   currentLabel,
		CurrentRatePercent:  currentRate,
		ValidInviteCount:    validCount,
		InviteePaidTotal:    response.MoneyFrom(inviteePaidTotal),
		CommissionBalance:   response.MoneyFrom(u.CommissionBalance),
		Tiers:               tierItems,
		Notes:               policyNotes,
	}, nil
}

func (s *Service) ListMembers(ctx context.Context, userID uint) (*response.InviteRebateMemberListResp, error) {
	const maxMembers = 500
	rows, err := s.users.ListByParentUserID(ctx, userID, 0, maxMembers)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	total, err := s.users.CountByParentUserID(ctx, userID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	items := make([]response.InviteRebateMemberItem, 0, len(rows))
	for _, m := range rows {
		item := response.InviteRebateMemberItem{
			ID:           m.ID,
			Status:       "registered",
			RegisteredAt: formatDateTime(m.CreatedAt),
		}
		if m.Nickname != nil {
			item.Nickname = *m.Nickname
		}
		if m.Email != nil {
			item.Email = *m.Email
		}
		if m.Phone != nil {
			item.Phone = *m.Phone
		}
		items = append(items, item)
	}
	return &response.InviteRebateMemberListResp{Items: items, Total: total}, nil
}

func (s *Service) ListCommissionRecords(ctx context.Context, userID uint, q *request.InviteRebatePageQuery) (*response.InviteCommissionRecordListPageResp, error) {
	page, pageSize, offset := normalizePage(q.Page, q.PageSize)
	total, err := s.records.CountByInviterUserID(ctx, userID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	rows, err := s.records.ListByInviterUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	items := make([]response.InviteCommissionRecordItem, 0, len(rows))
	for _, row := range rows {
		item := response.InviteCommissionRecordItem{
			ID:           row.ID,
			OrderNo:      row.OrderNo,
			ProductName:  row.ProductName,
			OrderAmount:  response.MoneyFrom(row.OrderAmount),
			RebateAmount: response.MoneyFrom(row.RebateAmount),
			CreatedAt:    formatDateTime(row.CreatedAt),
		}
		invitee, err := s.users.FindByID(ctx, row.InviteeUserID)
		if err == nil && invitee != nil {
			if invitee.Nickname != nil {
				item.InviteeNickname = *invitee.Nickname
			}
			if invitee.Email != nil {
				item.InviteeEmail = *invitee.Email
			}
		}
		if item.InviteeNickname == "" && item.InviteeEmail != "" {
			item.InviteeNickname = item.InviteeEmail
		}
		items = append(items, item)
	}
	return &response.InviteCommissionRecordListPageResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *Service) ListWithdrawals(ctx context.Context, userID uint, q *request.InviteRebatePageQuery) (*response.InviteWithdrawalListPageResp, error) {
	page, pageSize, offset := normalizePage(q.Page, q.PageSize)
	total, err := s.withdraws.CountByUserID(ctx, userID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	rows, err := s.withdraws.ListByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	items := make([]response.InviteWithdrawalItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, response.InviteWithdrawalItem{
			ID:        row.ID,
			Amount:    response.MoneyFrom(row.Amount),
			Channel:   row.Channel,
			Status:    row.Status,
			CreatedAt: formatDateTime(row.CreatedAt),
		})
	}
	return &response.InviteWithdrawalListPageResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *Service) GetPayoutConfig(ctx context.Context, userID uint) (*response.InvitePayoutConfigResp, error) {
	rows, err := s.payouts.ListByUserID(ctx, userID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	resp := &response.InvitePayoutConfigResp{}
	for _, row := range rows {
		dataURL := qrImageDataURL(row.QrMime, row.QrImage)
		if row.Channel == "alipay" {
			resp.AlipayQrDataURL = dataURL
			resp.AlipayConfigured = row.HasQR()
		}
		if row.Channel == "wechat" {
			resp.WechatQrDataURL = dataURL
			resp.WechatConfigured = row.HasQR()
		}
	}
	return resp, nil
}

func (s *Service) SavePayoutQR(ctx context.Context, userID uint, channel string, image []byte, mime string) error {
	channel = strings.TrimSpace(strings.ToLower(channel))
	if channel != "alipay" && channel != "wechat" {
		return errorx.ErrParamsError
	}
	if err := validatePayoutQRImage(image, mime); err != nil {
		return err
	}
	mime = strings.TrimSpace(mime)
	return s.payouts.Save(ctx, &entity.UserCommissionPayoutConfig{
		UserID:  userID,
		Channel: channel,
		QrMime:  &mime,
		QrImage: image,
	})
}

func normalizePage(page, pageSize int) (int, int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 9
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize, (page - 1) * pageSize
}

func formatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}
