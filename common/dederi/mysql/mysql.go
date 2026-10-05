package mysqlclient

import (
	"fmt"
	"github.com/qingpeng2016/ai-token-mall/conf"
	"gorm.io/driver/mysql"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"time"

	"gorm.io/gorm"
)

// NewDB 创建到mysql的链接
func NewDB(source *conf.Mysql, replicas ...*conf.Mysql) *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Asia%%2FShanghai",
			source.User,
			source.Password,
			source.Host,
			source.DbName), // DSN data source name
	}), &gorm.Config{
		Logger: &mysqlLogger{},
	})
	if err != nil {
		panic(fmt.Errorf("unable to connect to msyql, error is %s", err))
	}
	if err := db.Exec("SET time_zone = '+08:00'").Error; err != nil {
		panic(fmt.Errorf("unable to set mysql session time_zone: %s", err))
	}
	if source.LogMode {
		db.Debug()
	} else {
		// 当log_mode为false时，使用静默日志
		db.Logger = logger.Default.LogMode(logger.Silent)
	}
	if len(replicas) > 0 {
		dbResolver := new(dbresolver.DBResolver)
		for _, replica := range replicas {
			if replica == nil {
				continue
			}
			dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Asia%%2FShanghai",
				replica.User,
				replica.Password,
				replica.Host,
				replica.DbName)
			dbResolver = dbResolver.Register(dbresolver.Config{
				Replicas: []gorm.Dialector{mysql.Open(dsn)},
				// sources/replicas load balancing policy
				Policy: dbresolver.RandomPolicy{},
				// print sources/replicas mode in logger
				TraceResolverMode: true,
			})
		}
		if err = db.Use(dbResolver); err != nil {
			panic(fmt.Errorf("unable use dbResolver, error is %s", err.Error()))
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Errorf("unable to connect to DB, error is %s", err))
	}
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(source.MaxIdleConn)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(source.MaxOpenConn)
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db
}
