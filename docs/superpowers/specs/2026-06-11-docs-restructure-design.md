# 文档体系整治设计

## 目标

将 `docs/` 从松散的文件集合重构为分层的文档体系：统一格式模板、补齐缺失专题、建立交叉引用网络，同时精简 README 中冗余的代码示例。

## 文档分层（4 层）

| 层 | 定位 | 文件 |
|---|------|------|
| 入门 | 新用户从零上手 | `quick_start.md`, `config.md` |
| 核心机制 | 框架关键概念 | `routes_middleware.md`, `ioc_injection.md`, `component_replacement.md`, `error_response.md`(新) |
| 组件使用 | 各模块操作手册 | `database.md`, `cache.md`, `logging.md`, `trace.md`, `sse.md`, `paginator.md`, `validator.md`(新) |
| 参考 | 内部用途 | `superpowers/`（不动） |

## 具体动作

### 新增 2 个文档

**`docs/error_response.md` — 统一错误与响应**

| 章节 | 内容 |
|------|------|
| 概述 | 框架统一响应格式的设计意图 |
| 响应格式 | `{code, message, data}` JSON 结构说明 |
| 错误接口 | `IError` 定义 + 四个构造函数：`New`, `Newf`, `NewError`, `NewErrorf` |
| 响应码一览 | 全部 `response.Value` 枚举常量 + `CodeMap` 中文映射表 |
| 在 API 中使用 | 成功 `return data, nil` / 失败 `return nil, errorSvc.NewError(...)` |
| 相关文档 | 链接到 `routes_middleware.md`, `quick_start.md` |

**`docs/validator.md` — 参数校验**

| 章节 | 内容 |
|------|------|
| 概述 | 框架自动校验流程 |
| 校验流程 | 请求到达 → ShouldBind → 校验失败 → GetErrorMsg → 响应 |
| Binding Tag 参考 | 常用 tag 表格：`required`, `min`, `max`, `email`, `len`, `url`, `eqfield` 等 |
| 自定义错误信息 | `IValidator` 接口 + `GetMessages()` 返回 `entityApi.Messages` |
| 错误消息获取 | `util.GetErrorMsg()` 三级兜底：优先自定义 → validator 默认 → 原始 error |
| 相关文档 | 链接到 `routes_middleware.md` |

### 拆分 `docs/routes_middleware.md`

当前 583 行，包含路由、中间件、参数校验、响应包装、SSE 路由五个主题。拆分如下：

| 内容 | 去向 | 理由 |
|------|------|------|
| SSE 路由注册章节（L412-496） | 移到 `docs/sse.md` 的"路由注册"小节 | 与 SSE 组件文档合并，一处看完 |
| 响应码/错误处理（"返回值自动包装" L283-350） | 精简为 3-5 句概述 + 链到 `error_response.md` | 避免两处维护 |
| 参数校验细节（"校验流程""校验器错误处理"源码 L222-278） | 精简为概述 + 链到 `validator.md` | 同上 |

拆分后 `routes_middleware.md` 预计 ~350 行，聚焦路由注册和中间件。

### 精简 `docs/ioc_injection.md`

删：
- `inject()` 核心注入逻辑完整源码（L117-153）
- ASCII 流程图（L352-379，与 mermaid 重复）
- "单测示例"源码节选（保留链接即可）

保留：概念概述、`Set` / `SetWithName` 用法、`inject` tag 规则、字段顺序约定、命名注入。

### 精简 `docs/component_replacement.md`

- Xorm 完整替换实现删到骨架（保留方法签名 + 注释，删具体实现）
- 其余不变——接口→实现→注册三步模式保留

### 精简 README "核心组件概览"

当前 6 个组件各有一段代码示例（~60 行），这些在 `example.go` 和各组件文档中已有。改为简短的一句话描述 + 文档链接：

```markdown
### IOC 容器
通过 `inject:""` tag 自动注入依赖，无需手动传递。→ [IOC 自动注入](docs/ioc_injection.md)

### API 服务
注册路由时自动完成参数绑定、校验、响应包装。→ [路由与中间件](docs/routes_middleware.md)
```

### 全部组件文档添加交叉引用

6 个组件文档 + 4 个核心文档文末统一添加"相关文档"小节，用表格链接相关 md。格式：

```markdown
## 相关文档

| 文档 | 说明 |
|------|------|
| [路由与中间件](routes_middleware.md) | API 注册与中间件 |
| [统一错误与响应](error_response.md) | 错误码与响应格式 |
```

### 统一格式模板

**核心机制文档：** 概述 → 接口定义 → 核心用法 → 注意事项 → 相关文档

**组件文档：** 概述 → 接口定义 → 初始化 → 操作 → 注意事项 → 相关文档

## 不变

- 框架代码零改动
- `example.go` 不动
- `docs/superpowers/` 不动
- `quick_start.md` 和 `config.md` 内容不动（只加交叉引用）
