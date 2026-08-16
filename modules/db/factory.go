package db

import (
	"context"
	"fmt"

	dbContract "github.com/bestHeCC/hecc-db/contract"
	"github.com/bestHeCC/hecc-log/contract"
	dbConf "github.com/bestHeCC/hecc-db/config"
	dbEnum "github.com/bestHeCC/hecc-db/enum/db"
	"github.com/bestHeCC/hecc-core/util"

	"gorm.io/gorm"
)

type Factory struct {
	db        map[dbEnum.Value]dbContract.IDb
	defaultDb dbEnum.Value
}

func (f *Factory) Build(ctx context.Context, v ...dbEnum.Value) dbContract.IDb {
	t := f.defaultDb
	if len(v) > 0 {
		t = v[0]
	}

	dbSvc, ok := f.db[t]
	if !ok {
		panic(fmt.Sprintf("无效db类型:%v", v))
	}

	// 通过 GetInstance() 获取 GORM 实例，创建独立副本
	// 先提取真实 context（*gin.Context → Request.Context()），否则 GORM otel 插件
	// 读不到父 span，SQL 会变成独立 trace
	ctx = util.ExtractContext(ctx)
	clone := &BaseDbSvc{
		ctx: ctx,
		db:  dbSvc.GetInstance().(*gorm.DB).WithContext(ctx),
	}
	return clone
}

func (f *Factory) SetDefault(t dbEnum.Value) {
	if _, ok := f.db[t]; !ok {
		panic(fmt.Sprintf("无效db类型:%v", t))
	}
	f.defaultDb = t
}

func NewDbFactory(config *dbConf.Config, logger log.ILog) (dbContract.IDbFactory, func(), error) {
	f := Factory{
		db:        make(map[dbEnum.Value]dbContract.IDb),
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
