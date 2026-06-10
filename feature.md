# 框架优化规划

## 高优先级

### 1. SSE 实例共享导致并发数据竞争

**位置:** `service/sse/sse_svc.go:30-42`

**问题:** `sseInstance` 在 `Get()` 中注入一次后，所有并发 SSE 连接共享同一个实例。若 `Serve()` 中有状态修改（如计数器、内部缓冲），多连接并发读写会产生数据竞争。

**方案:** 参考 API 修复方案，用反射为每个连接创建独立实例：

```go
func (f *SseHandle) Get(apiPath string, sseInstance iCoreSse.ISse) {
    ioc.Inject(sseInstance)
    sseType := reflect.TypeOf(sseInstance).Elem()

    f.engine.GET(apiPath, func(c *gin.Context) {
        newInstance := reflect.New(sseType).Interface()
        ioc.Inject(newInstance)
        sse := newInstance.(iCoreSse.ISse)

        c.Writer.Header().Set("Content-Type", "text/event-stream")
        c.Writer.Header().Set("Cache-Control", "no-cache")
        c.Writer.Header().Set("Connection", "keep-alive")
        c.Writer.Flush()

        if err := sse.Serve(c); err != nil {
            c.SSEvent("error", err.Error())
        }
    })
}
```

---

### 2. 缺少连接数限制

**位置:** `service/sse/sse_svc.go:45-49` — `NewSseSvc()`

**问题:** 无任何并发连接数上限，恶意或异常的客户端可以无限打开 SSE 连接，耗尽 goroutine、内存和文件描述符。

**方案:** 在 `SseHandle` 中维护一个带缓冲的 channel 作为信号量：

```go
type SseHandle struct {
    engine    *gin.Engine
    semaphore chan struct{} // 连接数限制
}

func NewSseSvc(engine *gin.Engine, maxConns int) iCoreSse.ISseHandle {
    if maxConns <= 0 {
        maxConns = 1000
    }
    return &SseHandle{
        engine:    engine,
        semaphore: make(chan struct{}, maxConns),
    }
}

// 在 handler 中:
select {
case f.semaphore <- struct{}{}:
    defer func() { <-f.semaphore }()
default:
    c.JSON(http.StatusServiceUnavailable, gin.H{"error": "too many connections"})
    return
}
```

---

### 3. 缺少连接心跳与空闲超时

**位置:** `service/sse/sse_svc.go:39` — `sseInstance.Serve(c)` 调用处

**问题:** SSE 连接是长连接，若客户端非正常断开（如网线拔掉但没有发送 TCP RST），服务端 goroutine 会永久挂起在 `Serve()` 中。TCP keepalive 默认 2 小时，不可依赖。同时，Gin 层面的 `WriteTimeout` 在 SSE 场景下每帧写入都会重置计时器，无法保护空闲连接。

**方案:** 框架层面定期写入 SSE comment 作为心跳。同时还需在 `http.Server` 层为 SSE 路由设置独立的 `ReadTimeout`。更轻量的方案：在 `Serve()` 调用处用 `context.WithTimeout` 包一层，在 `SseHandle` 中注入心跳写入器：

```go
// 简化方案：在 handler 中注入带超时的 context
ctx, cancel := context.WithCancel(c.Request.Context())
defer cancel()

// 启动心跳 goroutine
go func() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            c.Writer.Write([]byte(": heartbeat\n\n"))
            c.Writer.Flush()
        }
    }
}()

if err := sse.Serve(c); err != nil {
    c.SSEvent("error", err.Error())
}
```

---

## 中优先级

### 4. `http.Flusher` 接口断言缺失

**位置:** `service/sse/sse_svc.go:37` — `c.Writer.Flush()`

**问题:** `gin.ResponseWriter` 默认实现了 `http.Flusher`，但若有自定义 Writer 包装器未实现该接口，`Flush()` 会 panic。应在获取 flusher 时做检查：

```go
flusher, ok := c.Writer.(http.Flusher)
if !ok {
    c.String(http.StatusInternalServerError, "streaming not supported")
    return
}
// ... 后续用 flusher.Flush()
```

---

### 5. SSE 错误响应写入了非标准数据

**位置:** `service/sse/sse_svc.go:40` — `c.SSEvent("error", err.Error())`

**问题:** `c.SSEvent("error", ...)` 向客户端推送一个 `event: error` 帧后，handler 直接 return。但此时 SSE 响应头已发送，HTTP 状态码固定为 200，无法返回 4xx/5xx。调用方接收到 error 事件后无法区分是业务错误还是连接终断。更好的做法是按约定格式封装：

```go
if err := sse.Serve(c); err != nil {
    // 用自定义格式发送错误，区分错误类型
    fmt.Fprintf(c.Writer, "event: error\ndata: {\"message\":\"%s\"}\n\n", err.Error())
    c.Writer.Flush()
}
```

---

### 6. 缺少 `Accept` 头校验

**位置:** `service/sse/sse_svc.go:33`

**问题:** 任何请求（包括浏览器直接访问、curl 无参数）都会建立 SSE 连接并设置响应头，但客户端可能不支持 SSE。应检查 `Accept: text/event-stream` 头：

```go
if !strings.Contains(c.GetHeader("Accept"), "text/event-stream") {
    c.String(http.StatusNotAcceptable, "SSE requires Accept: text/event-stream")
    return
}
```

---

## 低优先级

### 7. 手动构造 SSE 格式容易出错

**位置:** `example.go:328` — `fmt.Sprintf("data: 当前服务器时间：%s\n\n", ...)`

**问题:** 用户每次需要手动拼接 SSE 帧格式（`data: ...\n\n`、`event: ...\n\n`、`id: ...\n\n`），容易漏掉换行符或格式错误。框架可以提供辅助方法：

```go
// 添加到 ISseHandle 或新建 SSE util
func WriteSSE(w io.Writer, event, data string) error {
    if event != "" {
        if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
            return err
        }
    }
    if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
        return err
    }
    return nil
}
```

---

### 8. `fmt.Sprintf` 每帧分配

**位置:** `example.go:328`

**问题:** 每条 SSE 消息都用 `fmt.Sprintf` 拼接字符串，产生堆分配。高频推送场景下 GC 压力明显。

**方案:** 复用 `strings.Builder` 或 `[]byte` buffer：

```go
var buf bytes.Buffer
buf.WriteString("data: 当前服务器时间：")
buf.WriteString(time.Now().Format(time.RFC3339))
buf.WriteString("\n\n")
writer.Write(buf.Bytes())
```

---

## 总结

| # | 优先级 | 问题 | 影响 |
|---|--------|------|------|
| 1 | 高 | SSE 实例共享导致数据竞争 | 并发下状态错乱 |
| 2 | 高 | 缺少连接数限制 | 资源耗尽 |
| 3 | 高 | 缺少心跳与空闲超时 | goroutine 泄漏 |
| 4 | 中 | http.Flusher 断言缺失 | Writer 包装后可能 panic |
| 5 | 中 | 错误响应格式不规范 | 客户端无法区分错误类型 |
| 6 | 中 | 缺少 Accept 头校验 | 非 SSE 客户端误连接 |
| 7 | 低 | 手动拼 SSE 帧易出错 | 开发体验 |
| 8 | 低 | fmt.Sprintf 每帧分配 | GC 压力 |

---

## 生产就绪补充项

以下能力修复完前 8 条后仍缺失，需要在投入生产前补齐。

---

### 9. Last-Event-Id 断线续传

**问题:** SSE 标准规定：客户端断开重连时会在请求头中携带 `Last-Event-Id`，表示上次收到的事件 ID。服务端应从该 ID 之后开始推送，否则断线期间的事件永久丢失。目前框架不管理事件 ID，也不支持从指定位置开始重放。

**方案:**

1. `WriteSSE` 辅助方法支持 `id` 参数，写入 `id: {value}\n` 帧头
2. `SseHandle` 注册时提取 `Last-Event-Id` 请求头，注入到 `Context` 中供 `Serve()` 使用
3. 业务层自行实现消息队列回溯（如 Redis Stream），框架提供 `eventID` 的注入和提取能力

```go
// 框架层面：提取 Last-Event-Id 注入 Context
func (f *SseHandle) Get(apiPath string, sseInstance iCoreSse.ISse) {
    // ...
    f.engine.GET(apiPath, func(c *gin.Context) {
        lastEventID := c.GetHeader("Last-Event-Id")
        if lastEventID != "" {
            c.Set("sse.lastEventId", lastEventID)
        }
        // ... 创建实例, Serve
    })
}

// 用户在 Serve 中使用
func (e ExampleSse) Serve(ctx *gin.Context) error {
    lastID, _ := ctx.Get("sse.lastEventId")
    // 从 Redis / DB 中查 lastID 之后的事件开始推送
}
```

---

### 10. 优雅关闭通知

**位置:** `service/api/http_svc.go:63-107` — `Listen()` 方法

**问题:** 服务收到 SIGTERM 后，`Listen()` 只给 5 秒等待活跃请求完成就强制关停。SSE 长连接收不到任何通知，客户端表现为连接突然断开（报错而非优雅重连）。

**方案:**

1. `SseHandle` 维护活跃连接列表
2. 框架在 shutdown 时遍历所有活跃连接，发送关闭事件
3. 客户端收到关闭事件后主动重连

```go
type SseHandle struct {
    engine  *gin.Engine
    conns   map[*SSEConn]struct{}
    connMu  sync.Mutex
}

type SSEConn struct {
    Writer http.ResponseWriter
    Done   chan struct{}
}

// Listen 改造：先通知 SSE 连接关闭，再 Shutdown
func (f *ApiHandle) Listen() {
    // ...
    <-quit
    f.sseHandle.Shutdown() // 发关闭事件给所有 SSE 连接
    time.Sleep(2 * time.Second) // 等客户端收到后主动断开
    srv.Shutdown(ctx)
}

// SseHandle.Shutdown 发送关闭事件
func (f *SseHandle) Shutdown() {
    f.connMu.Lock()
    defer f.connMu.Unlock()
    for conn := range f.conns {
        fmt.Fprintf(conn.Writer, "event: shutdown\ndata: server is restarting\n\n")
        conn.Writer.(http.Flusher).Flush()
    }
}
```

---

### 11. CORS 支持

**问题:** 浏览器 `EventSource` API 遵循同源策略，跨域请求需要 `Access-Control-Allow-Origin` 头，且 `EventSource` 不支持自定义请求头（无法携带 token）。目前框架无内置 CORS 中间件。

**方案:**

1. 框架提供可复用的 CORS 中间件，或 SSE 路由默认设置 CORS 头
2. SSE handler 自动设置 `Access-Control-Allow-Origin` 和 `Access-Control-Allow-Credentials`

```go
// 在 SSE handler 中设置 CORS 头
func (f *SseHandle) Get(apiPath string, sseInstance iCoreSse.ISse) {
    f.engine.GET(apiPath, func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        // ... SSE 响应头
    })
}
```

> **注意:** `EventSource` 不支持自定义 header，跨域鉴权需通过 URL query 参数（如 `?token=xxx`）或在建立连接前通过其他 API 接口签发一次性 ticket。

---

### 12. 背压控制

**问题:** 服务端生产事件速度 > 客户端消费速度时，TCP 发送缓冲区填满后 `Write()` 会阻塞或失败。若业务层无条件循环写入，缓冲区无限增长导致 OOM，或写入速度失控导致其他连接饥饿。

**方案:**

1. 框架提供带缓冲和丢弃策略的写入封装
2. 非阻塞写入 + 丢弃最旧事件，避免拖慢整个事件循环

```go
// WriteSSEDrop 非阻塞写入，写入失败时跳过当前帧
func WriteSSEDrop(w io.Writer, flusher http.Flusher, event, data string) bool {
    var buf bytes.Buffer
    if event != "" {
        fmt.Fprintf(&buf, "event: %s\n", event)
    }
    fmt.Fprintf(&buf, "data: %s\n\n", data)

    _, err := w.Write(buf.Bytes())
    if err != nil {
        return false // 丢弃，不阻塞
    }
    flusher.Flush()
    return true
}
```

业务层自行判断是否需要丢弃或降速。

---

### 13. 可观测性指标

**问题:** 生产环境运行时无法回答：当前活跃 SSE 连接数？事件发送速率？连接断开原因分布？这些信息对排查问题和容量规划至关重要。目前完全不可见。

**方案:**

1. `SseHandle` 暴露指标接口或日志
2. 记录连接建立/断开事件，定期输出统计

```go
type SSEConn struct {
    ID        string
    StartTime time.Time
}

type SseHandle struct {
    engine     *gin.Engine
    activeConns map[string]*SSEConn
    mu         sync.RWMutex

    // 计数器
    totalConnects    atomic.Int64
    totalDisconnects atomic.Int64
}

func (f *SseHandle) Stats() map[string]interface{} {
    f.mu.RLock()
    defer f.mu.RUnlock()
    return map[string]interface{}{
        "active_connections":  len(f.activeConns),
        "total_connects":      f.totalConnects.Load(),
        "total_disconnects":   f.totalDisconnects.Load(),
    }
}
```

---

### 14. `Serve()` 调用前 Context 未注入 Trace

**位置:** `service/sse/sse_svc.go:39` — `sseInstance.Serve(c)`

**问题:** SSE handler 中 `Serve(c)` 的 `c` 是原始 `*gin.Context`。若 `NewSseSvc` 创建时 API 路由已注册了 `HttpTraceMiddleware`，SSE 路由的 trace span 在进入 handler 前已创建。但 `Serve()` 内部如果需要创建子 span（如追踪每条消息推送），需要从 `c.Request.Context()` 提取 span context，当前无封装。

**方案:** 框架在 `Serve` 调用前将 trace span 注入 Context，并提供便捷方法：

```go
// 在 SSE handler 中
span := iCoreTrace.FromContext(c.Request.Context())
c.Set("sse.traceSpan", span)

// 提供给 Serve 使用
func WriteSSEWithTrace(ctx *gin.Context, event, data string) {
    span, _ := ctx.Get("sse.traceSpan")
    // 创建子 span 追踪本次推送
}
```

---

## 总结（全部 14 条）

| # | 分类 | 优先级 | 问题 | 影响 |
|---|------|--------|------|------|
| 1 | 稳定性 | 高 | SSE 实例共享导致数据竞争 | 并发下状态错乱 |
| 2 | 稳定性 | 高 | 缺少连接数限制 | 资源耗尽 |
| 3 | 稳定性 | 高 | 缺少心跳与空闲超时 | goroutine 泄漏 |
| 4 | 稳定性 | 中 | http.Flusher 断言缺失 | Writer 包装后可能 panic |
| 5 | 稳定性 | 中 | 错误响应格式不规范 | 客户端无法区分错误类型 |
| 6 | 稳定性 | 中 | 缺少 Accept 头校验 | 非 SSE 客户端误连接 |
| 7 | 开发体验 | 低 | 手动拼 SSE 帧易出错 | 格式错误 |
| 8 | 性能 | 低 | fmt.Sprintf 每帧分配 | GC 压力 |
| **9** | **生产就绪** | **高** | **Last-Event-Id 断线续传** | **消息丢失** |
| **10** | **生产就绪** | **高** | **优雅关闭通知** | **重启时连接异常断开** |
| **11** | **生产就绪** | **中** | **CORS 支持** | **浏览器 EventSource 不可用** |
| **12** | **生产就绪** | **中** | **背压控制** | **OOM / 连接饥饿** |
| **13** | **生产就绪** | **低** | **可观测性指标** | **无法排障** |
| **14** | **生产就绪** | **低** | **Context 未注入 Trace** | **SSE 事件无链路追踪** |

### 阶段规划

```
第一阶段 (稳定性): #1-#6  + #9  → 不崩溃、不泄漏、不丢消息
第二阶段 (可运维): #10 + #13    → 可平滑发布、可监控
第三阶段 (可接入): #11 + #7     → 浏览器直接可用
第四阶段 (高性能): #8  + #12    → 高吞吐、抗背压
第五阶段 (可观测): #14          → 全链路追踪覆盖
```
