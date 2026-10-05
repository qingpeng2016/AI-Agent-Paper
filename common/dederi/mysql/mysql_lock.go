package mysqlclient

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

// AcquireNamedLock 获取 MySQL 命名锁（跨连接/实例有效）。返回 true 表示获锁成功。
func AcquireNamedLock(ctx context.Context, db *gorm.DB, lockName string, timeoutSec int) (bool, error) {
	if db == nil {
		return false, nil
	}
	var got sql.NullInt64
	err := db.WithContext(ctx).Raw("SELECT GET_LOCK(?, ?)", lockName, timeoutSec).Scan(&got).Error
	if err != nil {
		return false, err
	}
	return got.Valid && got.Int64 == 1, nil
}

// ReleaseNamedLock 释放 MySQL 命名锁。
func ReleaseNamedLock(ctx context.Context, db *gorm.DB, lockName string) error {
	if db == nil {
		return nil
	}
	var released sql.NullInt64
	return db.WithContext(ctx).Raw("SELECT RELEASE_LOCK(?)", lockName).Scan(&released).Error
}
