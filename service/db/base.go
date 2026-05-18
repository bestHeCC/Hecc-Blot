package db

import (
	"context"
	"database/sql"
	"time"

	"hecc-blot/contract/db"
	"hecc-blot/contract/log"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// BaseDbSvc 数据库服务基类
type BaseDbSvc struct {
	ctx   context.Context
	db    *gorm.DB
	model db.IDbModel
}

// Begin 开启事务，返回新的 BaseDbSvc 实例，原始实例不受影响。
// Commit/Rollback 只在返回的实例上调用才有效。
func (b *BaseDbSvc) Begin() db.IDb {
	gl := b.db.Statement.Logger.(logger.Interface)
	txDB := b.db.Begin()
	gl.Info(b.ctx, "transaction started")
	return &BaseDbSvc{
		ctx: b.ctx,
		db:  txDB,
	}
}

func (b *BaseDbSvc) Rollback() {
	gl := b.db.Statement.Logger.(logger.Interface)
	gl.Warn(b.ctx, "transaction rollback")
	b.db.Rollback()
}

func (b *BaseDbSvc) Commit() error {
	gl := b.db.Statement.Logger.(logger.Interface)
	gl.Info(b.ctx, "transaction committed")
	return b.db.Commit().Error
}

// Add 添加记录
func (b *BaseDbSvc) Add(entry db.IDbModel) error {
	defer b.reset()
	return b.db.Create(entry).Error
}

// Remove 删除记录
func (b *BaseDbSvc) Remove(entry db.IDbModel) error {
	defer b.reset()
	return b.db.Delete(&entry).Error
}

// Query 查询
func (b *BaseDbSvc) Query(entry db.IDbModel) db.IDb {
	b.db = b.db.Model(&entry)
	return b
}

// Save 保存记录
func (b *BaseDbSvc) Save(entry db.IDbModel) error {
	return b.db.Updates(entry).Error
}

// Count 统计数量
func (b *BaseDbSvc) Count() (int64, error) {
	defer b.reset()
	var count int64
	err := b.db.Count(&count).Error
	return count, err
}

// Order 排序
func (b *BaseDbSvc) Order(fields ...string) db.IDb {
	b.db = b.db.Order(fields)
	return b
}

// Select 选择字段
func (b *BaseDbSvc) Select(args ...interface{}) db.IDb {
	b.db = b.db.Select(args[0], args[1:]...)
	return b
}

// Offset 偏移
func (b *BaseDbSvc) Offset(v int) db.IDb {
	b.db = b.db.Offset(v)
	return b
}

// Limit 限制
func (b *BaseDbSvc) Limit(v int) db.IDb {
	b.db = b.db.Limit(v)
	return b
}

// Where 条件
func (b *BaseDbSvc) Where(args ...interface{}) db.IDb {
	b.db = b.db.Where(args[0], args[1:]...)
	return b
}

// Take 获取一条
func (b *BaseDbSvc) Take(dst interface{}) error {
	defer b.reset()
	return b.db.Take(dst).Error
}

// Find 查询多条
func (b *BaseDbSvc) Find(dst interface{}) error {
	defer b.reset()
	return b.db.Find(dst).Error
}

// WithContext 设置上下文
func (b *BaseDbSvc) WithContext(ctx context.Context) {
	b.ctx = ctx
	b.db = b.db.WithContext(ctx)
}

// reset 重置 GORM 会话，避免污染
func (b *BaseDbSvc) reset() {
	b.db = b.db.Session(&gorm.Session{NewDB: true})
}

// initGormConfig 初始化 GORM 通用配置
func initGormConfig(logger log.ILog, slowThreshold int) *gorm.Config {
	return &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   newILogGormLogger(logger, slowThreshold),
	}
}

// setSqlDbPool 设置数据库连接池配置
func setSqlDbPool(sqlDb *sql.DB, maxIdleConn, maxOpenConn, connMaxLifetime int) {
	sqlDb.SetMaxIdleConns(maxIdleConn)
	sqlDb.SetMaxOpenConns(maxOpenConn)
	sqlDb.SetConnMaxLifetime(time.Second * time.Duration(connMaxLifetime))
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
		slowThreshold: time.Duration(slowThreshold) * time.Millisecond,
	}
}
