# 快速开始指南

本文档基于 `example.go` 完整示例，带你从零搭建一个 Hecc-Blot 项目。

## 1. 项目初始化

```bash
go mod init your-project
go mod tidy
```

## 2. 创建配置文件

在项目根目录创建 `config.yaml`：

```yaml
server:
  port: "9500"
  env: dev
  name: Hecc-Blot
  enable_trace: true
  read_timeout: 30
  write_timeout: 30
  idle_timeout: 60
  body_size_limit: 10485760  # 10MB

db:
  mysql:
    username: root
    password: "123456"
    ip: 127.0.0.1
    port: 3306
    db_name: test
    max_idle_conn: 10
    max_open_conn: 100
    conn_max_lifetime: 3600
    slow_threshold: 200
  postgres:
    username: postgres
    password: "123456"
    ip: 127.0.0.1
    port: 5432
    db_name: test
    max_idle_conn: 10
    max_open_conn: 100
    conn_max_lifetime: 3600
    slow_threshold: 200

cache:
  local:
    enable: true
    clear_interval: 3600
  redis:
    addr: "127.0.0.1:6379"
    password: ""
    db: 0
    pool_size: 100

log:
  local:
    enable: true
    root_dir: runtime/logs
    max_size: 100
    max_backups: 30
    max_age: 7
    compress: true

trace:
  service_name: Hecc-Blot
  endpoint: 127.0.0.1:4318
  sampler:
    type: always
    ratio: 0.5
```

完整配置项说明见 [配置说明](config.md)。

## 3. 编写配置加载函数

```go
func initConf(confPath string) (*config.Config, error) {
    viper.SetConfigFile(confPath)
    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }
    var conf config.Config
    if err := viper.Unmarshal(&conf); err != nil {
        return nil, err
    }
    return &conf, nil
}
```

## 4. 定义数据模型

Model 需要实现 `IDbModel` 接口（返回主键 ID）：

```go
type AccountModel struct {
    ID          int    `json:"id" gorm:"primaryKey"`
    AccountName string `json:"account_name"`
    Password    string `json:"password"`
    CreatedAt   int    `json:"created_at"`
    UpdatedAt   int    `json:"updated_at"`
    DeletedAt   int    `json:"deleted_at"`
}

func (a AccountModel) TableName() string {
    return "account"
}

func (a AccountModel) GetID() int {
    return a.ID
}
```

## 5. 定义请求参数与校验

```go
type AddAccountRequest struct {
    AccountName string `json:"account_name" binding:"required"`
    Password    string `json:"password" binding:"required,min=6"`
}

// 可选：自定义校验错误信息
func (a AddAccountRequest) GetMessages() entityApi.Messages {
    return entityApi.Messages{
        "AccountName.required": "用户名不能为空",
        "Password.required":    "密码不能为空",
        "Password.min":         "密码长度至少6位",
    }
}
```

## 6. 实现 API

```go
type AddAccountApi struct {
    // 注入字段放在前面
    DbFactory    iCoreDb.IDbFactory    `inject:""`
    LogSvc       iCoreLog.ILog         `inject:""`
    CacheFactory iCoreCache.ICacheFactory `inject:""`

    // 请求参数通过匿名嵌入放在最后
    AddAccountRequest
}

func (a AddAccountApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
    data := AccountModel{
        AccountName: a.AccountName,
        Password:    a.Password,
    }

    // 数据库操作
    db := a.DbFactory.Build(ctx)
    if err := db.Add(&data); err != nil {
        return nil, errorSvc.NewError(response.Fail, err)
    }

    // 写入缓存
    a.CacheFactory.Local().Set(ctx, fmt.Sprintf("account:%d", data.ID), data, 10*time.Minute)

    // 日志记录
    a.LogSvc.Info(ctx, "account created", "id", data.ID)

    return data, nil
}
```

## 7. 实现中间件

```go
type TokenMiddleware struct {
    ResponseSvc iCoreApi.IResponse `inject:""`
}

func (t TokenMiddleware) Middleware() interface{} {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            t.ResponseSvc.Regular(c, nil, errorSvc.NewError(response.TokenInvalid, errors.New("token 为空")))
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## 8. 组装启动入口

```go
func main() {
    // 1. 加载配置
    config, err := initConf("/config.yaml")
    if err != nil {
        panic(err)
    }

    // 2. 初始化日志
    logSvc, err := log.NewLogger(&config.Log)
    if err != nil {
        panic(err)
    }

    // 3. 初始化链路追踪
    traceSvc, traceClearUp, err := trace.NewTraceSvc(&config.Trace)
    if err != nil {
        panic(err)
    }
    defer traceClearUp()

    // 4. 初始化数据库
    dbFactory, dbClearUp, err := db.NewDbFactory(&config.Db, logSvc)
    if err != nil {
        panic(err)
    }
    defer dbClearUp()

    // 5. 初始化缓存
    cacheFactory := cache.NewCacheFactory(&config.Cache, traceSvc)

    // 6. 初始化响应服务
    responseSvc := api.NewResponseSvc()

    // 7. 注册到 IOC 容器（顺序无关，但必须在路由注册之前）
    ioc.Set(new(iCoreDb.IDbFactory), dbFactory)
    ioc.Set(new(iCoreLog.ILog), logSvc)
    ioc.Set(new(iCoreCache.ICacheFactory), cacheFactory)
    ioc.Set(new(iCoreApi.IResponse), responseSvc)
    ioc.Set(new(iCoreTrace.ITrace), traceSvc)

    // 8. 创建 API 处理器并注册路由
    apiHandle := api.NewApiSvc(&config.Server, responseSvc, traceSvc)
    apiHandle.Middleware(&TokenMiddleware{})
	{
	apiHandle.Post("account/add", &AddAccountApi{})
	apiHandle.Get("account/list", &ListAccountApi{})
    }

    // 9. 注册 SSE 路由（可选，需要时使用）
    sseHandle := sse.NewSseSvc(apiHandle.Engine())
    sseHandle.Middleware(&TokenMiddleware{})
    {
        sseHandle.Get("events/time", &TimeSse{})
    }

    // 10. 启动服务（阻塞直到收到 SIGINT/SIGTERM）
    apiHandle.Listen()
}
```

## 9. 运行

```bash
go run main.go
```

服务启动后访问 `http://localhost:9500/account/add`，请求示例：

```json
{
    "account_name": "john",
    "password": "123456"
}
```

成功响应：

```json
{
    "code": 200,
    "message": "请求成功",
    "data": {
        "id": 1,
        "account_name": "john",
        "password": "123456",
        "created_at": 1779088256,
        "updated_at": 1779088256,
        "deleted_at": 0
    }
}
```

## 示例代码导航

完整的 [`example.go`](../example.go) 按模块分为 11 节，覆盖框架全部功能：

| # | 章节 | 演示内容 | 详文 |
|---|------|----------|------|
| 1 | 启动入口 | main() 骨架 | [快速开始](quick_start.md) |
| 2 | 配置加载 | viper 读取 config.yaml | [配置说明](config.md) |
| 3 | Model 定义 | IDbModel、TableName、多 Model | [数据库组件](database.md) |
| 4 | 请求参数与校验 | binding tag、GetMessages() | [路由与中间件](routes_middleware.md) |
| 5 | 中间件 | Authorization 校验、inject | [路由与中间件](routes_middleware.md) |
| 6 | 数据库 CRUD | Add/Take/Find/Save/Remove/Count/事务 | [数据库组件](database.md) |
| 7 | 多数据库切换 | MySQL ↔ PostgreSQL | [数据库组件](database.md) |
| 8 | 缓存操作 | Local/Redis 读写删/Hash/读穿透 | [缓存组件](cache.md) |
| 9 | 链路追踪 | Span/SetAttribute/子 Span | [链路追踪](trace.md) |
| 10 | 分页 | Offset 分页 + 游标分页 | [分页组件](paginator.md) |
| 11 | SSE 推送 | ISse 接口、心跳、Flusher | [SSE 服务](sse.md) |

## 相关文档

| 文档 | 说明 |
|------|------|
| [配置说明](config.md) | 全部配置项参考 |
| [路由与中间件](routes_middleware.md) | API 注册、校验、响应 |
| [IOC 注入](ioc_injection.md) | 依赖注入原理 |
| [统一错误与响应](error_response.md) | 错误码与响应格式 |

## 下一步

- [配置项完整参考](config.md)
- [路由与中间件详解](routes_middleware.md)
- [数据库操作与事务](database.md)
- [缓存使用](cache.md)
- [日志组件](logging.md)
- [链路追踪](trace.md)
- [IOC 注入机制](ioc_injection.md)
- [SSE 服务](sse.md)
- [替换框架组件](component_replacement.md)
- [分页组件](paginator.md)
