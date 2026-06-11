# 文档体系整治 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `docs/` 从松散文件集合重构为 4 层文档体系：统一格式模板、新增 2 个专题文档、拆分精简 3 个文档、全部 10 个文档添加交叉引用、精简 README 组件概览。

**Architecture:** 纯文档改动，框架代码零改动。按 4 层组织（入门 / 核心机制 / 组件使用 / 参考），每层使用统一的章节模板。所有改动独立可提交。

**Tech Stack:** Markdown

---

## File Structure

| 文件 | 操作 | 职责 |
|------|------|------|
| `docs/error_response.md` | 新建 | 统一错误码与响应格式 |
| `docs/validator.md` | 新建 | 参数校验机制 |
| `docs/routes_middleware.md` | 修改 | 拆分：移除 SSE 路由 / 响应码 / 校验细节 |
| `docs/sse.md` | 修改 | 合并 SSE 路由注册章节 |
| `docs/ioc_injection.md` | 修改 | 精简实现细节 |
| `docs/component_replacement.md` | 修改 | 精简 Xorm 示例 |
| `docs/database.md` | 修改 | 添加交叉引用 |
| `docs/cache.md` | 修改 | 添加交叉引用 |
| `docs/logging.md` | 修改 | 添加交叉引用 |
| `docs/trace.md` | 修改 | 添加交叉引用 |
| `docs/quick_start.md` | 修改 | 添加交叉引用 |
| `docs/config.md` | 修改 | 添加交叉引用 |
| `README.md` | 修改 | 精简核心组件概览 |

---

### Task 1: 新建 docs/error_response.md

**Files:**
- Create: `docs/error_response.md`

- [ ] **Step 1: 创建文件**

写入以下内容到 `D:\Code\hecc-blot\docs\error_response.md`：

```markdown
# 统一错误与响应

框架自动将 API 返回值包装为 `{code, message, data}` 统一格式，并通过 `IError` 接口传递业务错误。

## 响应格式

所有 API 返回统一的 JSON 结构：

```json
{
    "code": 10000,
    "message": "请求成功",
    "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | int | 响应码，详见下方响应码表 |
| `message` | string | 中文说明 |
| `data` | any | 业务数据，失败时为错误详情 |

## 错误接口

```go
// contract/error/error.go
type IError interface {
    error
    GetCode() response.Value
    GetData() interface{}
}
```

提供四个构造函数，位于 `service/error/error_svc.go`：

```go
// 传入响应码 + 原始 error
err := errorSvc.NewError(response.Fail, errors.New("数据库错误"))

// 传入响应码 + 任意 data（可以是 string / struct）
err := errorSvc.New(response.Fail, "用户名已存在")

// 格式化字符串版本
err := errorSvc.NewErrorf(response.Fail, "查询用户 %d 失败", userID)
err := errorSvc.Newf(response.ValidateError, "字段 %s 不能为空", "name")
```

## 响应码一览

定义在 `enum/response/index.go`：

| 常量 | 值 | 说明 |
|------|------|------|
| `Success` | 10000 | 成功 |
| `Processing` | 10001 | 处理中 |
| `Fail` | 40000 | 失败 |
| `Busy` | 40001 | 业务繁忙 |
| `ValidateError` | 40002 | 参数验证失败 |
| `TokenInvalid` | 40003 | 无效 token |
| `AccessDenied` | 40004 | 禁止访问 |
| `NoDataPermission` | 40005 | 无数据处理权限 |
| `Illegal` | 50000 | 非法请求 |
| `Panic` | 50001 | 服务器内部错误 |

中文映射 `CodeMap`：

```go
var CodeMap = map[Value]string{
    Success:          "请求成功",
    Fail:             "请求失败",
    ValidateError:    "参数校验失败",
    TokenInvalid:     "token失效",
    AccessDenied:     "无权访问",
    NoDataPermission: "无权处理",
    Illegal:          "非法请求",
    Panic:            "程序异常",
}
```

## 在 API 中使用

```go
func (a AddApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
    data, err := doSomething()
    if err != nil {
        // 失败：返回 nil + IError，框架包装为失败响应
        return nil, errorSvc.NewError(response.Fail, err)
    }
    // 成功：返回 data + nil，框架包装为成功响应
    return data, nil
}
```

- **成功** `return data, nil` → `{"code": 10000, "message": "请求成功", "data": {...}}`
- **失败** `return nil, errorSvc.NewError(code, err)` → `{"code": 40000, "message": "请求失败", "data": "..."}`

## 相关文档

| 文档 | 说明 |
|------|------|
| [路由与中间件](routes_middleware.md) | 响应自动包装流程 |
| [参数校验](validator.md) | 校验错误的响应格式 |
| [快速开始](quick_start.md) | 完整项目搭建教程 |
```

- [ ] **Step 2: Commit**

```bash
cd "D:\Code\hecc-blot" && git add -f docs/error_response.md && git commit -m "docs: add unified error and response reference"
```

---

### Task 2: 新建 docs/validator.md

**Files:**
- Create: `docs/validator.md`

- [ ] **Step 1: 创建文件**

写入以下内容到 `D:\Code\hecc-blot\docs\validator.md`：

```markdown
# 参数校验

框架基于 `go-playground/validator`，在路由注册时自动绑定请求参数并校验。校验失败时返回统一格式的错误响应。

## 校验流程

```
请求到达 → ShouldBind 绑定参数 → validator 校验
  ├── 通过 → 调用 API.Call()
  └── 失败 → GetErrorMsg() 获取错误消息 → 返回 {code: 40002, message: "..."}
```

## Binding Tag 参考

在请求结构体的字段上使用 `binding` tag：

```go
type AddRequest struct {
    Name     string `json:"name" binding:"required"`
    Age      int    `json:"age" binding:"required,min=1,max=150"`
    Email    string `json:"email" binding:"email"`
    Password string `json:"password" binding:"required,min=6"`
}
```

常用 tag：

| Tag | 说明 | 示例 |
|-----|------|------|
| `required` | 必填 | `binding:"required"` |
| `min` / `max` | 数值或字符串长度 | `binding:"min=1,max=150"` |
| `email` | 邮箱格式 | `binding:"email"` |
| `url` | URL 格式 | `binding:"url"` |
| `len` | 精确长度 | `binding:"len=11"` |
| `eqfield` | 等于另一个字段 | `binding:"eqfield=Password"` |
| `gt` / `gte` / `lt` / `lte` | 大于/大于等于/小于/小于等于 | `binding:"gt=0"` |
| `oneof` | 枚举值 | `binding:"oneof=male female"` |

## 自定义错误信息

实现 `IValidator` 接口，返回字段+规则对应的中文提示：

```go
// contract/api/validator.go
type IValidator interface {
    GetMessages() entityApi.Messages
}
```

```go
// entity/api/validator.go
type Messages map[string]string
```

使用示例：

```go
type AddRequest struct {
    Name string `json:"name" binding:"required"`
}

func (a AddRequest) GetMessages() entityApi.Messages {
    return entityApi.Messages{
        "Name.required": "用户名不能为空",
    }
}
```

Key 格式为 `字段名.规则名`，如 `AccountName.required`、`Password.min`。

## 错误消息获取

`util.GetErrorMsg()` 按三级优先级获取错误消息：

```go
func GetErrorMsg(request interface{}, err error) string
```

1. **自定义消息** — 结构体实现了 `IValidator` 且 `GetMessages()` 中有对应 key → 返回自定义中文
2. **validator 默认** — 返回 validator 内置英文消息（如 `"AccountName is required"`）
3. **原始 error** — 非 validator 错误（如空 body、JSON 格式错误）→ 返回 `err.Error()`

## 相关文档

| 文档 | 说明 |
|------|------|
| [路由与中间件](routes_middleware.md) | 校验触发入口与路由注册 |
| [统一错误与响应](error_response.md) | 校验失败时的响应格式 |
| [快速开始](quick_start.md) | 完整示例 |
```

- [ ] **Step 2: Commit**

```bash
cd "D:\Code\hecc-blot" && git add -f docs/validator.md && git commit -m "docs: add parameter validation reference"
```

---

### Task 3: 拆分 docs/routes_middleware.md（上）— 移除 SSE + 校验 + 响应细节

**Files:**
- Modify: `docs/routes_middleware.md`

**Context:** 当前文件 583 行，需要移除三块内容并替换为精简概述 + 链接。

- [ ] **Step 1: 读取并定位 SSE 路由章节**

SSE 路由注册章节从 "## SSE 路由注册" (约 L412) 到下一个 "##" 或 "***" 分隔符（约 L496）。将整段替换为一个简短引用：

```markdown
***

## SSE 路由注册

SSE（Server-Sent Events）与 API 共享 Gin Engine 和 IMiddleware 接口。详细用法见 [SSE 服务文档](sse.md)。
```

- [ ] **Step 2: 替换"返回值自动包装"章节**

从 "## 返回值自动包装" (约 L283) 到下一个 "***"（约 L350）。替换为：

```markdown
***

## 返回值自动包装

框架自动将 API 返回值包装为 `{code, message, data}` 统一格式。成功时返回 `code: 10000`，失败时根据 `IError.GetCode()` 映射对应响应码。

详见 [统一错误与响应](error_response.md)。
```

- [ ] **Step 3: 替换"校验器错误处理"章节**

从 "### 4. 校验器错误处理" (约 L256) 到下一个 "***"（约 L282）。替换为：

```markdown
### 4. 校验器错误处理

框架通过 `util.GetErrorMsg()` 获取校验错误消息，支持三级兜底：自定义消息 → validator 默认 → 原始 error。

详见 [参数校验](validator.md)。
```

- [ ] **Step 4: Commit**

```bash
cd "D:\Code\hecc-blot" && git add docs/routes_middleware.md && git commit -m "docs(routes): split SSE, validation, and error-response sections to dedicated docs"
```

---

### Task 4: 拆分 docs/sse.md — 合并 SSE 路由注册

**Files:**
- Modify: `docs/sse.md`

**Context:** 将 routes_middleware.md 移除的 SSE 路由注册内容合并到 sse.md。

- [ ] **Step 1: 在"注册路由"章节后追加"SSE 路由注册"小节**

读取 `docs/sse.md`，找到 "## 注册路由" 章节末尾（~L79）。在其后追加一个小节：

```markdown
### 路由注册原理

SSE 处理器接口定义在 `contract/sse/sse_handler.go`：

```go
type ISseHandle interface {
    Get(apiPath string, sse ISse)
    Middleware(middlewares ...iCoreApi.IMiddleware) ISseHandle
}
```

框架自动设置 SSE 响应头（`Content-Type: text/event-stream`、`Cache-Control: no-cache`、`Connection: keep-alive`），设置后立即 Flush 确保客户端及时识别流类型。

**请求处理流程：**

```
请求 → [中间件链] → [设置SSE响应头] → [Serve()] → [持续推送事件]
                                                 ↓
                                         客户端断开 / Serve 返回 error
                                                 ↓
                                         发送 error SSE 事件 → 结束
```

**注意事项：**

- **共享端口**: SSE 与 API 共用同一 Engine，只需调用 `apiHandle.Listen()` 一次
- **长连接**: Serve 方法需监听 `ctx.Request.Context().Done()` 感知客户端断开
- **流刷新**: 每次 `Write` 后需立即 `writer.Flush()`
- **中间件复用**: SSE 与 API 共用 `IMiddleware`，注册方式完全一致
- **错误处理**: Serve 返回 error 时，框架通过 `c.SSEvent("error", ...)` 发送，不修改 HTTP 状态码
```

- [ ] **Step 2: Commit**

```bash
cd "D:\Code\hecc-blot" && git add docs/sse.md && git commit -m "docs(sse): merge SSE routing section from routes_middleware"
```

---

### Task 5: 精简 docs/ioc_injection.md

**Files:**
- Modify: `docs/ioc_injection.md`

- [ ] **Step 1: 删除 inject() 源码节选（L117-153）**

定位到 "### 2. 核心注入逻辑" 章节（~L116）的代码块，将整个代码块 + 周围段落替换为简短说明：

```markdown
### 2. 核心注入逻辑

注入时框架遍历结构体字段，查找 `inject` tag，根据字段类型和名称从容器中获取对应实例并赋值。遇到没有 `inject` tag 的字段时停止注入（因此注入字段必须排在请求参数前面）。
```

- [ ] **Step 2: 删除 ASCII 流程图（L352-379）**

定位到 "## IOC 工作流程图" 章节的 ASCII 文本图。将整个 ASCII 图删除，保留前面的 mermaid 图即可。将章节名改为：

```markdown
## IOC 工作流程图
```

然后删除 ``` 代码块中的 ASCII 流程图（约 L352-378 的整个 ``` 块）。

- [ ] **Step 3: 精简"单测示例"章节（L322-346）**

替换为：

```markdown
## 单测示例

IOC 容器的单元测试见 `service/ioc/ioc_svc_test.go`，演示了 `Set`、`SetWithName`、`Inject` 的标准用法和验证方式。
```

- [ ] **Step 4: Commit**

```bash
cd "D:\Code\hecc-blot" && git add docs/ioc_injection.md && git commit -m "docs(ioc): simplify implementation details, keep concepts and usage"
```

---

### Task 6: 精简 docs/component_replacement.md

**Files:**
- Modify: `docs/component_replacement.md`

- [ ] **Step 1: 精简 Xorm 替换示例（L137-192）**

定位到 "#### 2.2 使用其他 ORM 框架（如 Xorm）" 章节。保留方法骨架，删具体实现：

将当前的完整实现替换为：

```go
// 实现 IDb 接口方法
func (x *XormDbSvc) Add(entry db.IDbModel) error {
    // 使用 x.engine 执行插入操作
    return err
}

func (x *XormDbSvc) Remove(entry db.IDbModel) error {
    // 使用 x.engine 执行删除操作
    return err
}

func (x *XormDbSvc) Query(entry db.IDbModel) db.IDb {
    // 使用 x.engine 构建查询链
    return x
}

func (x *XormDbSvc) Take(dst interface{}) error {
    // 使用 x.engine 执行单条查询
    return err
}

func (x *XormDbSvc) Find(dst interface{}) error {
    // 使用 x.engine 执行多条查询
    return err
}
```

- [ ] **Step 2: Commit**

```bash
cd "D:\Code\hecc-blot" && git add docs/component_replacement.md && git commit -m "docs(component): slim down Xorm replacement example to skeleton"
```

---

### Task 7: 全部文档添加交叉引用

**Files:**
- Modify: `docs/database.md`, `docs/cache.md`, `docs/logging.md`, `docs/trace.md`, `docs/paginator.md`, `docs/sse.md`, `docs/routes_middleware.md`, `docs/ioc_injection.md`, `docs/component_replacement.md`

- [ ] **Step 1: 在 database.md 末尾追加"相关文档"**

```markdown
## 相关文档

| 文档 | 说明 |
|------|------|
| [配置说明](config.md) | 数据库连接配置项 |
| [IOC 注入](ioc_injection.md) | 注入 IDbFactory |
| [缓存组件](cache.md) | 缓存与数据库配合的读穿透模式 |
```

- [ ] **Step 2: 在 cache.md 末尾追加"相关文档"**

```markdown
## 相关文档

| 文档 | 说明 |
|------|------|
| [配置说明](config.md) | 缓存配置项 |
| [链路追踪](trace.md) | 缓存操作的 Trace Span |
| [数据库组件](database.md) | 缓存读穿透模式 |
```

- [ ] **Step 3: 在 logging.md 末尾追加"相关文档"**

```markdown
## 相关文档

| 文档 | 说明 |
|------|------|
| [配置说明](config.md) | 日志配置项 |
| [链路追踪](trace.md) | TraceId 与日志关联 |
| [组件替换](component_replacement.md) | 替换为 logrus 等第三方库 |
```

- [ ] **Step 4: 在 trace.md 末尾追加"相关文档"**

```markdown
## 相关文档

| 文档 | 说明 |
|------|------|
| [配置说明](config.md) | Trace 配置项 |
| [日志组件](logging.md) | TraceId 自动关联日志 |
| [路由与中间件](routes_middleware.md) | HttpTraceMiddleware 自动追踪 |
```

- [ ] **Step 5: 在 paginator.md 末尾追加"相关文档"**

```markdown
## 相关文档

| 文档 | 说明 |
|------|------|
| [数据库组件](database.md) | Offset/Limit 和 WHERE 游标查询 |
| [路由与中间件](routes_middleware.md) | 注册分页 API |
```

- [ ] **Step 6: 在 sse.md 末尾追加"相关文档"**

```markdown
## 相关文档

| 文档 | 说明 |
|------|------|
| [路由与中间件](routes_middleware.md) | API 路由与中间件复用 |
| [日志组件](logging.md) | SSE 连接日志 |
```

- [ ] **Step 7: 在 routes_middleware.md 末尾追加"相关文档"**

读取当前 `docs/routes_middleware.md`，在最后的"总结"章节后追加：

```markdown

## 相关文档

| 文档 | 说明 |
|------|------|
| [统一错误与响应](error_response.md) | 响应格式与错误码 |
| [参数校验](validator.md) | Binding tag 与自定义错误 |
| [IOC 注入](ioc_injection.md) | 中间件和 API 的依赖注入 |
| [SSE 服务](sse.md) | SSE 路由与实时推送 |
```

- [ ] **Step 8: 在 ioc_injection.md 末尾追加"相关文档"**

```markdown
## 相关文档

| 文档 | 说明 |
|------|------|
| [路由与中间件](routes_middleware.md) | 注入在 API 和中间件中的使用 |
| [组件替换](component_replacement.md) | 通过 IOC 替换默认实现 |
```

- [ ] **Step 9: 在 component_replacement.md 末尾追加"相关文档"**

```markdown
## 相关文档

| 文档 | 说明 |
|------|------|
| [IOC 注入](ioc_injection.md) | Set/SetWithName 注册原理 |
| [数据库组件](database.md) | IDbFactory/IDb 接口定义 |
| [缓存组件](cache.md) | ICacheFactory 接口定义 |
| [日志组件](logging.md) | ILog 接口定义 |
```

- [ ] **Step 10: 在 quick_start.md 末尾追加"相关文档"**

在 "## 下一步" 之前追加：

```markdown
## 相关文档

| 文档 | 说明 |
|------|------|
| [配置说明](config.md) | 全部配置项参考 |
| [路由与中间件](routes_middleware.md) | API 注册、校验、响应 |
| [IOC 注入](ioc_injection.md) | 依赖注入原理 |
| [统一错误与响应](error_response.md) | 错误码与响应格式 |
```

- [ ] **Step 11: 在 config.md 末尾追加"相关文档"**

```markdown
## 相关文档

| 文档 | 说明 |
|------|------|
| [快速开始](quick_start.md) | 从零搭建项目 |
| [数据库组件](database.md) | 使用 db 配置 |
| [缓存组件](cache.md) | 使用 cache 配置 |
| [日志组件](logging.md) | 使用 log 配置 |
| [链路追踪](trace.md) | 使用 trace 配置 |
```

- [ ] **Step 12: Commit 全部交叉引用改动**

```bash
cd "D:\Code\hecc-blot" && git add docs/ && git commit -m "docs: add cross-reference sections to all documentation files"
```

---

### Task 8: 精简 README.md "核心组件概览"

**Files:**
- Modify: `README.md`

- [ ] **Step 1: 精简核心组件概览代码块**

定位到 `## 核心组件概览` 章节（约 L119-217），将 6 个组件的代码示例替换为简短描述 + 链接：

```markdown
## 核心组件概览

### IOC 容器

通过 `inject:""` tag 自动注入依赖，无需手动传递。→ [IOC 自动注入说明](docs/ioc_injection.md)

### API 服务

注册路由时自动完成参数绑定、校验、响应包装。→ [路由与中间件说明](docs/routes_middleware.md)

### 数据库服务

支持 MySQL 和 PostgreSQL，链式查询，事务操作。→ [数据库组件说明](docs/database.md)

### 缓存服务

本地内存缓存 + Redis 双层缓存，支持 Hash 操作和读穿透。→ [缓存组件说明](docs/cache.md)

### 日志服务

支持本地文件日志（Zap + lumberjack 滚动）和阿里云 SLS。→ [日志组件说明](docs/logging.md)

### 链路追踪

基于 OpenTelemetry，自动追踪 HTTP 请求并关联日志。→ [链路追踪说明](docs/trace.md)

### SSE 实时推送

与 API 共享端口，通过 `ISse` 接口实现服务端主动推送。→ [SSE 服务](docs/sse.md)

### 分页组件

提供 Offset/Limit 分页和游标分页两种模式。→ [分页组件](docs/paginator.md)
```

- [ ] **Step 2: Commit**

```bash
cd "D:\Code\hecc-blot" && git add README.md && git commit -m "docs(readme): simplify component overview section, replace code blocks with links"
```

---

### Task 9: 最终验证

**Files:**
- 验证: 全部 13 个文件

- [ ] **Step 1: 检查文件清单**

```bash
cd "D:\Code\hecc-blot" && ls -la docs/*.md README.md
```

Expected: 14 files (12 docs + README.md，含新建的 error_response.md 和 validator.md)

- [ ] **Step 2: 检查每个新建/修改的文件都存在且非空**

```bash
cd "D:\Code\hecc-blot" && for f in docs/error_response.md docs/validator.md docs/routes_middleware.md docs/sse.md docs/ioc_injection.md docs/component_replacement.md docs/database.md docs/cache.md docs/logging.md docs/trace.md docs/paginator.md docs/quick_start.md docs/config.md README.md; do echo "$f: $(wc -l < $f) lines"; done
```

- [ ] **Step 3: 确认 git status 干净**

```bash
cd "D:\Code\hecc-blot" && git status
```

Expected: clean（所有改动已提交）。
