package db

import (
	"fmt"

	"github.com/bestHeCC/hecc-core/contract/db"
	"github.com/bestHeCC/hecc-core/contract/log"
	dbConf "github.com/bestHeCC/hecc-core/entity/config/db"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MysqlSvc struct {
	BaseDbSvc
}

func newMysqlSvc(config *dbConf.MysqlConfig, logger log.ILog) (db.IDb, func(), error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&timeout=%ds",
		config.Username,
		config.Password,
		config.Ip,
		config.Port,
		config.DbName,
		config.ConnectTimeout,
	)

	mysqlDb, err := gorm.Open(mysql.Open(dsn), initGormConfig(logger, config.SlowThreshold))
	if err != nil {
		return nil, func() {}, err
	}

	sqlDb, err := mysqlDb.DB()
	if err != nil {
		return nil, func() {}, err
	}

	setSqlDbPool(sqlDb, config.MaxIdleConn, config.MaxOpenConn, config.ConnMaxLifetime)

	return &MysqlSvc{
			BaseDbSvc{db: mysqlDb},
		}, func() {
			sqlDb.Close()
		}, nil
}
