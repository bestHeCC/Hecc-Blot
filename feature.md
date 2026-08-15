# 框架优化规划

各模块待优化项与生产就绪补充项，按模块分章节，按阶段排列。

---

## SSE 模块

当前已具备基础推送能力（响应头、Flush、中间件复用、error 事件）。以下为完善计划。

### 第一阶段：稳定性 ✅ 已完成

#### 1.1 并发安全 — 实例隔离

**位置:** `service/sse/sse_svc.go:30-42`

**问题:** `sseInstance` 在注册时注入一次，所有并发连接共享。若 `Serve()` 中有任何状态字段，存在数据竞争。

**方案:** 参考 API 修复，每个连接用 `reflect.New` 创建独立实例并注入依赖。

#### 1.2 连接数上限

**位置:** `service/sse/sse_svc.go:45-49`

**问题:** 无并发连接数限制，恶意客户端可耗尽 goroutine / 内存 / fd。

**方案:** `SseHandle` 内部维护 buffered channel 作为信号量，超出上限返回 503。

#### 1.3 心跳与空闲超时

**位置:** `service/sse/sse_svc.go:39`

**问题:** 客户端非正常断开（网线拔掉、进程 kill 无 FIN）时，服务端 goroutine 永久泄漏。TCP keepalive 默认 2h，不可依赖；`WriteTimeout` 每次写入会重置，无法保护空闲连接。

**方案:** 框架在 handler 中启动心跳 goroutine，每 30s 发 SSE comment（`: heartbeat\n\n`）。心跳写入失败时 cancel context，结束 Serve。

#### 1.4 `http.Flusher` 断言

**位置:** `service/sse/sse_svc.go:37`

**问题:** `c.Writer.Flush()` 直接调用，若 Writer 被中间件包装且未实现 `http.Flusher`，会 panic。

**方案:** 写入前断言 `flusher, ok := c.Writer.(http.Flusher)`，不支持则返回 500。

#### 1.5 `Accept` 头校验

**位置:** `service/sse/sse_svc.go:33`

**问题:** 任何请求（浏览器直接访问、curl 默认）都建立 SSE 连接，客户端可能不支持。

**方案:** 通过中间件实现（框架不内置策略性校验），参考 `example.go` 的 `SseAcceptMiddleware`。

#### 1.6 错误事件格式规范

**位置:** `service/sse/sse_svc.go:40`

**问题:** `c.SSEvent("error", err.Error())` 只发纯文本，客户端无法区分错误类型。HTTP 状态码已固定为 200，无法表达 4xx/5xx。

**方案:** 错误帧使用 JSON 格式：`event: error\ndata: {"code":"xxx","message":"..."}\n\n`。

#### 1.7 Last-Event-Id 断线续传

**位置:** `service/sse/sse_svc.go:30-42`

**问题:** 不支持 SSE 标准的 `Last-Event-Id` 机制。客户端断开重连时无法从断点续推，断线期间事件丢失。

**方案:**
1. 框架提取请求头 `Last-Event-Id`，注入 `Context`
2. 提供 `WriteSSE(id, event, data)` 辅助方法，自动写入 `id:` 帧头
3. 业务层自行实现消息回溯（Redis Stream / DB），框架只提供 ID 传递通道

---

### 第二阶段：可运维

#### 2.1 优雅关闭通知

**位置:** `service/api/http_svc.go:63-107` + `service/sse/sse_svc.go`

**问题:** `Listen()` 收到 SIGTERM 后仅等 5s 就强制关停，SSE 连接无任何通知，客户端表现为连接突然断开而非主动重连。

**方案:** `SseHandle` 维护活跃连接表；收到关闭信号后遍历发送 `event: shutdown` 帧，等待 2s 让客户端收到后主动断开，再执行 `srv.Shutdown()`。

#### 2.2 可观测性指标

**位置:** `service/sse/sse_svc.go` — 新增

**问题:** 无法回答：当前活跃连接数？事件发送速率？断开原因分布？排障和容量规划完全黑盒。

**方案:** `SseHandle` 暴露 `Stats()` 接口，维护活跃连接 map + atomic 计数器（连接总数 / 断开总数）。

---

### 第三阶段：可接入

#### 3.1 CORS 支持

**位置:** `service/sse/sse_svc.go:33`

**问题:** 浏览器 `EventSource` 遵循同源策略，跨域需 `Access-Control-Allow-Origin`。且 `EventSource` 不支持自定义请求头，无法携带 token。

**方案:** SSE handler 自动设置 CORS 响应头。跨域鉴权通过 URL query 参数（`?token=xxx`）或预签发一次性 ticket 实现。

#### 3.2 SSE 帧辅助方法

**位置:** `util/` — 新增 `sse_writer.go`

**问题:** 用户手动拼接 `data: ...\n\n`、`event: ...\n\n`、`id: ...\n\n`，容易漏换行符。

**方案:** 提供 `util.WriteSSE(w, id, event, data)` 辅助方法，自动处理格式。

---

### 第四阶段：高性能

#### 4.1 背压控制

**位置:** `util/sse_writer.go` — 新增

**问题:** 生产速度 > 消费速度时，TCP 发送缓冲区填满导致 `Write()` 阻塞或 OOM。

**方案:** 提供 `WriteSSEDrop()` 非阻塞写入，写入失败时丢弃当前帧，不阻塞事件循环。业务层自行决定丢弃或降速策略。

#### 4.2 帧 Buffer 复用

**位置:** `example.go:328` 及用户代码

**问题:** 每条消息用 `fmt.Sprintf` 拼接，高频推送时 GC 压力显著。

**方案:** `WriteSSE` 内部使用 `bytes.Buffer` 池或 `strings.Builder` 复用。

---

### 第五阶段：可观测

#### 5.1 Trace Span 注入

**位置:** `service/sse/sse_svc.go:39`

**问题:** SSE 事件推送无链路追踪，出问题时无法关联到具体连接和消息。

**方案:** 框架在 `Serve()` 调用前将当前 trace span 注入 Context，`WriteSSE` 支持创建子 span 追踪每条推送。

---

### 汇总

| # | 阶段 | 条目 | 影响 |
|---|------|------|------|
| 1.1 | 稳定性 | ✅ 实例隔离 | 并发数据竞争 |
| 1.2 | 稳定性 | ✅ 连接数上限 | 资源耗尽 |
| 1.3 | 稳定性 | ✅ 心跳超时 | goroutine 泄漏 |
| 1.4 | 稳定性 | ✅ Flusher 断言 | panic |
| 1.5 | 稳定性 | ✅ Accept 校验（中间件） | 误连接 |
| 1.6 | 稳定性 | ✅ 错误格式规范 | 客户端误判 |
| 1.7 | 稳定性 | ✅ Last-Event-Id | 断线丢消息 |
| 2.1 | 可运维 | 优雅关闭 | 重启连接异常断开 |
| 2.2 | 可运维 | 可观测指标 | 排障黑盒 |
| 3.1 | 可接入 | CORS | 浏览器不可用 |
| 3.2 | 可接入 | WriteSSE 辅助 | 手动拼帧易出错 |
| 4.1 | 高性能 | 背压控制 | OOM / 连接饥饿 |
| 4.2 | 高性能 | Buffer 复用 | GC 压力 |
| 5.1 | 可观测 | Trace 注入 | SSE 无链路追踪 |
