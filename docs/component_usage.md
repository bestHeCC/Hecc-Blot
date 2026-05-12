# 组件使用说明

## 概述

Hecc-Go-Core 框架采用面向接口的设计，各个组件通过契约接口进行交互。下面详细说明各个核心组件的使用方法。

## 1. 日志组件 (Log)

### 接口定义
```go
type ILog interface {
    Debug(ctx context.Context, msg string, fields ...interface{})
    Error(ctx context.Context, msg string, fields ...interface{})
    Info(ctx context.Context, msg string, fields ...interface{})
    Warn(ctx context.Context, msg string, fields ...interface{})
}
```

### 使用方法

```go
// 创建日志服务
logSvc, err := log.NewLogger(&config.Log)
if err != nil {
    panic(err)
}

// 注册到IOC容器
ioc.Set(new(iCoreLog.ILog), logSvc)

// 使用日志服务
logSvc.Info(ctx, "user login", "user_id", 123)
logSvc.Error(ctx, "login failed", "error", err)
```

### 配置说明

日志组件支持两种实现：
- **本地日志**: 使用 zap + lumberjack 实现文件滚动
- **SLS日志**: 阿里云日志服务

```yaml
log:
  local:
    enable: true
    root_dir: ./logs
    max_size: 100
    max_backups: 30
    max_age: 7
    compress: true
  sls:
    enable: false
```

### 单测参考

查看 `service/log/logSvc_test.go` 和 `service/log/slsSvc_test.go` 了解测试方式。

---

## 2. 数据库组件 (DB)

### 接口定义

```go
type IDbFactory interface {
    Build(ctx context.Context, v ...dbEnum.Value) IDb
}

type IDb interface {
    Add(entry IDbModel) error
    Remove(entry IDbModel) error
    Query(entry IDbModel) IDb
    Save(entry IDbModel) error
    Count() (int64, error)
    Order(fields ...string) IDb
    Select(args ...interface{}) IDb
    Offset(v int) IDb
    Limit(v int) IDb
    Where(args ...interface{}) IDb
    Take(dst interface{}) error
    Find(dst interface{}) error
    WithContext(ctx context.Context)
}
```

### 使用方法

```go
// 创建数据库工厂
dbFactory, clearUp, err := db.NewDbFactory(&config.Db, logSvc)
if err != nil {
    panic(err)
}
defer clearUp()

// 注册到IOC容器
ioc.Set(new(iCoreDb.IDbFactory), dbFactory)

// 在API中使用（通过IOC自动注入）
type AddApi struct {
    DbFactory iCoreDb.IDbFactory `inject:""`
    AddRequest
}

func (a AddApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
    newAccount := AccountModel{AccountName: a.Name}
    err := a.DbFactory.Build(ctx).Add(&newAccount)
    if err != nil {
        return nil, errorSvc.NewError(response.Fail, err)
    }
    return newAccount, nil
}
```

### Model 定义要求

Model 必须实现 `IDbModel` 接口：

```go
type AccountModel struct {
    ID          int    `json:"id" gorm:"primaryKey"`
    AccountName string `json:"account_name"`
}

func (b AccountModel) TableName() string {
    return "account"
}

func (b AccountModel) GetID() int {
    return b.ID
}
```

### 配置说明

```yaml
db:
  mysql:
    username: root
    password: password
    ip: localhost
    port: 3306
    db_name: test
    max_idle_conn: 10
    max_open_conn: 100
    conn_max_lifetime: 3600
    slow_threshold: 100
```

### 单测参考

查看 `service/db/mysqlSvc_test.go` 了解测试方式。

---

## 3. 缓存组件 (Cache)

### 接口定义

```go
type ICacheFactory interface {
    Local() ILocalCache
    Redis() IRedisCache
}

type ILocalCache interface {
    Set(key string, value interface{}, expiration time.Duration) error
    Get(key string) (interface{}, error)
    Delete(key string) error
    Clear() error
}

type IRedisCache interface {
    Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
    Get(ctx context.Context, key string) (string, error)
    Delete(ctx context.Context, key string) error
    // ... 更多方法
}
```

### 使用方法

```go
// 创建缓存工厂
cacheFactory := cache.NewCacheFactory(&config.Cache)

// 注册到IOC容器
ioc.Set(new(iCoreCache.ICacheFactory), cacheFactory)

// 使用本地缓存
localCache := cacheFactory.Local()
err := localCache.Set("key", "value", time.Minute)

// 使用Redis缓存
redisCache := cacheFactory.Redis()
err := redisCache.Set(ctx, "key", "value", time.Minute)
```

### 配置说明

```yaml
cache:
  local:
    enable: true
    default_expire: 3600
  redis:
    enable: true
    addr: localhost:6379
    password: ""
    db: 0
```

### 单测参考

查看 `service/cache/localCacheSvc_test.go` 和 `service/cache/redisCacheSvc_test.go` 了解测试方式。

---

## 4. API响应组件 (Response)

### 接口定义

```go
type IResponse interface {
    Regular(ctx context.Context, data interface{}, err coreError.IError)
}
```

### 使用方法

```go
// 创建响应服务
responseSvc := api.NewResponseSvc()

// 注册到IOC容器
ioc.Set(new(iCoreApi.IResponse), responseSvc)

// 在中间件中使用
type TokenMiddleware struct {
    ResponseSvc iCoreApi.IResponse `inject:""`
}

func (t TokenMiddleware) Middleware() interface{} {
    return func(c *gin.Context) {
        // 校验失败
        t.ResponseSvc.Regular(c, nil, errorSvc.NewError(response.TokenInvalid, err))
        c.Abort()
    }
}
```

### 响应格式

框架统一返回格式：

```json
{
    "code": 200,
    "message": "请求成功",
    "data": {}
}
```

---

## 5. 错误处理组件 (Error)

### 接口定义

```go
type IError interface {
    GetCode() response.Value
    GetMessage() string
    GetData() interface{}
}
```

### 使用方法

```go
// 创建错误
err := errorSvc.NewError(response.Fail, errors.New("操作失败"))
err := errorSvc.Newf(response.ValidateError, "参数校验失败: %s", errMsg)

// 在API中返回错误
func (a AddApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
    if err != nil {
        return nil, errorSvc.NewError(response.Fail, err)
    }
    return result, nil
}
```

### 错误码枚举

| 错误码 | 含义 |
|-------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | Token无效 |
| 500 | 服务器内部错误 |
