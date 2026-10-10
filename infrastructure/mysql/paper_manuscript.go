package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/repository"
	"gorm.io/gorm"
)

type PaperManuscriptImpl struct {
	db *gorm.DB
}

func NewPaperManuscriptImpl(db *gorm.DB) repository.PaperManuscriptRepo {
	return &PaperManuscriptImpl{db: db}
}

func (r *PaperManuscriptImpl) GetByIDForUser(ctx context.Context, manuscriptID, userID uint) (*entity.PaperManuscript, error) {
	var row entity.PaperManuscript
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ? AND status = ?", manuscriptID, userID, "active").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperManuscriptImpl) ListByUser(ctx context.Context, userID uint) ([]entity.PaperManuscript, error) {
	var rows []entity.PaperManuscript
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, "active").
		Order("is_current DESC, updated_at DESC, id DESC").
		Find(&rows).Error
	return rows, err
}

func (r *PaperManuscriptImpl) Create(ctx context.Context, userID uint, title string, disciplineID *uint64, contentLanguage string) (*entity.PaperManuscript, error) {
	row := entity.PaperManuscript{
		UserID:          userID,
		Title:           title,
		DisciplineID:    disciplineID,
		ContentLanguage: contentLanguage,
		Status:          "active",
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PaperManuscriptImpl) SetCurrentForUser(ctx context.Context, userID, manuscriptID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var target entity.PaperManuscript
		if err := tx.Where("id = ? AND user_id = ? AND status = ?", manuscriptID, userID, "active").
			First(&target).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return gorm.ErrRecordNotFound
			}
			return err
		}
		if err := tx.Model(&entity.PaperManuscript{}).
			Where("user_id = ? AND status = ?", userID, "active").
			Update("is_current", false).Error; err != nil {
			return err
		}
		return tx.Model(&entity.PaperManuscript{}).
			Where("id = ? AND user_id = ?", manuscriptID, userID).
			Update("is_current", true).Error
	})
}
