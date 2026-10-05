package user

import (
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	couponSvc "github.com/qingpeng2016/ai-token-mall/application/core-service/coupon"
	"github.com/qingpeng2016/ai-token-mall/common/auth"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var cnPhonePattern = regexp.MustCompile(`^1\d{10}$`)

type UserService struct {
	users      repository.UsersRepo
	wallets    repository.UserWalletFlowsRepo
	invoices   repository.UserInvoicesRepo
	vipConfigs repository.VipConfigRepo
	vipDomains repository.VipDomainConfigRepo
	coupons    *couponSvc.Service
}

func NewUserService(
	users repository.UsersRepo,
	wallets repository.UserWalletFlowsRepo,
	invoices repository.UserInvoicesRepo,
	vipConfigs repository.VipConfigRepo,
	vipDomains repository.VipDomainConfigRepo,
	coupons *couponSvc.Service,
) *UserService {
	return &UserService{
		users:      users,
		wallets:    wallets,
		invoices:   invoices,
		vipConfigs: vipConfigs,
		vipDomains: vipDomains,
		coupons:    coupons,
	}
}

func (s *UserService) GetProfile(ctx context.Context, userID uint) (*response.UserProfileResp, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.ErrUserNotFound
		}
		return nil, errorx.ErrDbError
	}
	return toUserProfile(u), nil
}

func (s *UserService) ChangePassword(ctx context.Context, userID uint, req *request.ChangePasswordReq) error {
	if req.NewPassword != req.ConfirmPassword {
		return errorx.ErrPasswordMismatch
	}
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorx.ErrUserNotFound
		}
		return errorx.ErrDbError
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.OldPassword)); err != nil {
		return errorx.ErrWrongPassword
	}
	if req.OldPassword == req.NewPassword {
		return errorx.ErrParamsError
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errorx.ErrUnknown
	}
	if err := s.users.UpdatePassword(ctx, userID, string(hash), req.NewPassword); err != nil {
		return errorx.ErrDbError
	}
	return nil
}

func (s *UserService) ListWalletFlows(ctx context.Context, userID uint, q *request.ListWalletFlowsQuery) (*response.UserWalletFlowListPageResp, error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize < 1 {
		pageSize = 9
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	total, err := s.wallets.CountByUserID(ctx, userID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	rows, err := s.wallets.ListByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	items := make([]response.UserWalletFlowItem, 0, len(rows))
	for _, row := range rows {
		remark := ""
		if row.Remark != nil {
			remark = *row.Remark
		}
		items = append(items, response.UserWalletFlowItem{
			ID:        row.ID,
			Type:      row.Type,
			Amount:    response.MoneyFrom(row.Amount),
			Remark:    remark,
			CreatedAt: formatUserDateTime(row.CreatedAt),
		})
	}
	return &response.UserWalletFlowListPageResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *UserService) ListInvoices(ctx context.Context, userID uint, q *request.ListInvoicesQuery) (*response.UserInvoiceListPageResp, error) {
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize < 1 {
		pageSize = 9
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	total, err := s.invoices.CountByUserID(ctx, userID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	rows, err := s.invoices.ListByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	items := make([]response.UserInvoiceListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, response.UserInvoiceListItem{
			ID:        row.ID,
			OrderNo:   row.OrderNo,
			Title:     row.Title,
			Amount:    response.MoneyFrom(row.Amount),
			Status:    row.Status,
			CreatedAt: formatUserDateTime(row.CreatedAt),
		})
	}
	return &response.UserInvoiceListPageResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *UserService) Register(ctx context.Context, req *request.RegisterUserReq) (*response.LoginUserResp, error) {
	if req.Password != req.ConfirmPassword {
		return nil, errorx.ErrPasswordMismatch
	}

	email := strings.TrimSpace(req.Email)
	phone := strings.TrimSpace(req.Phone)
	if err := validateLoginIdentity(email, phone); err != nil {
		return nil, err
	}

	if email != "" {
		ex, err := s.users.FindByEmail(ctx, email)
		if err != nil {
			return nil, errorx.ErrDbError
		}
		if ex != nil {
			return nil, errorx.ErrUserExists
		}
	}
	if phone != "" {
		ex, err := s.users.FindByPhone(ctx, phone)
		if err != nil {
			return nil, errorx.ErrDbError
		}
		if ex != nil {
			return nil, errorx.ErrUserExists
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errorx.ErrUnknown
	}

	u := &entity.Users{
		PasswordHash:  string(hash),
		PasswordPlain: req.Password,
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if email != "" {
		u.Email = &email
	}
	if phone != "" {
		u.Phone = &phone
	}
	inviteProfile, err := s.resolveRegistrationInviteProfile(ctx, req.RegistrationHost)
	if err != nil {
		if isRegistrationInviteConfigErr(err) {
			return nil, errorx.ErrParamsError
		}
		return nil, errorx.ErrDbError
	}
	applyRegistrationInviteProfile(u, inviteProfile)
	if !inviteProfile.Apply {
		u.VipConfigID = 1
	}

	withInvite := inviteProfile.Apply
	if err := s.users.CreateRegister(ctx, nil, u, withInvite); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, errorx.ErrUserExists
		}
		if withInvite && isMissingSchemaErr(err) {
			fallbackErr := s.users.CreateRegister(ctx, nil, u, false)
			if fallbackErr == nil {
				logger.WarnZ(ctx, "users-register-invite-columns-missing-fallback",
					zap.String("registration_host", req.RegistrationHost),
					zap.Error(err))
				return s.finishRegistration(ctx, u)
			}
			err = fallbackErr
		}
		logger.ErrorZ(ctx, "users-register-create-failed", zap.Error(err))
		return nil, errorx.ErrDbError
	}
	return s.finishRegistration(ctx, u)
}

func (s *UserService) finishRegistration(ctx context.Context, u *entity.Users) (*response.LoginUserResp, error) {
	resp, err := s.loginUserResp(u)
	if err != nil {
		return nil, err
	}
	if s.coupons == nil {
		return resp, nil
	}
	n, grantErr := s.coupons.GrantRegisterCoupons(ctx, u.ID)
	if grantErr != nil {
		logger.WarnZ(ctx, "register-coupon-grant-failed", zap.Error(grantErr))
		return resp, nil
	}
	if n > 0 {
		resp.RegisterCouponsGranted = n
	}
	return resp, nil
}

func (s *UserService) Login(ctx context.Context, req *request.LoginUserReq) (*response.LoginUserResp, error) {
	email := strings.TrimSpace(req.Email)
	phone := strings.TrimSpace(req.Phone)
	if err := validateLoginIdentity(email, phone); err != nil {
		return nil, err
	}

	var u *entity.Users
	var err error
	if email != "" {
		u, err = s.users.FindByEmail(ctx, email)
	} else {
		u, err = s.users.FindByPhone(ctx, phone)
	}
	if err != nil {
		return nil, errorx.ErrDbError
	}
	if u == nil {
		return nil, errorx.ErrUserNotFound
	}
	if u.Status != "active" {
		return nil, errorx.ErrUserDisabled
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errorx.ErrWrongPassword
	}
	_ = s.users.UpdateLastLogin(ctx, u.ID)
	return s.loginUserResp(u)
}

func (s *UserService) loginUserResp(u *entity.Users) (*response.LoginUserResp, error) {
	token, err := auth.IssueUserToken(u.ID, auth.SessionTokenSecret, auth.SessionTokenTTL)
	if err != nil {
		return nil, errorx.ErrUnknown
	}
	profile := toUserProfile(u)
	return &response.LoginUserResp{
		Token: token,
		User:  *profile,
	}, nil
}

func validateLoginIdentity(email, phone string) error {
	if email == "" && phone == "" {
		return errorx.ErrParamsError
	}
	if email != "" && phone != "" {
		return errorx.ErrParamsError
	}
	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return errorx.ErrInvalidCredential
		}
		return nil
	}
	if !cnPhonePattern.MatchString(phone) {
		return errorx.ErrInvalidCredential
	}
	return nil
}

func toUserProfile(u *entity.Users) *response.UserProfileResp {
	resp := &response.UserProfileResp{
		ID:                u.ID,
		Status:            u.Status,
		WalletBalance:     response.MoneyFrom(u.WalletBalance),
		CommissionBalance: response.MoneyFrom(u.CommissionBalance),
		CreatedAt:         formatUserDateTime(u.CreatedAt),
	}
	if u.Email != nil {
		resp.Email = *u.Email
	}
	if u.Phone != nil {
		resp.Phone = *u.Phone
	}
	if u.Nickname != nil {
		resp.Nickname = *u.Nickname
	}
	if u.LastLoginAt != nil && !u.LastLoginAt.IsZero() {
		resp.LastLoginAt = formatUserDateTime(*u.LastLoginAt)
	}
	return resp
}

func formatUserDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}
