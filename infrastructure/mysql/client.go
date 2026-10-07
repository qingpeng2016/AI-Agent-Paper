package mysql

import (
	mysqlclient "github.com/qingpeng2016/ai-agent-paper/common/dederi/mysql"
	"github.com/qingpeng2016/ai-agent-paper/conf"
	"gorm.io/gorm"
	"sync"
)

var db *gorm.DB
var once sync.Once

func GetDBClient() *gorm.DB {
	once.Do(func() {
		db = mysqlclient.NewDB(conf.GetMysqlMasterConf(), conf.GetMysqlReplicaConf())
	})
	return db
}
