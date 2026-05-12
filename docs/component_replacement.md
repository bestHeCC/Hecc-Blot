# 组件替换说明

## 概述

Hecc-Go-Core 框架采用**面向接口编程**的设计理念，所有组件都通过接口契约进行交互。这种设计使得各个组件可以方便地进行替换，而不影响整体架构。

## 替换原则

### 1. 遵循接口契约

替换组件时，必须实现对应的接口：

```go
// 假设要替换日志组件
type MyCustomLog struct{}

// 必须实现 ILog 接口的所有方法
func (m MyCustomLog) Debug(ctx context.Context, msg string, fields ...interface{}) {
    // 自定义实现
}

func (m MyCustomLog) Error(ctx context.Context, msg string, fields ...interface{}) {
    // 自定义实现
}

func (m MyCustomLog) Info(ctx context.Context, msg string, fields ...interface{}) {
    // 自定义实现
}

func (m MyCustomLog) Warn(ctx context.Context, msg string, fields ...interface{}) {
    // 自定义实现
}
```

### 2. 注册到IOC容器

实现接口后，通过 IOC 容器注册即可替换默认实现：

```go
// 创建自定义实现
myLog := &MyCustomLog{}

// 注册到IOC容器，覆盖默认实现
ioc.Set(new(iCoreLog.ILog), myLog)
```

---

## 各组件替换示例

### 1. 替换日志组件

**需求**: 将默认的 zap 日志替换为 logrus

**步骤**:

```go
// 1. 创建自定义日志服务，实现 ILog 接口
type LogrusLogSvc struct {
    logger *logrus.Logger
}

func NewLogrusLogSvc() iCoreLog.ILog {
    logger := logrus.New()
    logger.SetFormatter(&logrus.JSONFormatter{})
    logger.SetLevel(logrus.InfoLevel)
    return &LogrusLogSvc{logger: logger}
}

func (l LogrusLogSvc) Debug(ctx context.Context, msg string, fields ...interface{}) {
    l.logger.Debug(msg, fields...)
}

func (l LogrusLogSvc) Error(ctx context.Context, msg string, fields ...interface{}) {
    l.logger.Error(msg, fields...)
}

func (l LogrusLogSvc) Info(ctx context.Context, msg string, fields ...interface{}) {
    l.logger.Info(msg, fields...)
}

func (l LogrusLogSvc) Warn(ctx context.Context, msg string, fields ...interface{}) {
    l.logger.Warn(msg, fields...)
}

// 2. 在 main 函数中注册
func main() {
    // 使用自定义日志服务
    logSvc := NewLogrusLogSvc()
    
    // 注册到IOC容器
    ioc.Set(new(iCoreLog.ILog), logSvc)
    
    // ... 其他初始化代码
}
```

### 2. 替换数据库组件

**需求**: 将 MySQL 替换为 PostgreSQL

**步骤**:

```go
// 1. 创建 PostgreSQL 实现，实现 IDb 接口
type PostgresSvc struct {
    ctx   context.Context
    db    *gorm.DB
}

func NewPostgresSvc(config *PostgresConfig) (iCoreDb.IDb, func(), error) {
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        config.Host, config.Port, config.Username, config.Password, config.DbName)
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, func() {}, err
    }
    
    sqlDb, _ := db.DB()
    return &PostgresSvc{db: db}, func() { sqlDb.Close() }, nil
}

// 实现 IDb 接口方法
func (p PostgresSvc) Add(entry iCoreDb.IDbModel) error {
    return p.db.Create(entry).Error
}

func (p PostgresSvc) Remove(entry iCoreDb.IDbModel) error {
    return p.db.Delete(entry).Error
}

// ... 其他接口方法

// 2. 自定义 Factory
type PostgresFactory struct {
    db iCoreDb.IDb
}

func (f PostgresFactory) Build(ctx context.Context, v ...iCoreDb.DbType) iCoreDb.IDb {
    f.db.WithContext(ctx)
    return f.db
}

// 3. 在 main 函数中注册
func main() {
    postgresDb, clearUp, err := NewPostgresSvc(&config.Postgres)
    if err != nil {
        panic(err)
    }
    defer clearUp()
    
    dbFactory := &PostgresFactory{db: postgresDb}
    ioc.Set(new(iCoreDb.IDbFactory), dbFactory)
    
    // ...
}
```

### 3. 替换缓存组件

**需求**: 将默认缓存替换为自定义缓存实现

**步骤**:

```go
// 1. 创建自定义缓存服务
type MyCacheSvc struct{}

func (m MyCacheSvc) Set(key string, value interface{}, expiration time.Duration) error {
    // 自定义缓存逻辑
    return nil
}

func (m MyCacheSvc) Get(key string) (interface{}, error) {
    // 自定义缓存逻辑
    return nil, nil
}

// ... 其他方法

// 2. 自定义 Factory
type MyCacheFactory struct{}

func (f MyCacheFactory) Local() iCoreCache.ILocalCache {
    return &MyCacheSvc{}
}

func (f MyCacheFactory) Redis() iCoreCache.IRedisCache {
    return &MyRedisSvc{}
}

// 3. 在 main 函数中注册
func main() {
    cacheFactory := &MyCacheFactory{}
    ioc.Set(new(iCoreCache.ICacheFactory), cacheFactory)
}
```

---

## 替换注意事项

### 1. 接口兼容性

- 必须实现接口的**所有方法**
- 方法签名必须完全匹配
- 返回值类型必须一致

### 2. 生命周期管理

- 如果组件需要资源清理（如数据库连接），应返回清理函数
- 在 main 函数中使用 `defer` 确保资源被正确释放

### 3. 配置兼容性

- 如果自定义组件需要新的配置字段，需要更新 `entity/config` 目录下的配置结构体
- 配置文件格式保持 YAML 格式

### 4. 测试验证

替换组件后，应编写相应的单元测试：

```go
func TestMyCustomLog(t *testing.T) {
    logSvc := NewLogrusLogSvc()
    
    // 验证接口实现
    var _ iCoreLog.ILog = logSvc
    
    // 测试功能
    logSvc.Info(context.Background(), "test message")
}
```

---

## 替换流程图

```
┌─────────────────────────────────────────────────────────────┐
│                      组件替换流程                            │
├─────────────────────────────────────────────────────────────┤
│  1. 定义自定义实现 (实现对应接口)                             │
│         ↓                                                   │
│  2. 创建实例 (可能需要配置参数)                               │
│         ↓                                                   │
│  3. 注册到 IOC 容器 (ioc.Set)                               │
│         ↓                                                   │
│  4. 框架自动注入到需要的地方                                 │
└─────────────────────────────────────────────────────────────┘
```

---

## 总结

框架的面向接口设计使得组件替换非常简单：

1. **实现接口** - 创建自定义实现类，实现对应接口的所有方法
2. **注册到IOC** - 通过 `ioc.Set()` 将实例注册到容器
3. **自动生效** - 框架会自动注入新的实现，无需修改其他代码

这种设计实现了**依赖倒置原则**，高层模块不依赖低层模块的具体实现，只依赖抽象接口。
