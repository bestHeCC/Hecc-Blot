# 缓存组件

Hecc-Blot 支持双层缓存：本地内存缓存（`ILocalCache`）和 Redis 缓存（`IRedisCache`），通过 `ICacheFactory` 统一管理。

## 接口定义

```go
type ICacheFactory interface {
    Local() ILocalCache
    Redis() IRedisCache
}

type IBaseCache interface {
    Set(ctx context.Context, key string, val interface{}, expire time.Duration) error
    Get(ctx context.Context, key string) (interface{}, error)
    Del(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
}

type ILocalCache interface {
    IBaseCache
}

type IRedisCache interface {
    IBaseCache
    HSet(ctx context.Context, key string, values ...interface{}) error
    HGet(ctx context.Context, key, field string) (string, error)
    HDel(ctx context.Context, key string, fields ...string) error
    Close() error
}
```

## 初始化

```go
// 传入 traceSvc 以开启缓存操作的链路追踪，不需要可传 nil
cacheFactory := cache.NewCacheFactory(&config.Cache, traceSvc)

// 注册到 IOC 容器
container := ioc.New()
container.Set(new(iCoreCache.ICacheFactory), cacheFactory)
```

## 本地缓存

基于 `sync.RWMutex` 的内存缓存，定时清理过期条目。

### Set — 写入缓存

```go
err := cacheFactory.Local().Set(ctx, "user:1", userData, 10*time.Minute)
```

`expire` 为 0 表示永不过期。

### Get — 读取缓存

```go
result, err := cacheFactory.Local().Get(ctx, "user:1")
if result != nil {
    user := result.(UserModel)
}
```

key 不存在或已过期返回 `nil, nil`。

### Del — 删除缓存

```go
err := cacheFactory.Local().Del(ctx, "user:1")
```

### Exists — 判断 key 是否存在

```go
ok, err := cacheFactory.Local().Exists(ctx, "user:1")
```

### 过期清理

配置 `clear_interval` 后，框架启动独立 goroutine 定期清理过期条目。清理分两阶段：读锁收集过期 key → 写锁二次确认后删除，不会阻塞正常读写。

```yaml
cache:
  local:
    clear_interval: 3600  # 每隔 3600 秒清理一次
```

## Redis 缓存

基于 go-redis v9，支持 String 和 Hash 操作。

### 基础操作

```go
// Set
err := cacheFactory.Redis().Set(ctx, "key", "value", time.Hour)

// Get
result, err := cacheFactory.Redis().Get(ctx, "key")

// Del
err := cacheFactory.Redis().Del(ctx, "key")

// Exists
ok, err := cacheFactory.Redis().Exists(ctx, "key")
```

### Hash 操作

```go
// HSet
err := cacheFactory.Redis().HSet(ctx, "user:profile", "name", "john", "age", "30")

// HGet
name, err := cacheFactory.Redis().HGet(ctx, "user:profile", "name")

// HDel
err := cacheFactory.Redis().HDel(ctx, "user:profile", "name", "age")
```

### 关闭连接

框架不自动关闭 Redis 连接，如需关闭：

```go
cacheFactory.Redis().Close()
```

## 在 API 中使用

通过 IOC 注入 `ICacheFactory`：

```go
type GetApi struct {
    CacheFactory iCoreCache.ICacheFactory `inject:""`
    DbFactory    iCoreDb.IDbFactory       `inject:""`
}

func (a GetApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
    // 先查本地缓存
    cached, _ := a.CacheFactory.Local().Get(ctx, "account:1")
    if cached != nil {
        return cached, nil
    }

    // 缓存未命中，查数据库
    db := a.DbFactory.Build(ctx)
    data := AccountModel{}
    if err := db.Where("id = ?", 1).Take(&data); err != nil {
        return nil, errorSvc.NewError(response.Fail, err)
    }

    // 回写缓存
    a.CacheFactory.Local().Set(ctx, "account:1", data, 10*time.Minute)

    return data, nil
}
```

## 链路追踪集成

传入 `traceSvc` 后，每次缓存操作自动创建 Span，可在追踪系统中查看缓存操作的耗时和调用关系：

- **本地缓存**: 记录 SET / GET / DEL / EXISTS 操作
- **Redis 缓存**: 记录 SET / GET / DEL / EXISTS / HSET / HGET / HDEL 操作

Span 属性包含 `cache.type`、`cache.operation`、`cache.key`（Redis 为 `db.system: redis`）。

## 相关文档

| 文档 | 说明 |
|------|------|
| [配置说明](config.md) | 缓存配置项 |
| [链路追踪](trace.md) | 缓存操作的 Trace Span |
| [数据库组件](database.md) | 缓存读穿透模式 |
