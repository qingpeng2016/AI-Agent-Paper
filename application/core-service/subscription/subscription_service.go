package subscription

import (
	"context"
	"sort"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
)

type SubscriptionService struct {
	subs repository.UserSubscriptionsRepo
}

func NewSubscriptionService(subs repository.UserSubscriptionsRepo) *SubscriptionService {
	return &SubscriptionService{subs: subs}
}

func (s *SubscriptionService) ListMine(ctx context.Context, userID uint) ([]response.UserSubscriptionItem, error) {
	rows, err := s.subs.ListByUserID(ctx, userID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	items := make([]response.UserSubscriptionItem, 0, len(rows))
	for _, row := range rows {
		name := row.ProductCardTitle
		if name == "" {
			name = row.SKUProductName
		}
		items = append(items, response.UserSubscriptionItem{
			ID:             row.ID,
			ProductID:      row.ProductID,
			ProductName:    name,
			Status:         row.Status,
			LimitTokens:    row.LimitTokens,
			UsedTokens:     row.UsedTokens,
			PeriodEnd:      formatSubscriptionDate(row.PeriodEnd),
			ExpiresAt:      formatSubscriptionDate(row.ExpiresAt),
			SKUProductName:       row.SKUProductName,
			ProductsCategoryName: row.ProductsCategoryName,
		})
	}
	sortSubscriptionsForDisplay(items)
	return items, nil
}

func sortSubscriptionsForDisplay(items []response.UserSubscriptionItem) {
	sort.SliceStable(items, func(i, j int) bool {
		ai := items[i].Status == "active"
		aj := items[j].Status == "active"
		if ai != aj {
			return ai
		}
		return items[i].ID > items[j].ID
	})
}

func formatSubscriptionDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}
