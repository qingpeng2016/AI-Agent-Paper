package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

type EnterpriseUsersRepo interface {
	Create(ctx context.Context, tx *gorm.DB, row *entity.EnterpriseUsers) error
	Update(ctx context.Context, tx *gorm.DB, row *entity.EnterpriseUsers) error
	FindByID(ctx context.Context, id uint) (*entity.EnterpriseUsers, error)
	FindByIDForOwner(ctx context.Context, ownerUserID, id uint) (*entity.EnterpriseUsers, error)
	ListByOwnerUserID(ctx context.Context, ownerUserID uint, offset, limit int) ([]entity.EnterpriseUsers, error)
	CountByOwnerUserID(ctx context.Context, ownerUserID uint) (int64, error)
	FindActiveByLinkedUserID(ctx context.Context, userID uint) (*entity.EnterpriseUsers, error)
	LinkedUserIDsByOwner(ctx context.Context, ownerUserID uint) ([]uint, error)
}
