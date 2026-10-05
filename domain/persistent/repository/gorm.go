package repository

import (
	"context"

	"gorm.io/gorm"
)

// GormDB 事务内传 tx，否则用 db。
func GormDB(ctx context.Context, db *gorm.DB, tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}
