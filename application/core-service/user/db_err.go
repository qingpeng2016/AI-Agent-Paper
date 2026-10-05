package user

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

func isMissingSchemaErr(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		switch me.Number {
		case 1054, 1146: // unknown column / unknown table
			return true
		}
	}
	return false
}
