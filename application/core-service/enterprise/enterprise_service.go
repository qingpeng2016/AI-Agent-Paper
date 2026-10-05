package enterprise

import (
	"context"
	"errors"
	"strings"

	"github.com/qingpeng2016/ai-token-mall/application/core-service/shared"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
)

var ErrEnterpriseProductNotFound = errors.New("enterprise product not found")

type EnterpriseService struct {
	inquiryRepo  repository.EnterpriseInquiryRepo
	productsRepo repository.EnterpriseProductsRepo
}

func NewEnterpriseService(inquiryRepo repository.EnterpriseInquiryRepo, productsRepo repository.EnterpriseProductsRepo) *EnterpriseService {
	return &EnterpriseService{inquiryRepo: inquiryRepo, productsRepo: productsRepo}
}

func (s *EnterpriseService) SubmitInquiry(ctx context.Context, req *request.SubmitEnterpriseInquiryReq, ownerUserID *uint) (uint, error) {
	row := &entity.EnterpriseInquiry{
		CompanyName: strings.TrimSpace(req.CompanyName),
		ContactName: strings.TrimSpace(req.ContactName),
		Phone:       strings.TrimSpace(req.Phone),
		Status:      "pending",
	}
	if ownerUserID != nil && *ownerUserID > 0 {
		row.OwnerUserID = ownerUserID
	}
	email := strings.TrimSpace(req.Email)
	if email != "" {
		row.Email = &email
	}
	if err := s.inquiryRepo.Create(ctx, row); err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (s *EnterpriseService) ListProducts(ctx context.Context) (*response.EnterpriseProductListResp, error) {
	rows, err := s.productsRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]response.EnterpriseProductItemResp, 0, len(rows))
	for i := range rows {
		items = append(items, mapEnterpriseProductItem(&rows[i]))
	}
	return &response.EnterpriseProductListResp{Products: items}, nil
}

func (s *EnterpriseService) GetProduct(ctx context.Context, code string) (*response.EnterpriseProductItemResp, error) {
	p, err := s.productsRepo.FindActiveByCode(ctx, strings.TrimSpace(code))
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrEnterpriseProductNotFound
	}
	item := mapEnterpriseProductItem(p)
	return &item, nil
}

func mapEnterpriseProductItem(p *entity.EnterpriseProducts) response.EnterpriseProductItemResp {
	features := shared.DecodeStringJSONArray(p.Features)
	if features == nil {
		features = []string{}
	}
	return response.EnterpriseProductItemResp{
		Code:        p.Code,
		Name:        p.Name,
		Badge:       p.Badge,
		PriceHint:   p.PriceHint,
		Seats:       p.Seats,
		Features:    features,
		Tagline:     p.Tagline,
		ButtonLabel: p.ButtonLabel,
		IsFeatured:  p.IsFeatured != 0,
		Sort:        p.Sort,
	}
}
