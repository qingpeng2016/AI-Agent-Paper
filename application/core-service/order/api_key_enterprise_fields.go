package order

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"gorm.io/gorm"
)

// MainAPIKeyEnterpriseFields 新开套餐主 Key：填充 owner_user_id、enterprise_inquiry_id。
func MainAPIKeyEnterpriseFields(ctx context.Context, tx *gorm.DB, buyerUserID uint) (ownerUserID, enterpriseInquiryID uint, err error) {
	ownerUserID = buyerUserID
	if member, err := findActiveEnterpriseUserByLinkedUserID(ctx, tx, buyerUserID); err != nil {
		return 0, 0, err
	} else if member != nil {
		return member.OwnerUserID, member.EnterpriseInquiryID, nil
	}
	inquiryID, err := latestEnterpriseInquiryIDByOwner(ctx, tx, buyerUserID)
	if err != nil {
		return 0, 0, err
	}
	return ownerUserID, inquiryID, nil
}

func findActiveEnterpriseUserByLinkedUserID(ctx context.Context, tx *gorm.DB, userID uint) (*entity.EnterpriseUsers, error) {
	var row entity.EnterpriseUsers
	err := tx.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, "active").
		Order("id DESC").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func latestEnterpriseInquiryIDByOwner(ctx context.Context, tx *gorm.DB, ownerUserID uint) (uint, error) {
	var id uint
	err := tx.WithContext(ctx).Model(&entity.EnterpriseInquiry{}).
		Select("id").
		Where("owner_user_id = ?", ownerUserID).
		Order("id DESC").
		Limit(1).
		Scan(&id).Error
	return id, err
}
