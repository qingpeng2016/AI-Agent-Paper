package user_api_key

import (
	"context"
	"strings"

	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
)

func (s *Service) latestEnterpriseInquiryID(ctx context.Context, ownerUserID uint) (uint, error) {
	row, err := s.inquiries.FindLatestByOwnerUserID(ctx, ownerUserID)
	if err != nil {
		return 0, errorx.ErrDbError
	}
	if row == nil {
		return 0, nil
	}
	return row.ID, nil
}

func (s *Service) GetEnterpriseInquiryStatus(ctx context.Context, ownerUserID uint) (*response.ApiTeamEnterpriseInquiryStatusResp, error) {
	row, err := s.inquiries.FindLatestByOwnerUserID(ctx, ownerUserID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	if row == nil {
		return &response.ApiTeamEnterpriseInquiryStatusResp{HasInquiry: false}, nil
	}
	return &response.ApiTeamEnterpriseInquiryStatusResp{
		HasInquiry: true,
		Inquiry:    mapInquiryBrief(row),
	}, nil
}

func (s *Service) SubmitEnterpriseInquiry(ctx context.Context, ownerUserID uint, req *request.ApiTeamEnterpriseInquiryReq) (*response.ApiTeamEnterpriseInquirySubmitResp, error) {
	existing, err := s.inquiries.FindLatestByOwnerUserID(ctx, ownerUserID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	if existing != nil {
		return &response.ApiTeamEnterpriseInquirySubmitResp{
			ID:      existing.ID,
			Message: "企业信息已存在",
		}, nil
	}
	u, err := s.users.FindByID(ctx, ownerUserID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	contactName := inquiryContactNameFromUser(u)
	phone := inquiryPhoneFromUser(u)
	owner := ownerUserID
	row := &entity.EnterpriseInquiry{
		OwnerUserID: &owner,
		CompanyName: strings.TrimSpace(req.CompanyName),
		ContactName: contactName,
		Phone:       phone,
		Status:      "pending",
	}
	if u.Email != nil {
		if email := strings.TrimSpace(*u.Email); email != "" {
			row.Email = &email
		}
	}
	if err := s.inquiries.Create(ctx, row); err != nil {
		return nil, errorx.ErrDbError
	}
	return &response.ApiTeamEnterpriseInquirySubmitResp{
		ID:      row.ID,
		Message: "企业信息已保存",
	}, nil
}

func inquiryContactNameFromUser(u *entity.Users) string {
	if u.Nickname != nil {
		if n := strings.TrimSpace(*u.Nickname); n != "" {
			return n
		}
	}
	if u.Email != nil {
		if e := strings.TrimSpace(*u.Email); e != "" {
			return e
		}
	}
	if u.Phone != nil {
		if p := strings.TrimSpace(*u.Phone); p != "" {
			return p
		}
	}
	return "团队管理员"
}

func inquiryPhoneFromUser(u *entity.Users) string {
	if u.Phone != nil {
		if p := strings.TrimSpace(*u.Phone); p != "" {
			return p
		}
	}
	return "-"
}

func mapInquiryBrief(row *entity.EnterpriseInquiry) *response.ApiTeamEnterpriseInquiryBrief {
	b := &response.ApiTeamEnterpriseInquiryBrief{
		ID:          row.ID,
		CompanyName: row.CompanyName,
		ContactName: row.ContactName,
		Phone:       row.Phone,
		Status:      row.Status,
	}
	if row.Email != nil {
		b.Email = *row.Email
	}
	return b
}
