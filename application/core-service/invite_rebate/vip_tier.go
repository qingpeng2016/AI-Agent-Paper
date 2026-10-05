package invite_rebate

import (
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/shopspring/decimal"
)

// BestQualifyingVipTier 在已启用档位中，取邀请人数与下级累计消费均满足的最高档（sort_order 最大）。
func BestQualifyingVipTier(tiers []entity.VipConfig, validInvites int64, inviteePaidTotal decimal.Decimal) *entity.VipConfig {
	var best *entity.VipConfig
	for i := range tiers {
		t := &tiers[i]
		if !t.Enabled {
			continue
		}
		if validInvites < int64(t.MinValidInvites) {
			continue
		}
		if inviteePaidTotal.LessThan(t.MinInviteePaidAmount) {
			continue
		}
		if best == nil || t.SortOrder > best.SortOrder ||
			(t.SortOrder == best.SortOrder && t.ID > best.ID) {
			best = t
		}
	}
	return best
}
