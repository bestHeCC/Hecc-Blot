# Hecc-Blot

[![Go Version](https://img.shields.io/badge/Go-1.26.1-blue)](https://github.com/hecc/hecc-blot)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)
[![Gitee Repo](https://img.shields.io/badge/Gitee-hecc--blot-red)](https://gitee.com/bestHeCC/hecc-blot)

Hecc-Blot 是一个基于 Go 语言的轻量级后端框架，采用面向接口的设计理念，提供依赖注入、路由注册、参数校验、统一响应等核心功能。

## 框架特性

- **面向接口**: 所有组件通过接口契约交互，易于替换和扩展
- **依赖注入**: 基于反射实现的 IOC 容器，通过 `inject` tag 自动注入
- **路由管理**: 基于 Gin 框架，支持 GET/POST 路由注册和中间件链
- **参数校验**: 自动参数绑定和校验，支持自定义校验错误信息
- **统一响应**: 自动包装返回值为 `{code, message, data}` 统一格式
- **多数据库**: 同时支持 MySQL 和 PostgreSQL，运行时可动态切换
- **事务支持**: 链式调用风格的数据库事务 API
- **双层缓存**: 本地内存缓存 + Redis 缓存，支持过期清理
- **链路追踪**: 基于 OpenTelemetry，支持 OTLP 导出到 Jaeger
- **SSE 推送**: 支持 Server-Sent Events，与 API 共享端口，适用于实时数据推送
- **可替换**: 所有组件可独立替换，只需实现对应接口并注册到 IOC

## 快速开始

```bash
go mod tidy
```

创建 `config.yaml`，参考 [配置说明](docs/config.md)。

```go
func main() {
    config, _ := initConf("/config.yaml")

    logSvc, _ := log.NewLogger(&config.Log)
    traceSvc, _ := trace.NewTraceSvc(&config.Trace)
    dbFactory, clearUp, _ := db.NewDbFactory(&config.Db, logSvc)
    cacheFactory := cache.NewCacheFactory(&config.Cache, traceSvc)
    responseSvc := api.NewResponseSvc()

    ioc.Set(new(iCoreDb.IDbFactory), dbFactory)
    ioc.Set(new(iCoreLog.ILog), logSvc)
    ioc.Set(new(iCoreCache.ICacheFactory), cacheFactory)
    ioc.Set(new(iCoreApi.IResponse), responseSvc)
    ioc.Set(new(iCoreTrace.ITrace), traceSvc)

    apiHandle := api.NewApiSvc(&config.Server, responseSvc, traceSvc)
    apiHandle.Middleware(&TokenMiddleware{})
    apiHandle.Post("account/add", &AddApi{})

    sseHandle := sse.NewSseSvc(apiHandle.Engine())
    sseHandle.Middleware(&TokenMiddleware{})
    sseHandle.Get("events/time", &TimeSse{})

    apiHandle.Listen()
}
```

完整教程见 [快速开始指南](docs/quick_start.md)。

## 目录结构

```
├── contract/         # 接口契约定义
│   ├── api/          # API、中间件、响应、校验器接口
│   ├── cache/        # 本地缓存、Redis 缓存、缓存工厂接口
│   ├── db/           # 数据库操作、数据库工厂、模型接口
│   ├── error/        # 统一错误接口
│   ├── log/          # 日志接口
│   ├── sse/          # SSE 接口
│   └── trace/        # 链路追踪接口
├── entity/           # 实体与配置结构体
│   ├── api/          # 校验器消息类型
│   └── config/       # 配置文件映射（server / db / cache / log / trace）
├── enum/             # 枚举定义（环境 / 数据库类型 / 响应码 / 采样类型）
├── service/          # 服务实现
│   ├── api/          # HTTP 路由、响应包装、校验器
│   ├── cache/        # 本地缓存、Redis 缓存、缓存工厂
│   ├── db/           # MySQL / PostgreSQL（基于 GORM）、工厂、事务
│   ├── error/        # 错误处理
│   ├── ioc/          # IOC 容器
│   ├── log/          # 本地日志（Zap）、阿里云 SLS 日志
│   ├── sse/          # SSE 推送服务
│   └── trace/        # OpenTelemetry 追踪、HTTP 中间件
├── docs/             # 文档
├── util/             # 工具函数
├── example.go        # 完整使用示例
└── README.md
```

## 文档索引

### 入门

| 文档 | 说明 |
|------|------|
| [快速开始指南](docs/quick_start.md) | 从零搭建项目的完整教程 |
| [配置说明](docs/config.md) | config.yaml 全部配置项参考 |

### 核心机制

| 文档 | 说明 |
|------|------|
| [路由与中间件](docs/routes_middleware.md) | 路由注册、中间件定义、参数自动校验、响应包装 |
| [IOC 自动注入](docs/ioc_injection.md) | 依赖注入原理、注入规则、命名注入 |
| [组件替换](docs/component_replacement.md) | 替换日志/数据库/缓存等组件的完整示例 |

### 组件使用

| 文档 | 说明 |
|------|------|
| [日志组件](docs/logging.md) | 本地文件日志、阿里云 SLS 日志的使用与配置 |
| [数据库组件](docs/database.md) | CRUD 操作、事务、多数据库切换、Model 定义 |
| [缓存组件](docs/cache.md) | 本地缓存、Redis 缓存、过期清理、链路追踪集成 |
| [链路追踪](docs/trace.md) | OpenTelemetry 集成、Span 操作、跨服务传递 |
| [SSE 服务](docs/sse.md) | SSE 推送使用、路由注册、中间件复用、错误处理 |
| [分页组件](docs/paginator.md) | offset/limit 分页与游标分页的使用 |

## 核心组件概览

### IOC 容器

通过 `inject:""` tag 自动注入依赖，无需手动传递。

```go
type AddApi struct {
    DbFactory iCoreDb.IDbFactory `inject:""`
    LogSvc    iCoreLog.ILog      `inject:""`
    AddRequest
}
```

→ [IOC 自动注入说明](docs/ioc_injection.md)

### API 服务

注册路由时自动完成参数绑定、校验、响应包装。

```go
apiHandle.Post("account/add", &AddApi{})
```

→ [路由与中间件说明](docs/routes_middleware.md)

### 数据库服务

支持 MySQL 和 PostgreSQL，链式查询，事务操作。

```go
db := a.DbFactory.Build(ctx)
tx := db.Begin()
tx.Add(&account)
tx.Commit()
```

→ [数据库组件说明](docs/database.md)

### 缓存服务

本地内存缓存 + Redis 双层缓存。

```go
a.CacheFactory.Local().Set(ctx, "key", data, time.Minute)
a.CacheFactory.Redis().Get(ctx, "key")
```

→ [缓存组件说明](docs/cache.md)

### 日志服务

支持本地文件日志（Zap + lumberjack 滚动）和阿里云 SLS。

```go
a.LogSvc.Info(ctx, "user login", "user_id", 123)
```

→ [日志组件说明](docs/logging.md)

### 链路追踪

基于 OpenTelemetry，自动追踪 HTTP 请求并关联日志。

```go
ctx, span := a.TraceSvc.Start(ctx, "operation-name", "key", "value")
defer span.End()
```

→ [链路追踪说明](docs/trace.md)

### SSE 实时推送

与 API 共享端口，通过 `ISse` 接口实现服务端主动推送。

```go
type TimeSse struct {
    LogSvc iCoreLog.ILog `inject:""`
}

func (t TimeSse) Serve(ctx *gin.Context) error {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Request.Context().Done():
            return nil
        case <-ticker.C:
            writer := ctx.Writer
            writer.WriteString(fmt.Sprintf("data: %s\n\n", time.Now()))
            writer.Flush()
        }
    }
}
```

→ [SSE 服务](docs/sse.md)

## 设计原则

1. **依赖倒置**: 高层模块依赖抽象接口，而非具体实现
2. **接口隔离**: 每个接口只定义单一职责
3. **开闭原则**: 对扩展开放，对修改关闭

## 路线图

框架的优化规划，详见 [feature.md](feature.md)。

## 感谢

如果你觉得 Hecc-Blot 对你有帮助，欢迎给我们一个 ⭐️

### 反馈与贡献

- **Bug 反馈和功能建议**: 欢迎提交 [Issue](https://gitee.com/bestHeCC/hecc-blot/issues)
- **代码贡献**: 欢迎提交 Pull Request

### 致谢

- [Gin](https://github.com/gin-gonic/gin) — 高性能 Go Web 框架
- [GORM](https://github.com/go-gorm/gorm) — Go ORM 库
- [Zap](https://github.com/uber-go/zap) — 高性能日志库
- [OpenTelemetry](https://opentelemetry.io/) — 分布式追踪标准

## 许可证

MIT License
