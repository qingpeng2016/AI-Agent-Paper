package viplevelsync

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qingpeng2016/ai-token-mall/application/core-service/invite_rebate"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const (
	ModuleAITokenMall = "ai_token_mall"
	TaskVipLevelSync  = "vip_level_sync"
	syncBatchSize     = 100
	listPageSize      = 500
)

type vipSyncPlan struct {
	enabledTiers     []entity.VipConfig
	tierRankByID     map[uint]int
	inviteCounts     map[uint]int64
	inviteePaidTotal map[uint]decimal.Decimal
}

type VipLevelSyncJob struct {
	users         repository.UsersRepo
	vipConfigs    repository.VipConfigRepo
	notifications repository.UserNotificationsRepo
}

func NewVipLevelSyncJob(
	users repository.UsersRepo,
	vipConfigs repository.VipConfigRepo,
	notifications repository.UserNotificationsRepo,
) *VipLevelSyncJob {
	return &VipLevelSyncJob{
		users:         users,
		vipConfigs:    vipConfigs,
		notifications: notifications,
	}
}

// Run 扫描全部用户，按有效邀请人数匹配 VIP 档位；仅当可达最高档高于当前档时才升级 vip_config_id。
func (j *VipLevelSyncJob) Run(ctx context.Context) {
	plan, err := j.buildSyncPlan(ctx)
	if err != nil {
		return
	}
	if plan == nil {
		return
	}

	allUsers, err := j.listAllUsersForSync(ctx)
	if err != nil {
		logger.ErrorZ(ctx, "vip-level-sync-list-users-failed", zap.Error(err))
		return
	}

	var scanned, upgraded int64
	if err := j.runUsersByBatch(ctx, plan, allUsers, &scanned, &upgraded); err != nil {
		logger.ErrorZ(ctx, "vip-level-sync-batch-failed", zap.Error(err))
		return
	}

	logger.InfoZ(ctx, "vip-level-sync-finished",
		zap.Int64("scanned", scanned),
		zap.Int64("upgraded", upgraded))
}

func (j *VipLevelSyncJob) buildSyncPlan(ctx context.Context) (*vipSyncPlan, error) {
	enabledTiers, err := j.vipConfigs.ListEnabled(ctx)
	if err != nil {
		logger.ErrorZ(ctx, "vip-level-sync-list-tiers-failed", zap.Error(err))
		return nil, err
	}
	if len(enabledTiers) == 0 {
		logger.InfoZ(ctx, "vip-level-sync-no-enabled-tiers")
		return nil, nil
	}
	allTiers, err := j.vipConfigs.ListAll(ctx)
	if err != nil {
		logger.ErrorZ(ctx, "vip-level-sync-list-all-tiers-failed", zap.Error(err))
		return nil, err
	}
	tierRankByID := make(map[uint]int, len(allTiers))
	for i := range allTiers {
		tierRankByID[allTiers[i].ID] = allTiers[i].SortOrder
	}

	inviteCounts, err := j.users.CountInviteesGroupByInviter(ctx)
	if err != nil {
		logger.ErrorZ(ctx, "vip-level-sync-count-invitees-failed", zap.Error(err))
		return nil, err
	}

	inviteePaidTotal, err := j.users.SumInviteeCompletedOrderAmountGroupByInviter(ctx)
	if err != nil {
		logger.ErrorZ(ctx, "vip-level-sync-sum-invitee-paid-failed", zap.Error(err))
		return nil, err
	}

	return &vipSyncPlan{
		enabledTiers:     enabledTiers,
		tierRankByID:     tierRankByID,
		inviteCounts:     inviteCounts,
		inviteePaidTotal: inviteePaidTotal,
	}, nil
}

func (j *VipLevelSyncJob) listAllUsersForSync(ctx context.Context) ([]repository.UserVipConfigRow, error) {
	var all []repository.UserVipConfigRow
	offset := 0
	for {
		rows, err := j.users.ListIDAndVipConfigID(ctx, offset, listPageSize)
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			break
		}
		all = append(all, rows...)
		offset += len(rows)
		if len(rows) < listPageSize {
			break
		}
	}
	return all, nil
}

func (j *VipLevelSyncJob) runUsersByBatch(
	ctx context.Context,
	plan *vipSyncPlan,
	users []repository.UserVipConfigRow,
	scanned, upgraded *int64,
) error {
	for start := 0; start < len(users); start += syncBatchSize {
		end := start + syncBatchSize
		if end > len(users) {
			end = len(users)
		}
		batch := users[start:end]
		var wg sync.WaitGroup
		for _, u := range batch {
			wg.Add(1)
			row := u
			go func() {
				defer wg.Done()
				atomic.AddInt64(scanned, 1)
				didUpgrade, err := j.syncOneUser(ctx, plan, row)
				if err != nil {
					logger.ErrorZ(ctx, "vip-level-sync-user-failed",
						zap.Uint("user_id", row.ID),
						zap.Error(err))
					return
				}
				if didUpgrade {
					atomic.AddInt64(upgraded, 1)
				}
			}()
		}
		wg.Wait()
	}
	return nil
}

func (j *VipLevelSyncJob) syncOneUser(
	ctx context.Context,
	plan *vipSyncPlan,
	u repository.UserVipConfigRow,
) (bool, error) {
	valid := plan.inviteCounts[u.ID]
	paidTotal := plan.inviteePaidTotal[u.ID]
	best := invite_rebate.BestQualifyingVipTier(plan.enabledTiers, valid, paidTotal)
	if best == nil {
		return false, nil
	}
	currentRank, ok := plan.tierRankByID[u.VipConfigID]
	if !ok {
		currentRank = -1
	}
	if best.SortOrder <= currentRank {
		return false, nil
	}
	if err := j.users.UpdateVipConfigID(ctx, u.ID, best.ID); err != nil {
		return false, err
	}
	if err := j.insertVipUpgradeNotification(ctx, u.ID); err != nil {
		return false, err
	}
	return true, nil
}

func (j *VipLevelSyncJob) insertVipUpgradeNotification(ctx context.Context, userID uint) error {
	now := time.Now()
	sentAt := now
	return j.notifications.Create(ctx, nil, &entity.UserNotifications{
		UserID:       userID,
		Channel:      "in_app",
		TemplateCode: "vip_level_upgraded",
		Status:       entity.NotificationInAppUnread,
		SentAt:       &sentAt,
		CreatedAt:    now,
	})
}
