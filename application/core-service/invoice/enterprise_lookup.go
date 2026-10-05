package invoice

import (
	"context"
	"strings"

	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
	"github.com/qingpeng2016/ai-token-mall/infrastructure/http/invoicelookup"
)

type EnterpriseLookup struct {
	client *invoicelookup.Client
}

func NewEnterpriseLookup(client *invoicelookup.Client) *EnterpriseLookup {
	return &EnterpriseLookup{client: client}
}

func (s *EnterpriseLookup) Lookup(ctx context.Context, keyword string) (*response.EnterpriseInvoiceLookupResult, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, errorx.ErrParamsError
	}
	rows, err := s.client.Lookup(ctx, keyword)
	if err != nil || len(rows) == 0 {
		return nil, errorx.ErrInvoiceLookupNotFound
	}
	items := make([]response.EnterpriseInvoiceLookupItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, lookupItemFromRow(row))
	}
	match := pickBestMatch(keyword, items)
	var candidates []response.EnterpriseInvoiceLookupItem
	if len(items) > 1 {
		for _, it := range items {
			if it.Title != match.Title || it.TaxNo != match.TaxNo {
				candidates = append(candidates, it)
			}
		}
	}
	return &response.EnterpriseInvoiceLookupResult{
		Match:      match,
		Candidates: candidates,
	}, nil
}

func lookupItemFromRow(row invoicelookup.LookupItem) response.EnterpriseInvoiceLookupItem {
	return response.EnterpriseInvoiceLookupItem{
		Title:       row.Title,
		TaxNo:       row.TaxNo,
		BankName:    row.BankName,
		BankAccount: row.BankAccount,
		Address:     row.Address,
		Phone:       row.Phone,
	}
}

func pickBestMatch(keyword string, items []response.EnterpriseInvoiceLookupItem) response.EnterpriseInvoiceLookupItem {
	kw := strings.TrimSpace(keyword)
	kwUpper := strings.ToUpper(kw)
	for _, it := range items {
		if strings.EqualFold(strings.TrimSpace(it.TaxNo), kw) || strings.ToUpper(strings.TrimSpace(it.TaxNo)) == kwUpper {
			return it
		}
	}
	for _, it := range items {
		if strings.TrimSpace(it.Title) == kw {
			return it
		}
	}
	for _, it := range items {
		if strings.Contains(it.Title, kw) || strings.Contains(kw, it.Title) {
			return it
		}
	}
	return items[0]
}
