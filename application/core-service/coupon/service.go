package coupon

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
	"github.com/shopspring/decimal"
)

type Service struct {
	campaigns repository.CouponCampaignsRepo
	coupons   repository.UserCouponsRepo
}

func NewService(
	campaigns repository.CouponCampaignsRepo,
	coupons repository.UserCouponsRepo,
) *Service {
	return &Service{campaigns: campaigns, coupons: coupons}
}

func (s *Service) RegisterPromo(ctx context.Context) (*response.RegisterCouponPromoResp, error) {
	rows, err := s.campaigns.ListAutoGrantOnRegister(ctx)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	if len(rows) == 0 {
		return &response.RegisterCouponPromoResp{Active: false}, nil
	}
	c := rows[0]
	sub := ""
	if c.Subtitle != nil {
		sub = *c.Subtitle
	}
	return &response.RegisterCouponPromoResp{
		Active:         true,
		Title:          c.Title,
		Subtitle:       sub,
		DiscountLabel:  formatDiscountLabel(c.DiscountType, c.DiscountValue),
		ValidDays:      c.ValidDays,
		CampaignCode:   c.Code,
	}, nil
}

func (s *Service) ListMine(ctx context.Context, userID uint) (*response.UserCouponListResp, error) {
	rows, err := s.coupons.ListByUserID(ctx, userID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	items := make([]response.UserCouponItem, 0, len(rows))
	for _, row := range rows {
		camp, _ := s.campaigns.FindByID(ctx, row.CampaignID)
		title := ""
		code := ""
		if camp != nil {
			title = camp.Title
			code = camp.Code
		}
		items = append(items, response.UserCouponItem{
			ID:             row.ID,
			CouponCode:     row.CouponCode,
			CampaignCode:   code,
			Title:          title,
			DiscountType:   row.DiscountType,
			DiscountValue:  row.DiscountValue.StringFixed(2),
			MinOrderAmount: row.MinOrderAmount.StringFixed(2),
			Status:         row.Status,
			ValidFrom:      row.ValidFrom.Format("2006-01-02"),
			ValidUntil:     row.ValidUntil.Format("2006-01-02"),
		})
	}
	return &response.UserCouponListResp{Items: items}, nil
}

func (s *Service) GrantRegisterCoupons(ctx context.Context, userID uint) (int, error) {
	if userID == 0 {
		return 0, nil
	}
	campaigns, err := s.campaigns.ListAutoGrantOnRegister(ctx)
	if err != nil {
		return 0, err
	}
	granted := 0
	now := time.Now()
	for _, camp := range campaigns {
		dup, err := s.coupons.ExistsByUserAndCampaign(ctx, userID, camp.ID)
		if err != nil {
			return granted, err
		}
		if dup {
			continue
		}
		validUntil := now.AddDate(0, 0, int(camp.ValidDays))
		row := &entity.UserCoupons{
			UserID:         userID,
			CampaignID:     camp.ID,
			CouponCode:     fmt.Sprintf("UC-%d-%d-%d", userID, camp.ID, now.UnixNano()),
			DiscountType:   camp.DiscountType,
			DiscountValue:  camp.DiscountValue,
			MinOrderAmount: camp.MinOrderAmount,
			Status:         "available",
			ValidFrom:      now,
			ValidUntil:     validUntil,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := s.coupons.Create(ctx, nil, row); err != nil {
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
				continue
			}
			return granted, err
		}
		granted++
	}
	return granted, nil
}

func formatDiscountLabel(discountType string, value decimal.Decimal) string {
	if discountType == "percent" {
		return value.StringFixed(0) + "% OFF"
	}
	return "¥" + value.StringFixed(0)
}
