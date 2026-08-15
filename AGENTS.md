# Hecc-Blot 开发指南（Agent 手册）

面向 AI 编码代理与贡献者的项目约定文档。写代码前先读这里，尤其是「关键约定」一节。

## 1. 项目定位

Hecc-Blot 是一个基于 Go 的轻量级后端框架，核心理念是**面向接口**：

- 所有能力通过 `hecc-core` 模块的 `contract/` 接口契约暴露，具体实现按模块拆分、可替换。
- 依赖通过反射实现的 IOC 容器自动注入，容器本身也通过 `IContainer` 接口约束、可替换。
- 业务方（如 `example.go`）只依赖接口，不 import 具体实现（IOC 容器创建除外）。

技术栈：Gin（HTTP）、GORM（MySQL/PostgreSQL）、go-redis（缓存）、Zap + lumberjack（日志）、OpenTelemetry（链路追踪）、viper（配置）。

## 2. 模块架构（Monorepo 多 module）

`go.work` 统一管理，根模块 `hecc-blot` 承载 `example.go` 示例，`modules/` 下按依赖分层拆分：

```
modules/
├── ioc/      → github.com/bestHeCC/hecc-ioc    依赖注入容器（仅标准库，零依赖）
├── core/     → github.com/bestHeCC/hecc-core   契约 SDK（contract/entity/enum/util）
│   ├── contract/   接口契约（I* 接口，含 contract/ioc 的 IContainer）
│   ├── entity/     数据实体与配置结构体
│   ├── enum/       枚举（env/db/response/trace）
│   └── util/       工具函数（分页、校验消息、上下文提取）
├── api/      → github.com/bestHeCC/hecc-api    HTTP 内核（api + error + sse + trace中间件）
├── db/       → github.com/bestHeCC/hecc-db     数据库（GORM MySQL/PostgreSQL）
├── cache/    → github.com/bestHeCC/hecc-cache  缓存（本地 + Redis）
├── log/      → github.com/bestHeCC/hecc-log    日志（Zap + SLS）
└── trace/    → github.com/bestHeCC/hecc-trace  链路追踪（OpenTelemetry）
```

依赖方向严格单向：`core → 第三方`，`api/db/cache/log/trace → core`，`api → ioc 接口`。**禁止反向依赖**（core 不得 import 实现模块，实现模块之间不得互相依赖具体实现）。

## 3. 包与导入别名约定

`example.go` 是导入别名的权威范例，遵循：

| 模块路径 | 别名 | 示例 |
|------|------|------|
| `hecc-core/contract/*` | `iCore*` | `iCoreApi "github.com/bestHeCC/hecc-core/contract/api"` |
| `hecc-core/entity/config` | `coreConfig` | `coreConfig "github.com/bestHeCC/hecc-core/entity/config"` |
| `hecc-core/entity/api` | `entityApi` | `entityApi "github.com/bestHeCC/hecc-core/entity/api"` |
| `hecc-core/enum/*` | `*Enum` 或原包名 | `dbEnum`、`envEnum`；`"github.com/bestHeCC/hecc-core/enum/response"` |
| 实现模块（`hecc-api/*` 等） | 原包名，冲突时加后缀 | `errorSvc "github.com/bestHeCC/hecc-api/error"` |

接口命名：`I` 前缀（`IApi`、`IDb`、`ILog`、`ITrace`、`IError`、`ISse`、`IContainer`）。工厂/处理器接口为 `I*Factory`、`I*Handle`。

## 4. 关键约定（不得违反）

### 4.1 IOC 注入字段顺序：服务在前，请求参数在后

`ioc.Inject` 遍历结构体字段时，遇到第一个**没有** `inject` tag 的字段就 `return`（`modules/ioc/ioc_svc.go`）——这是**刻意设计**，不是 bug。

因此定义 API / 中间件结构体时，带 `inject:""` 的依赖字段必须放在最前面，请求参数（含 embed 的 Request 结构体）放在最后：

```go
type AddAccountApi struct {
    DbFactory iCoreDb.IDbFactory `inject:""`  // 注入字段在前
    LogSvc    iCoreLog.ILog      `inject:""`
    AddAccountRequest                          // 请求参数在最后
}
```

不要把这个 `return` 改成 `continue`，也不要打乱字段顺序。

### 4.2 IOC 容器可替换（依赖接口）

框架组件（`hecc-api` 的 `ApiHandle`/`SseHandle`）依赖 `hecc-core/contract/ioc` 的 `IContainer` 接口，**不依赖 `hecc-ioc` 的具体实现**。默认实现是 `hecc-ioc` 的 `Container`（`ioc.New()` 创建），业务方可替换为自己的实现。

初始化时显式创建容器并传入：`api.NewApiSvc(..., container)`（container 为 nil 会 panic）。

### 4.3 DB 链式方法返回新实例（不可变）

`modules/db/base.go` 中 `Where/Order/Limit/Offset/Select/Query` 每次都返回新的 `&BaseDbSvc{...}`，**不是性能问题，是正确性要求**——防止链式调用条件累加污染实例。不要改成原地修改。

需要复用基础查询时，用 `DbFactory.Build(ctx)` 创建新实例，而不是在已带条件的实例上继续链式。

### 4.4 请求级实例隔离

`ApiHandle.registerAPI` 每个请求通过 `reflect.New` 创建独立实例再注入（`modules/api/http_svc.go`），避免并发写同一实例。因此：

- API 结构体不要在注入字段上保存请求态数据。
- 注册时传的是「模板实例」（`apiHandle.Post("path", &AddAccountApi{})`），运行时按类型反射重建。

### 4.5 错误必须走统一错误

`Call` 返回 `(interface{}, IError)`，错误统一用 `hecc-api/error` 构造，配合 `hecc-core/enum/response` 的响应码：

```go
return nil, errorSvc.NewError(response.Fail, err)
```

不要在 `Call` 里直接 `panic` 或裸返回 `fmt.Errorf`（框架层由 `gin.Recovery` 兜底）。

## 5. 核心机制速览

- **IOC**：`container := ioc.New()` 创建容器；`container.Set(new(iCoreDb.IDbFactory), dbFactory)` 注册；`container.Inject(instance)` 注入。框架组件依赖 `IContainer` 接口，容器可替换。
- **路由**：`IApiHandle.Get/Post` 注册，自动完成 `ShouldBind` + 校验 + 响应包装。自定义校验错误信息通过实现 `IValidator` 的 `GetMessages()`（`entity/api` 的 `Messages` map，key 为 `Field.Tag`）。
- **响应**：统一 `{code, message, data}`；`code` 见 `enum/response`（`Success=10000`、`Fail=40000`、`ValidateError=40002`…）。
- **数据库**：`IDbModel`（`GetID() int` + `TableName()`）定义模型；`IDbFactory.Build(ctx, [dbEnum.Postgres])` 取库；事务用 `Begin()/Commit()/Rollback()`（在返回的 tx 实例上调用）。
- **缓存**：`ICacheFactory.Local()/Redis()`，本地缓存 + Redis 双层；`IBaseCache` 提供 `Set/Get/Del/Exists`。
- **日志**：`ILog.Info/Debug/Warn/Error(ctx, msg, fields...)`，`fields` 用 `zap` 字段（`zap.String(...)`）。自动附加 `traceId`。
- **链路追踪**：`ITrace.Start/FromContext` 返回 `Span`，支持 `SetAttribute/RecordError/End`。`HttpTraceMiddleware`（在 `hecc-api` 内）自动创建请求 Span 并注入 `traceId` 到 context（key 见 `enum/trace`）。
- **SSE**：实现 `ISse.Serve(ctx) error`，通过 `ISseHandle.Get` 注册，与 API 共享端口。注意 `http.Flusher` 断言与 `Accept: text/event-stream` 校验（见 `feature.md` 已知问题）。

## 6. 新增一个业务接口的标准步骤

以「新增账户查询接口」为例：

1. 定义模型（若新表）：实现 `IDbModel` + `TableName()`。
2. 定义请求结构体：`binding` tag 校验 + 可选 `GetMessages()`。
3. 定义 API 结构体：注入字段在前，embed 请求在后，实现 `Call(ctx) (interface{}, IError)`。
4. 在路由注册处 `apiHandle.Get("xxx/yyy", &XxxApi{})`。
5. 启动前在 `main` 把依赖 `container.Set` 进容器。

完整可运行范式直接照抄 `example.go`（11 个分节覆盖全部能力）。

## 7. 新增 / 替换组件

替换某个组件（如日志、缓存、数据库，甚至 IOC 容器）只需：

1. 在对应 `modules/` 下实现对应 `contract/` 接口。
2. 在初始化处用新实现注册（IOC 容器用 `ioc.New()` 换成自己的实现，普通组件用 `container.Set`）。

不要改动 `contract/` 已有接口签名；确需扩展时，新增方法或新接口而非破坏旧方法。

## 8. 代码风格

- **注释用中文**，关键设计决策写清「为什么」（参考 `modules/db/base.go`、`modules/ioc/ioc_svc.go` 的注释风格）。
- 错误初始化用 `must/must2`（panic 语义），业务错误走统一 `IError`。
- 泛型用于分页工具（`util.NewPage[T]`、`util.NewCursor[T]`），不要为微优化滥用。
- 测试用 `testify`，测试文件与实现同包（`_test.go` 已有范例：`modules/db/*_test.go`、`modules/cache/*_test.go`、`modules/log/*_test.go`）。

## 9. 测试与验证

Monorepo 多 module，用 `go.work` 管理。在**各模块目录内**运行：

```bash
cd modules/<xxx> && go mod tidy && go build ./... && go test ./...
```

或从根目录对根模块（`example.go`）：

```bash
go build ./... && go test ./...
```

注意：测试涉及 Redis/MySQL/PostgreSQL 的用例（`modules/db`、`modules/cache` 部分）可能依赖真实服务，本地跑不动时优先跑纯单元测试。

## 10. 已知规划

`feature.md` 记录了 SSE 模块的待优化项（并发安全、连接上限、心跳超时、`Last-Event-Id` 断线续传等），是后续迭代的路标。改动 SSE 前先对照该文件，避免重复设计。

更详细的分模块文档见 `docs/`（`quick_start.md`、`ioc_injection.md`、`routes_middleware.md`、`database.md`、`cache.md`、`logging.md`、`trace.md`、`sse.md`、`paginator.md`、`component_replacement.md`）。
