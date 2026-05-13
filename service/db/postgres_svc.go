package db

import (
	"fmt"

	"hecc-blot/contract/db"
	"hecc-blot/contract/log"
	dbConf "hecc-blot/entity/config/db"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresSvc struct {
	BaseDbSvc
}

func newPostgresSvc(config *dbConf.PostgresConfig, logger log.ILog) (db.IDb, func(), error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable connect_timeout=%d",
		config.Ip,
		config.Username,
		config.Password,
		config.DbName,
		config.Port,
		config.ConnectTimeout,
	)

	postgresDb, err := gorm.Open(postgres.Open(dsn), initGormConfig(logger, config.SlowThreshold))
	if err != nil {
		return nil, func() {}, err
	}

	sqlDb, err := postgresDb.DB()
	if err != nil {
		return nil, func() {}, err
	}

	setSqlDbPool(sqlDb, config.MaxIdleConn, config.MaxOpenConn, config.ConnMaxLifetime)

	return &PostgresSvc{
			BaseDbSvc{db: postgresDb},
		}, func() {
			sqlDb.Close()
		}, nil
}
