package product

import (
	"context"
	"errors"
	"strings"

	"github.com/qingpeng2016/ai-token-mall/application/core-service/shared"
	"github.com/qingpeng2016/ai-token-mall/common/billing"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
)

var ErrProductNotFound = errors.New("product not found")

type ProductService struct {
	categoryRepo repository.ProductsCategoryRepo
	productRepo  repository.ProductsRepo
}

func NewProductService(categoryRepo repository.ProductsCategoryRepo, productRepo repository.ProductsRepo) *ProductService {
	return &ProductService{categoryRepo: categoryRepo, productRepo: productRepo}
}

func (s *ProductService) List(ctx context.Context, categoryID uint) (*response.ProductCatalogResp, error) {
	categories, err := s.categoryRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	products, err := s.productRepo.ListOnSale(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	return buildProductCatalog(categories, products), nil
}

func (s *ProductService) NavMenu(ctx context.Context) (*response.NavMenuResp, error) {
	categories, err := s.categoryRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.productRepo.ListOnSale(ctx, 0)
	if err != nil {
		return nil, err
	}
	return s.buildNavMenu(categories, rows), nil
}

func buildProductCatalog(categories []entity.ProductsCategory, products []entity.Products) *response.ProductCatalogResp {
	byCategoryID := map[uint][]response.ProductItemResp{}
	for i := range products {
		p := &products[i]
		if p.ProductsCategoryID == nil {
			continue
		}
		id := *p.ProductsCategoryID
		byCategoryID[id] = append(byCategoryID[id], mapProductItem(p))
	}

	groups := make([]response.ProductCategoryGroupResp, 0, len(categories))
	for i := range categories {
		c := &categories[i]
		items := byCategoryID[c.ID]
		if items == nil {
			items = []response.ProductItemResp{}
		}
		groups = append(groups, response.ProductCategoryGroupResp{
			ID:         c.ID,
			Name:       c.Name,
			DotColor:   c.DotColor,
			ActiveBg:   c.ActiveBg,
			Sort:       c.Sort,
			HotTagName: strings.TrimSpace(c.HotTagName),
			Products:   items,
		})
	}
	return &response.ProductCatalogResp{Categories: groups}
}

func mapProductItem(p *entity.Products) response.ProductItemResp {
	item := response.ProductItemResp{
		ID:                   p.ID,
		SKUCode:              p.SKUCode,
		CardTitle:            p.CardTitle,
		CardSubtitle:         p.CardSubtitle,
		CardFeatures:         expandCardFeatures(shared.DecodeStringJSONArray(p.CardFeatures), p),
		ShareSeats:           p.ShareSeats,
		ProductsCategoryID:   p.ProductsCategoryID,
		ProductsCategoryName: p.ProductsCategoryName,
		SKUProductName:       p.SKUProductName,
		LimitTokens:          p.LimitTokens,
		RPMLimit:             p.RPMLimit,
		TPMLimit:             p.TPMLimit,
		AllowedModels:        shared.DecodeStringJSONArray(p.AllowedModels),
		BillingPeriod:        p.BillingPeriod,
		PeriodDays:           billing.PeriodDays(p.PeriodDays, p.BillingPeriod),
		Price:                response.MoneyFrom(p.Price),
		Currency:             p.Currency,
		HotTagName:           strings.TrimSpace(p.HotTagName),
		Sort:                 p.SortOrder,
		Status:               p.Status,
	}
	if len(item.CardFeatures) == 0 {
		item.CardFeatures = []string{}
	}
	if len(item.AllowedModels) == 0 {
		item.AllowedModels = []string{}
	}
	return item
}

func (s *ProductService) DetailBySlug(ctx context.Context, slug string) (*response.ProductDetailResp, error) {
	sku := skuFromSlug(slug)
	if sku == "" {
		return nil, ErrProductNotFound
	}
	p, err := s.productRepo.FindOnSaleBySKUCode(ctx, sku)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProductNotFound
	}
	item := mapProductItem(p)
	detail := buildProductDetailContent(slug, p, item)
	return &response.ProductDetailResp{
		Slug:    canonicalSlugForSKU(p.SKUCode),
		Product: item,
		Detail:  detail,
	}, nil
}
