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
	db        map[dbEnum.Value]db.IDb
	defaultDb dbEnum.Value
}

func (f *Factory) Build(ctx context.Context, v ...dbEnum.Value) db.IDb {
	t := f.defaultDb
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

func (f *Factory) SetDefault(t dbEnum.Value) {
	if _, ok := f.db[t]; !ok {
		panic(fmt.Sprintf("无效db类型:%v", t))
	}
	f.defaultDb = t
}

func NewDbFactory(config *dbConf.Config, logger log.ILog) (db.IDbFactory, func(), error) {
	f := Factory{
		db:        make(map[dbEnum.Value]db.IDb),
		defaultDb: dbEnum.Mysql, // 默认使用mysql
	}
	var cleanupFuncs []func()

	if config.Mysql.Ip != "" {
		mysql, cleanup, err := newMysqlSvc(&config.Mysql, logger)
		if err != nil {
			return nil, func() {}, err
		}
		f.db[dbEnum.Mysql] = mysql
		cleanupFuncs = append(cleanupFuncs, cleanup)
	}

	if config.Postgres.Ip != "" {
		postgres, cleanup, err := newPostgresSvc(&config.Postgres, logger)
		if err != nil {
			return nil, func() {}, err
		}
		f.db[dbEnum.Postgres] = postgres
		cleanupFuncs = append(cleanupFuncs, cleanup)
	}

	if len(f.db) == 0 {
		return nil, func() {}, fmt.Errorf("未配置任何数据库")
	}

	return &f, func() {
		for _, fn := range cleanupFuncs {
			fn()
		}
	}, nil
}
