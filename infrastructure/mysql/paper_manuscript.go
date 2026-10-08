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

func (r *PaperManuscriptImpl) Create(ctx context.Context, userID uint, title string) (*entity.PaperManuscript, error) {
	row := entity.PaperManuscript{
		UserID: userID,
		Title:  title,
		Status: "active",
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}
