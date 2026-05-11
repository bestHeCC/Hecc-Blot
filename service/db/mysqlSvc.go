package db

import (
	"context"

	"core/contract/log"
	dbConf "core/entity/config/db"

	"fmt"
	"time"

	"core/contract/db"

	"go.uber.org/zap"
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

func newMysqlSvc(config *dbConf.MysqlConfig, logger log.ILog) (db.IDb, func(), error) {
	// 配置 GORM
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&timeout=%ds",
		config.Username,
		config.Password,
		config.Ip,
		config.Port,
		config.DbName,
		config.ConnectTimeout,
	)

	mysqlDb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,                                            // 禁用自动创建外键约束
		Logger:                                   newILogGormLogger(logger, config.SlowThreshold), // 使用 ILog 适配器
	})
	if err != nil {
		return nil, func() {}, err
	}

	sqlDb, err := mysqlDb.DB()
	if err != nil {
		return nil, func() {}, err
	}

	// 设置空闲连接池中链接的最大数量
	sqlDb.SetMaxIdleConns(config.MaxIdleConn)
	// 设置打开数据库链接的最大数量
	sqlDb.SetMaxOpenConns(config.MaxOpenConn)
	// 设置链接可复用的最大时间
	sqlDb.SetConnMaxLifetime(time.Second * time.Duration(config.ConnMaxLifetime))

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
		gl.Error(ctx, "SQL Trace",
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql))
	case elapsed > gl.slowThreshold && gl.slowThreshold > 0:
		gl.Warn(ctx, "Slow SQL",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql))
	default:
		gl.Info(ctx, "SQL Trace",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql))
	}
}

// newILogGormLogger 创建基于 ILog 的 GORM Logger
func newILogGormLogger(logger log.ILog, slowThreshold int) logger.Interface {
	return &iLogGormLogger{
		logger:        logger,
		slowThreshold: time.Duration(slowThreshold) * time.Millisecond, // 默认慢查询阈值
	}
}
