package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
)

type EnterpriseInquiryRepo interface {
	Create(ctx context.Context, row *entity.EnterpriseInquiry) error
	FindLatestByOwnerUserID(ctx context.Context, ownerUserID uint) (*entity.EnterpriseInquiry, error)
}
