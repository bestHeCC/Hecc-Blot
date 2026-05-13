# Hecc-Go-Core

Hecc-Go-Core 是一个基于 Go 语言的轻量级后端框架，采用面向接口的设计理念，提供依赖注入、路由注册、参数校验、统一响应等核心功能。

## 框架特性

- **面向接口**: 所有组件通过接口契约交互，易于替换和扩展
- **依赖注入**: 基于反射实现的 IOC 容器，支持自动注入
- **路由管理**: 基于 Gin 框架的路由注册机制
- **参数校验**: 自动参数绑定和校验，支持自定义错误信息
- **统一响应**: 自动包装返回值为统一格式
- **组件化设计**: 日志、数据库、缓存等组件可独立替换

## 目录结构

```
├── contract/         # 接口契约定义
│   ├── api/          # API 相关接口
│   ├── cache/        # 缓存相关接口
│   ├── db/           # 数据库相关接口
│   ├── error/        # 错误处理接口
│   └── log/          # 日志相关接口
├── entity/           # 实体定义
│   ├── api/          # API 实体
│   └── config/       # 配置实体
├── enum/             # 枚举定义
├── service/          # 服务实现
│   ├── api/          # API 服务
│   ├── cache/        # 缓存服务
│   ├── db/           # 数据库服务
│   ├── error/        # 错误处理服务
│   ├── ioc/          # IOC 容器服务
│   └── log/          # 日志服务
├── docs/             # 文档目录
├── example.go        # 使用示例
└── README.md         # 项目说明
```

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 创建配置文件

```yaml
# config.yaml
server:
  port: "8080"
  env: dev

log:
  local:
    enable: true
    root_dir: ./logs
    max_size: 100
    max_backups: 30
    max_age: 7
    compress: true

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
  postgres:
    username: root
    password: password
    ip: localhost
    port: 5432
    db_name: test
    max_idle_conn: 10
    max_open_conn: 100
    conn_max_lifetime: 3600
    slow_threshold: 100

cache:
  local:
    enable: true
    default_expire: 3600
```

### 3. 编写 API

```go
type AddRequest struct {
    Name string `json:"name" binding:"required"`
}

func (a AddRequest) GetMessages() entityApi.Messages {
    return entityApi.Messages{
        "Name.required": "名字不能为空",
    }
}

type AddApi struct {
    DbFactory iCoreDb.IDbFactory `inject:""`
    LogSvc    iCoreLog.ILog      `inject:""`
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

### 4. 注册路由

```go
func main() {
    config, _ := initConf("/config.yaml")
    
    logSvc, _ := log.NewLogger(&config.Log)
    dbFactory, clearUp, _ := db.NewDbFactory(&config.Db, logSvc)
    cacheFactory := cache.NewCacheFactory(&config.Cache)
    responseSvc := api.NewResponseSvc()
    
    ioc.Set(new(iCoreDb.IDbFactory), dbFactory)
    ioc.Set(new(iCoreLog.ILog), logSvc)
    ioc.Set(new(iCoreCache.ICacheFactory), cacheFactory)
    ioc.Set(new(iCoreApi.IResponse), responseSvc)
    
    apiHandle := api.NewApiSvc(&config.Server, responseSvc)
    apiHandle.Middleware(&TokenMiddleware{})
    {
        apiHandle.Post("account/add", &AddApi{})
    }
    apiHandle.Listen()
}
```

## 文档目录

- [组件使用说明](docs/component_usage.md) - 各组件的详细使用方法
- [组件替换说明](docs/component_replacement.md) - 如何替换框架组件
- [IOC 自动注入说明](docs/ioc_injection.md) - IOC 容器工作原理
- [路由和中间件说明](docs/routes_middleware.md) - 路由注册和参数校验机制
- [链路追踪说明](docs/trace.md) - OpenTelemetry + Jaeger 集成使用

## 核心组件

### 1. IOC 容器
负责依赖注入和组件生命周期管理

### 2. API 服务
处理路由注册、参数校验、响应包装

### 3. 日志服务
支持本地日志和阿里云 SLS，可自动关联 TraceId

### 4. 数据库服务
支持 MySQL 和 PostgreSQL，基于 GORM 实现。框架通过工厂模式支持同时配置多个数据库类型，并可动态切换默认数据库。

```go
// 设置默认数据库为 PostgreSQL
dbFactory.SetDefault(dbEnum.Postgres)

// 或者在运行时指定数据库类型
dbFactory.Build(ctx, dbEnum.Mysql)  // 使用 MySQL
dbFactory.Build(ctx, dbEnum.Postgres)  // 使用 PostgreSQL
```

### 5. 缓存服务
支持本地缓存和 Redis

### 6. 链路追踪服务
基于 OpenTelemetry 的分布式追踪，默认支持 Jaeger

## 设计原则

1. **依赖倒置**: 高层模块依赖抽象接口，而非具体实现
2. **接口隔离**: 每个接口只定义单一职责
3. **开闭原则**: 对扩展开放，对修改关闭
4. **单一职责**: 每个组件只负责一个功能

## 许可证

MIT License
