package mysqlclient

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

// IsDuplicateError is duplicate error
func IsDuplicateError(err error) bool {
	if mysqlErr, ok := err.(*mysql.MySQLError); !ok {
		return false
	} else if mysqlErr.Number == 1062 {
		return true
	}
	return false
}

// IsDeadlockError 识别 MySQL 1213 死锁（可安全重试）。
func IsDeadlockError(err error) bool {
	if err == nil {
		return false
	}
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1213
}
