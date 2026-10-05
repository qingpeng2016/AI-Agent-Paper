package mysql

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type TransactorImpl struct {
	db *gorm.DB
}

func NewTransactorImpl(db *gorm.DB) repository.Transactor {
	return &TransactorImpl{db: db}
}

func (t *TransactorImpl) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return t.db.WithContext(ctx).Transaction(fn)
}
