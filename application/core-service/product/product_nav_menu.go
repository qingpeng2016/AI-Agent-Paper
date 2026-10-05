package product

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
)

const navTopBrandCategoryLimit = 3

func (s *ProductService) buildNavMenu(categories []entity.ProductsCategory, products []entity.Products) *response.NavMenuResp {
	byCategoryID := map[uint][]response.NavPlanItemResp{}
	for i := range products {
		p := &products[i]
		if p.ProductsCategoryID == nil {
			continue
		}
		id := *p.ProductsCategoryID
		byCategoryID[id] = append(byCategoryID[id], mapProductToNavItem(p))
	}

	mega := make([]response.NavMegaColumnResp, 0, len(categories))
	brands := make([]response.NavBrandMenuResp, 0, navTopBrandCategoryLimit)

	for _, cat := range categories {
		items := byCategoryID[cat.ID]
		if len(items) == 0 {
			continue
		}
		mega = append(mega, response.NavMegaColumnResp{Title: cat.Name, Items: items})
		if len(brands) < navTopBrandCategoryLimit {
			brands = append(brands, response.NavBrandMenuResp{
				ID:         strconv.FormatUint(uint64(cat.ID), 10),
				Label:      cat.Name,
				HotTagName: strings.TrimSpace(cat.HotTagName),
				Items:      items,
			})
		}
	}

	return &response.NavMenuResp{MegaMenu: mega, BrandMenus: brands}
}

func mapProductToNavItem(p *entity.Products) response.NavPlanItemResp {
	label := strings.TrimSpace(p.CardTitle)
	label = strings.TrimSuffix(label, "月卡")
	label = strings.TrimSpace(label)
	if label == "" {
		label = p.SKUProductName
	}
	return response.NavPlanItemResp{
		Label: label,
		Price: formatNavPrice(p.Price, p.BillingPeriod),
		Href:  productDetailHref(p.SKUCode),
	}
}

func productDetailHref(skuCode string) string {
	if slug, ok := skuToSlug[skuCode]; ok {
		return "/p/" + slug
	}
	return "/p/" + strings.ToLower(strings.ReplaceAll(skuCode, "_", "-"))
}

func formatNavPrice(price decimal.Decimal, billingPeriod string) string {
	yuan := formatCnyYuan(price)
	switch billingPeriod {
	case "month":
		return yuan + "/月"
	case "year":
		return yuan + "/年"
	case "once":
		return yuan
	default:
		return yuan + "/月"
	}
}

func formatCnyYuan(price decimal.Decimal) string {
	if price.Equal(price.Truncate(0)) {
		return fmt.Sprintf("¥%s", price.StringFixed(0))
	}
	return fmt.Sprintf("¥%s", price.StringFixed(2))
}
