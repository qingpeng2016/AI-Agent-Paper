package boot

import (
	"github.com/qingpeng2016/ai-agent-paper/infrastructure/mysql"
	"gorm.io/gorm"
)

func NewDBClient() *gorm.DB {
	return mysql.GetDBClient()
}
