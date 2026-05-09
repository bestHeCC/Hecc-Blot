package db

import (
	"context"

	"core/contract/log"
	"core/entity/config"

	"fmt"
	"time"

	"core/contract/db"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type MysqlSvc struct {
	ctx   context.Context
	db    *gorm.DB
	model db.IDbModel
}

func (m MysqlSvc) Add(entry db.IDbModel) error {
	return m.db.Create(entry).Error
}

func (m MysqlSvc) Remove(entry db.IDbModel) error {
	return m.db.Delete(&entry).Error
}

func (m MysqlSvc) Query(entry db.IDbModel) db.IDb {
	m.db = m.db.Model(&entry)
	return m
}

func (m MysqlSvc) Save(entry db.IDbModel) error {
	return m.db.Updates(entry).Error
}

func (m MysqlSvc) Count() (int64, error) {
	var count int64
	err := m.db.Count(&count).Error
	return count, err
}

func (m MysqlSvc) Order(fields ...string) db.IDb {
	m.db = m.db.Order(fields)
	return m
}

func (m MysqlSvc) Select(args ...interface{}) db.IDb {
	m.db = m.db.Select(args[0], args[1:]...)
	return m
}

func (m MysqlSvc) Offset(v int) db.IDb {
	m.db = m.db.Offset(v)
	return m
}

func (m MysqlSvc) Limit(v int) db.IDb {
	m.db = m.db.Limit(v)
	return m
}

func (m MysqlSvc) Where(args ...interface{}) db.IDb {
	m.db = m.db.Where(args[0], args[1:]...)
	return m
}

func (m MysqlSvc) Take(dst interface{}) error {
	return m.db.Take(dst).Error
}

func (m MysqlSvc) Find(dst interface{}) error {
	return m.db.Find(dst).Error
}

func (m MysqlSvc) WithContext(ctx context.Context) {
	m.ctx = ctx
}

func newMysqlSvc(config *config.Config, logger log.ILog) (db.IDb, func(), error) {
	// 配置 GORM
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&timeout=%ds",
		config.Db.Mysql.Username,
		config.Db.Mysql.Password,
		config.Db.Mysql.Ip,
		config.Db.Mysql.Port,
		config.Db.Mysql.DbName,
		config.Db.Mysql.ConnectTimeout,
	)

	mysqlDb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,                                                     // 禁用自动创建外键约束
		Logger:                                   newILogGormLogger(logger, config.Db.Mysql.SlowThreshold), // 使用 ILog 适配器
	})
	if err != nil {
		return nil, func() {}, err
	}

	sqlDb, err := mysqlDb.DB()
	if err != nil {
		return nil, func() {}, err
	}

	// 设置空闲连接池中链接的最大数量
	sqlDb.SetMaxIdleConns(config.Db.Mysql.MaxIdleConn)
	// 设置打开数据库链接的最大数量
	sqlDb.SetMaxOpenConns(config.Db.Mysql.MaxOpenConn)
	// 设置链接可复用的最大时间
	sqlDb.SetConnMaxLifetime(time.Second * time.Duration(config.Db.Mysql.ConnMaxLifetime))

	return MysqlSvc{
			db: mysqlDb,
		}, func() {
			sqlDb.Close()
		}, nil
}

// iLogGormLogger 是 log.ILog 到 GORM logger.Interface 的适配器
type iLogGormLogger struct {
	logger        log.ILog
	slowThreshold time.Duration
}

func (gl *iLogGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return gl
}

func (gl *iLogGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	gl.logger.Info(ctx, msg, data...)
}

func (gl *iLogGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	gl.logger.Warn(ctx, msg, data...)
}

func (gl *iLogGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	gl.logger.Error(ctx, msg, data...)
}

func (gl *iLogGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	switch {
	case err != nil:
		gl.logger.Error(ctx, "SQL Trace",
			"error", err,
			"elapsed", elapsed.String(),
			"rows", rows,
			"sql", sql)
	case elapsed > gl.slowThreshold && gl.slowThreshold > 0:
		gl.logger.Warn(ctx, "Slow SQL",
			"elapsed", elapsed.String(),
			"rows", rows,
			"sql", sql)
	default:
		gl.logger.Info(ctx, "SQL Trace",
			"elapsed", elapsed.String(),
			"rows", rows,
			"sql", sql)
	}
}

// newILogGormLogger 创建基于 ILog 的 GORM Logger
func newILogGormLogger(logger log.ILog, slowThreshold int) logger.Interface {
	return &iLogGormLogger{
		logger:        logger,
		slowThreshold: time.Duration(slowThreshold) * time.Millisecond, // 默认慢查询阈值
	}
}
