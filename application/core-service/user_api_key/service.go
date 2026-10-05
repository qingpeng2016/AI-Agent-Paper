package user_api_key

import (
	"context"
	"errors"
	"time"

	"github.com/qingpeng2016/ai-token-mall/application/core-service/order"
	"github.com/qingpeng2016/ai-token-mall/common/apikey"
	"github.com/qingpeng2016/ai-token-mall/common/constants"
	"github.com/qingpeng2016/ai-token-mall/common/errorx"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
	"gorm.io/gorm"
)

type Service struct {
	tx              repository.Transactor
	apiKeys         repository.UserAPIKeysRepo
	subs            repository.UserSubscriptionsRepo
	users           repository.UsersRepo
	enterpriseUsers repository.EnterpriseUsersRepo
	inquiries       repository.EnterpriseInquiryRepo
}

func NewService(
	tx repository.Transactor,
	apiKeys repository.UserAPIKeysRepo,
	subs repository.UserSubscriptionsRepo,
	users repository.UsersRepo,
	enterpriseUsers repository.EnterpriseUsersRepo,
	inquiries repository.EnterpriseInquiryRepo,
) *Service {
	return &Service{
		tx:              tx,
		apiKeys:         apiKeys,
		subs:            subs,
		users:           users,
		enterpriseUsers: enterpriseUsers,
		inquiries:       inquiries,
	}
}

func (s *Service) ListMainKeys(ctx context.Context, userID uint) ([]response.UserAPIKeyItem, error) {
	rows, err := s.apiKeys.ListMainByUserID(ctx, userID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	names := s.subscriptionNamesForKeyRows(ctx, rows)
	items := make([]response.UserAPIKeyItem, 0, len(rows))
	for _, row := range rows {
		name := names[row.UserSubscriptionID]
		if name == "" {
			name = "团队套餐"
		}
		items = append(items, mapKeyItem(&row, name, "", ""))
	}
	return items, nil
}

func (s *Service) ListTeamSubKeys(ctx context.Context, ownerUserID uint) ([]response.UserAPIKeyItem, error) {
	rows, err := s.apiKeys.ListSubByOwnerUserID(ctx, ownerUserID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	names, err := s.subscriptionNameMap(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}
	items := make([]response.UserAPIKeyItem, 0, len(rows))
	for _, row := range rows {
		nick, email := s.memberLabels(ctx, row.UserID)
		items = append(items, mapKeyItem(&row, names[row.UserSubscriptionID], nick, email))
	}
	return items, nil
}

func (s *Service) CreateSubKey(ctx context.Context, ownerUserID uint, req *request.CreateSubAPIKeyReq) (*response.UserAPIKeyItem, error) {
	var created *entity.UserAPIKeys
	var createdPlain string
	err := s.tx.Transaction(ctx, func(tx *gorm.DB) error {
		sub, err := s.subs.FindOne(ctx, tx, map[string]interface{}{
			"id":      req.UserSubscriptionID,
			"user_id": ownerUserID,
		}, "")
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errorx.ErrRenewNoSubscription
			}
			return errorx.ErrDbError
		}
		if sub.Status != "active" {
			return errorx.ErrRenewNoSubscription
		}

		member, err := s.enterpriseUsers.FindActiveByLinkedUserID(ctx, req.MemberUserID)
		if err != nil {
			return errorx.ErrDbError
		}
		if member == nil || member.OwnerUserID != ownerUserID {
			return errorx.ErrAPIKeyForbidden
		}

		dup, err := s.apiKeys.ExistsActiveSubKey(ctx, req.UserSubscriptionID, req.MemberUserID)
		if err != nil {
			return errorx.ErrDbError
		}
		if dup {
			return errorx.ErrSubKeyDuplicate
		}

		sumSub, err := s.apiKeys.SumActiveSubKeyLimitTokensBySubscription(ctx, req.UserSubscriptionID, 0)
		if err != nil {
			return errorx.ErrDbError
		}
		if sumSub+req.LimitTokens > sub.LimitTokens {
			return errorx.ErrSubKeyLimitExceeded
		}

		ownerUID, inquiryID, err := order.MainAPIKeyEnterpriseFields(ctx, tx, ownerUserID)
		if err != nil {
			return err
		}

		plain, hash, err := apikey.Generate()
		if err != nil {
			return err
		}
		createdPlain = plain
		now := time.Now()
		row := entity.UserAPIKeys{
			UserID:               req.MemberUserID,
			OwnerUserID:          ownerUID,
			EnterpriseInquiryID:  inquiryID,
			UserSubscriptionID:   sub.ID,
			KeyType:              constants.APIKeyTypeSub,
			KeyHash:              hash,
			ProductsCategoryName: sub.ProductsCategoryName,
			LimitTokens:          req.LimitTokens,
			UsedTokens:           0,
			Status:               "active",
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if err := s.apiKeys.Create(ctx, tx, &row); err != nil {
			return errorx.ErrDbError
		}
		created = &row
		return nil
	})
	if err != nil {
		return nil, err
	}
	names, _ := s.subscriptionNameMap(ctx, ownerUserID)
	nick, email := s.memberLabels(ctx, created.UserID)
	item := mapKeyItem(created, names[created.UserSubscriptionID], nick, email)
	item.APIKey = createdPlain
	return &item, nil
}

func (s *Service) UpdateSubKeyLimit(ctx context.Context, ownerUserID, keyID uint, req *request.UpdateSubAPIKeyLimitReq) error {
	return s.tx.Transaction(ctx, func(tx *gorm.DB) error {
		row, err := s.apiKeys.FindByID(ctx, keyID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errorx.ErrAPIKeyNotFound
			}
			return errorx.ErrDbError
		}
		if row.OwnerUserID != ownerUserID || row.KeyType != constants.APIKeyTypeSub {
			return errorx.ErrAPIKeyForbidden
		}
		if req.LimitTokens < row.UsedTokens {
			return errorx.ErrParamsError
		}
		sub, err := s.subs.FindOne(ctx, tx, map[string]interface{}{"id": row.UserSubscriptionID}, "")
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errorx.ErrRenewNoSubscription
			}
			return errorx.ErrDbError
		}
		sumSub, err := s.apiKeys.SumActiveSubKeyLimitTokensBySubscription(ctx, row.UserSubscriptionID, row.ID)
		if err != nil {
			return errorx.ErrDbError
		}
		if sumSub+req.LimitTokens > sub.LimitTokens {
			return errorx.ErrSubKeyLimitExceeded
		}
		now := time.Now()
		return s.apiKeys.UpdateWhere(ctx, tx, map[string]interface{}{"id": keyID}, map[string]interface{}{
			"limit_tokens": req.LimitTokens,
			"updated_at":   now,
		})
	})
}

func (s *Service) ListTeamMembers(ctx context.Context, ownerUserID uint) (*response.ApiTeamMemberListResp, error) {
	rows, err := s.enterpriseUsers.ListByOwnerUserID(ctx, ownerUserID, 0, 500)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	items := make([]response.ApiTeamMemberItem, 0, len(rows))
	for _, row := range rows {
		item := response.ApiTeamMemberItem{
			ID:     row.ID,
			Status: row.Status,
		}
		if row.UserID != nil {
			item.UserID = *row.UserID
			nick, email := s.memberLabels(ctx, *row.UserID)
			item.Nickname = nick
			item.Email = email
			if item.Nickname == "" {
				item.Nickname = row.MemberName
			}
		} else {
			item.Nickname = row.MemberName
		}
		items = append(items, item)
	}
	return &response.ApiTeamMemberListResp{Items: items}, nil
}

func (s *Service) ListAddableInvitees(ctx context.Context, ownerUserID uint) (*response.ApiTeamAddableListResp, error) {
	linked, err := s.enterpriseUsers.LinkedUserIDsByOwner(ctx, ownerUserID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	inTeam := make(map[uint]struct{}, len(linked))
	for _, id := range linked {
		inTeam[id] = struct{}{}
	}
	invitees, err := s.users.ListByParentUserID(ctx, ownerUserID, 0, 500)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	items := make([]response.ApiTeamAddableInviteeItem, 0)
	for _, u := range invitees {
		if _, ok := inTeam[u.ID]; ok {
			continue
		}
		item := response.ApiTeamAddableInviteeItem{
			UserID:       u.ID,
			RegisteredAt: u.CreatedAt.Format("2006-01-02 15:04"),
		}
		if u.Nickname != nil {
			item.Nickname = *u.Nickname
		}
		if u.Email != nil {
			item.Email = *u.Email
		}
		if item.Nickname == "" && item.Email != "" {
			item.Nickname = item.Email
		}
		items = append(items, item)
	}
	return &response.ApiTeamAddableListResp{Items: items}, nil
}

func (s *Service) AddTeamMember(ctx context.Context, ownerUserID uint, req *request.AddApiTeamMemberReq) error {
	invitee, err := s.users.FindByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorx.ErrUserNotFound
		}
		return errorx.ErrDbError
	}
	if invitee.ParentUserID == 0 || invitee.ParentUserID != ownerUserID {
		return errorx.ErrTeamMemberNotInvited
	}
	existing, err := s.enterpriseUsers.FindActiveByLinkedUserID(ctx, req.UserID)
	if err != nil {
		return errorx.ErrDbError
	}
	if existing != nil {
		if existing.OwnerUserID == ownerUserID {
			return errorx.ErrTeamMemberExists
		}
		return errorx.ErrTeamMemberExists
	}
	inquiryID, err := s.latestEnterpriseInquiryID(ctx, ownerUserID)
	if err != nil {
		return err
	}
	name := ""
	if invitee.Nickname != nil {
		name = *invitee.Nickname
	}
	if name == "" && invitee.Email != nil {
		name = *invitee.Email
	}
	if name == "" {
		name = "团队成员"
	}
	uid := req.UserID
	now := time.Now()
	row := &entity.EnterpriseUsers{
		OwnerUserID:         ownerUserID,
		EnterpriseInquiryID: inquiryID,
		UserID:              &uid,
		MemberName:          name,
		Email:               invitee.Email,
		Phone:               invitee.Phone,
		Role:                "member",
		Status:              "active",
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	return s.enterpriseUsers.Create(ctx, nil, row)
}

func (s *Service) subscriptionNameMap(ctx context.Context, userID uint) (map[uint]string, error) {
	rows, err := s.subs.ListByUserID(ctx, userID)
	if err != nil {
		return nil, errorx.ErrDbError
	}
	return subscriptionNamesFromSubscriptionRows(rows), nil
}

func subscriptionDisplayName(productCardTitle, skuProductName string) string {
	if productCardTitle != "" {
		return productCardTitle
	}
	if skuProductName != "" {
		return skuProductName
	}
	return "团队套餐"
}

func subscriptionNamesFromSubscriptionRows(rows []repository.UserSubscriptionsListRow) map[uint]string {
	out := make(map[uint]string, len(rows))
	for _, row := range rows {
		out[row.ID] = subscriptionDisplayName(row.ProductCardTitle, row.SKUProductName)
	}
	return out
}

func (s *Service) subscriptionNamesForKeyRows(ctx context.Context, keyRows []entity.UserAPIKeys) map[uint]string {
	out := make(map[uint]string)
	for _, k := range keyRows {
		if _, ok := out[k.UserSubscriptionID]; ok {
			continue
		}
		sub, err := s.subs.FindOne(ctx, nil, map[string]interface{}{"id": k.UserSubscriptionID}, "")
		if err != nil {
			out[k.UserSubscriptionID] = "团队套餐"
			continue
		}
		out[k.UserSubscriptionID] = subscriptionDisplayName("", sub.SKUProductName)
	}
	return out
}

func (s *Service) memberLabels(ctx context.Context, userID uint) (nickname, email string) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil || u == nil {
		return "", ""
	}
	if u.Nickname != nil {
		nickname = *u.Nickname
	}
	if u.Email != nil {
		email = *u.Email
	}
	if nickname == "" {
		nickname = email
	}
	return nickname, email
}

func mapKeyItem(row *entity.UserAPIKeys, subName, memberNick, memberEmail string) response.UserAPIKeyItem {
	active := row.Status == "active"
	_ = active
	return response.UserAPIKeyItem{
		ID:                 row.ID,
		KeyType:            row.KeyType,
		KeyMasked:          apikey.MaskedFromHash(row.KeyHash),
		Status:             row.Status,
		UserSubscriptionID: row.UserSubscriptionID,
		SubscriptionName:   subName,
		LimitTokens:        row.LimitTokens,
		UsedTokens:         row.UsedTokens,
		MemberUserID:       row.UserID,
		MemberNickname:     memberNick,
		MemberEmail:        memberEmail,
	}
}

