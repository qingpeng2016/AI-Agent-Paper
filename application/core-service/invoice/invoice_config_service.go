package invoice

import (
	"context"
	"strings"
	"time"

	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
)

type InvoiceConfigService struct {
	configs repository.UserInvoiceConfigRepo
}

func NewInvoiceConfigService(configs repository.UserInvoiceConfigRepo) *InvoiceConfigService {
	return &InvoiceConfigService{configs: configs}
}

func (s *InvoiceConfigService) ListMine(ctx context.Context, userID uint) ([]response.InvoiceConfigItem, error) {
	rows, err := s.configs.ListByUserID(ctx, userID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	items := make([]response.InvoiceConfigItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, invoiceConfigItemFromEntity(&row))
	}
	return items, nil
}

func (s *InvoiceConfigService) Create(ctx context.Context, userID uint, req *request.SaveInvoiceConfigReq) (*response.InvoiceConfigItem, error) {
	if err := validateInvoiceConfigReq(req); err != nil {
		return nil, err
	}
	now := time.Now()
	row := entityFromSaveReq(userID, req, now)
	if req.IsDefault {
		_ = s.configs.ClearDefaultForUser(ctx, userID, 0)
		row.IsDefault = 1
	}
	if err := s.configs.Create(ctx, &row); err != nil {
		return nil, errorx.ErrDbError
	}
	item := invoiceConfigItemFromEntity(&row)
	return &item, nil
}

func (s *InvoiceConfigService) Update(ctx context.Context, userID, id uint, req *request.SaveInvoiceConfigReq) (*response.InvoiceConfigItem, error) {
	if err := validateInvoiceConfigReq(req); err != nil {
		return nil, err
	}
	existed, err := s.configs.FindByIDForUser(ctx, userID, id)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	if existed == nil {
		return nil, errorx.ErrParamsError
	}
	if req.IsDefault {
		_ = s.configs.ClearDefaultForUser(ctx, userID, id)
	}
	now := time.Now()
	title := strings.TrimSpace(req.Title)
	if title == "" && req.TaxNo != nil {
		title = strings.TrimSpace(*req.TaxNo)
	}
	fields := map[string]interface{}{
		"profile_type": normalizeProfileType(req.ProfileType),
		"title":        title,
		"tax_no":       trimOptional(req.TaxNo),
		"bank_name":    trimOptional(req.BankName),
		"bank_account": trimOptional(req.BankAccount),
		"address":      trimOptional(req.Address),
		"phone":        trimOptional(req.Phone),
		"is_default":   boolToInt(req.IsDefault),
		"updated_at":   now,
	}
	if err := s.configs.Update(ctx, id, fields); err != nil {
		return nil, errorx.ErrDbError
	}
	updated, err := s.configs.FindByIDForUser(ctx, userID, id)
	if err != nil || updated == nil {
		return nil, errorx.ErrDbError
	}
	item := invoiceConfigItemFromEntity(updated)
	return &item, nil
}

func (s *InvoiceConfigService) SetDefault(ctx context.Context, userID, id uint) error {
	existed, err := s.configs.FindByIDForUser(ctx, userID, id)
	if err != nil {
		return errorx.ErrDbError
	}
	if existed == nil {
		return errorx.ErrParamsError
	}
	if existed.IsDefault == 1 {
		return nil
	}
	if err := s.configs.ClearDefaultForUser(ctx, userID, id); err != nil {
		return errorx.ErrDbError
	}
	now := time.Now()
	if err := s.configs.Update(ctx, id, map[string]interface{}{
		"is_default": 1,
		"updated_at": now,
	}); err != nil {
		return errorx.ErrDbError
	}
	return nil
}

func validateInvoiceConfigReq(req *request.SaveInvoiceConfigReq) error {
	pt := normalizeProfileType(req.ProfileType)
	title := strings.TrimSpace(req.Title)
	taxNo := ""
	if req.TaxNo != nil {
		taxNo = strings.TrimSpace(*req.TaxNo)
	}
	if pt == "personal" {
		if title == "" {
			return errorx.ErrParamsError
		}
		return nil
	}
	if title == "" && taxNo == "" {
		return errorx.ErrParamsError
	}
	return nil
}

func entityFromSaveReq(userID uint, req *request.SaveInvoiceConfigReq, now time.Time) entity.UserInvoiceConfig {
	title := strings.TrimSpace(req.Title)
	if title == "" && req.TaxNo != nil {
		title = strings.TrimSpace(*req.TaxNo)
	}
	row := entity.UserInvoiceConfig{
		UserID:      userID,
		ProfileType: normalizeProfileType(req.ProfileType),
		Title:       title,
		TaxNo:       trimOptional(req.TaxNo),
		BankName:    trimOptional(req.BankName),
		BankAccount: trimOptional(req.BankAccount),
		Address:     trimOptional(req.Address),
		Phone:       trimOptional(req.Phone),
		IsDefault:   boolToInt(req.IsDefault),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return row
}

func invoiceConfigItemFromEntity(row *entity.UserInvoiceConfig) response.InvoiceConfigItem {
	item := response.InvoiceConfigItem{
		ID:          row.ID,
		ProfileType: row.ProfileType,
		Title:       row.Title,
		IsDefault:   row.IsDefault == 1,
		CreatedAt:   formatConfigTime(row.CreatedAt),
		UpdatedAt:   formatConfigTime(row.UpdatedAt),
	}
	if row.TaxNo != nil {
		item.TaxNo = *row.TaxNo
	}
	if row.BankName != nil {
		item.BankName = *row.BankName
	}
	if row.BankAccount != nil {
		item.BankAccount = *row.BankAccount
	}
	if row.Address != nil {
		item.Address = *row.Address
	}
	if row.Phone != nil {
		item.Phone = *row.Phone
	}
	return item
}

func normalizeProfileType(raw string) string {
	if strings.ToLower(strings.TrimSpace(raw)) == "personal" {
		return "personal"
	}
	return "enterprise"
}

func trimOptional(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil
	}
	return &t
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func formatConfigTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}
