package db

import (
	"context"

	"core/contract/log"
	dbConf "core/entity/config/db"

	"fmt"

	"core/contract/db"
	dbEnum "core/enum/db"
)

type Factory struct {
	db map[dbEnum.Value]db.IDb
}

func (f Factory) Build(ctx context.Context, v ...dbEnum.Value) db.IDb {
	t := dbEnum.Mysql
	if len(v) > 0 {
		t = v[0]
	}

	dbSvc, ok := f.db[t]
	if !ok {
		panic(fmt.Sprintf("无效db类型:%v", v))
	}

	dbSvc.WithContext(ctx)
	return dbSvc
}

func NewDbFactory(config *dbConf.Config, logger log.ILog) (db.IDbFactory, func(), error) {
	mysql, clearUp, err := newMysqlSvc(&config.Mysql, logger)
	if err != nil {
		return nil, func() {}, err
	}

	return Factory{
		db: map[dbEnum.Value]db.IDb{
			dbEnum.Mysql: mysql,
		},
	}, clearUp, nil
}
