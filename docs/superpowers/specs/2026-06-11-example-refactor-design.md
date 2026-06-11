# example.go 重构设计

## 目标

将 `example.go` 重新整理为按模块分节的单文件 Demo，覆盖框架全部功能，配合 README / 快速开始指南中的导航表格，让使用者能顺序阅读或按需跳转学习。

## 分节结构（11 节）

每节开头统一格式：

```go
// ===== N. Section 名称 =====
// 演示：...
// 详见：docs/xxx.md
```

| # | 章节 | 演示内容 | 相关文档 |
|---|------|----------|----------|
| 1 | 启动入口 | `main()` 骨架：初始化→注册 IOC→路由→启动，使用 must 辅助函数简化错误处理，清理逻辑提取 | docs/quick_start.md |
| 2 | 配置加载 | viper 读取 config.yaml，直接传路径，去掉调试打印 | docs/config.md |
| 3 | Model 定义 | IDbModel 接口 (GetID)、TableName、多 Model（AccountModel + OrderModel） | docs/database.md |
| 4 | 请求参数与校验 | binding tag (required/min/max/email)、自定义错误信息 GetMessages() | docs/routes_middleware.md |
| 5 | 中间件 | Token 校验（Authorization 头）、中间件中使用 inject 注入依赖 | docs/routes_middleware.md |
| 6 | 数据库 CRUD | Add / Find / Take / Select / Save / Remove / Order / Count + 事务 Begin/Commit/Rollback | docs/database.md |
| 7 | 多数据库切换 | SetDefault() 切换默认库、Build(ctx, dbEnum.Postgres) 运行时指定 | docs/database.md |
| 8 | 缓存操作 | Local: Get/Set/Del/Exists；Redis: Get/HSet/HGet/HDel；缓存读穿透回写模式 | docs/cache.md |
| 9 | 链路追踪 | FromContext / SetAttribute / RecordError / Start 子 Span / defer span.End() | docs/trace.md |
| 10 | 分页 | Offset 分页 (NewPage)、游标分页 (NewCursor)、nil list 边界 | docs/paginator.md |
| 11 | SSE 推送 | ISse 接口、心跳 goroutine、http.Flusher 断言、Accept 头校验说明 | docs/sse.md |

## 代码改进清单

### 删

- `ReplayMiddleware` 空壳
- `fmt.Println` SSE 调试残留
- `allErrors` 收集模式——demo 不需要这种错误聚合

### 改

- 错误处理：用 `must()` 辅助函数，初始化失败直接 panic，简化 demo 阅读路径
- 配置加载：`os.Getwd()` → 直接传参
- Token 中间件：从 `strconv.ParseUint("id")` 改为标准的 `Authorization` 头校验
- 注入字段顺序：注入服务在前，请求参数在后（遵循现有约定）

### 增

- `OrderModel` 展示多 Model
- CRUD 各操作独立 API（FindApi / TakeApi / UpdateApi / DeleteApi）
- 多数据库切换 API（DbSwitchApi）
- 缓存读穿透 API（CacheReadThroughApi）、Redis Hash 操作 API
- 链路追踪 API（TraceDemoApi）
- SSE 心跳 goroutine
- SSE `http.Flusher` 断言保护

## README / quick_start.md 改动

两处均添加"示例代码导航"表格：

```markdown
### 示例代码导航

`example.go` 按模块分节，以下是各节行号及对应文档：

| 章节 | 行号范围 | 说明 | 详文 |
|------|----------|------|------|
| 1. 启动入口 | L30-L80 | ... | [快速开始](docs/quick_start.md) |
| ... | ... | ... | ... |
```

## 不变

- 框架代码零改动，只改 example.go + README.md + docs/quick_start.md
- Go module 结构不变
- 接口签名不变
