# SSE 服务

Hecc-Blot 支持 Server-Sent Events（SSE），用于向客户端主动推送实时数据。SSE 服务与 API 服务共享同一个 Gin Engine 和端口。

## 接口定义

```go
// modules/core/contract/sse/sse.go
type ISse interface {
    Serve(ctx context.Context, w Writer) error
}

// Writer SSE 写入抽象，由框架实现并注入给业务
type Writer interface {
    Send(id, event, data string) error
    LastEventID() string
}

// modules/core/contract/sse/sse_handler.go
type ISseHandle interface {
    Get(apiPath string, sse ISse)
    Post(apiPath string, sse ISse)
    Middleware(middlewares ...iCoreApi.IMiddleware) ISseHandle
    Group(relativePath string, middlewares ...iCoreApi.IMiddleware) ISseHandle
}
```

SSE 复用 `modules/core/contract/api` 中的 `IMiddleware` 接口，无需单独定义中间件接口。

框架在底层封装了 `http.Flusher` 断言、心跳保活、并发锁、连接数上限与错误帧格式化，业务只需通过 `Writer` 写入、通过 `ctx` 感知连接断开。策略性校验（如 `Accept` 头）建议通过中间件实现。

## 初始化

SSE 服务与 API 服务共享 Engine，创建时传入 API 的 Engine：

```go
// 创建 API 处理器
apiHandle := api.NewApiSvc(&config.Server, responseSvc, traceSvc, container)

// 创建 SSE 处理器，共享 API 的 Engine
sseHandle := sse.NewSseSvc(apiHandle.Engine(), container)

// 启动服务（仅 API 调用 Listen，SSE 共享同一端口）
apiHandle.Listen()
```

## 定义 SSE 端点

实现 `ISse` 接口：

```go
type ExampleSse struct {
    LogSvc iCoreLog.ILog `inject:""`
}

func (e ExampleSse) Serve(ctx context.Context, w iCoreSse.Writer) error {
    e.LogSvc.Info(ctx, "sse start")

    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            // 客户端断开或心跳写入失败
            return nil
        case <-ticker.C:
            msg := fmt.Sprintf("当前服务器时间：%s", time.Now().Format(time.RFC3339))
            if err := w.Send("", "", msg); err != nil {
                return err
            }
        }
    }
}
```

## 注册路由

```go
func registerSse(sseHandle iCoreSse.ISseHandle) {
    sseHandle.Middleware(&TokenMiddleware{})
    {
        sseHandle.Get("example/sse", &ExampleSse{})
    }
}
```

### 路由注册原理

SSE 处理器接口定义在 `modules/core/contract/sse/sse_handler.go`：

```go
type ISseHandle interface {
    Get(apiPath string, sse ISse)
    Post(apiPath string, sse ISse)
    Middleware(middlewares ...iCoreApi.IMiddleware) ISseHandle
    Group(relativePath string, middlewares ...iCoreApi.IMiddleware) ISseHandle
}
```

框架自动设置 SSE 响应头（`Content-Type: text/event-stream`、`Cache-Control: no-cache`、`Connection: keep-alive`），设置后立即 Flush 确保客户端及时识别流类型。

**请求处理流程：**

```
请求 → [中间件链] → [Flusher 断言] → [连接数限流] → [设置SSE响应头] → [Serve()] → [持续推送事件]
                                                                              ↓
                                                                   客户端断开 / 心跳失败 / Serve 返回 error
                                                                              ↓
                                                               发送 error SSE 事件（JSON）→ 结束
```

**框架自动处理的能力：**

- **实例隔离**: 每个连接创建独立业务实例，避免并发共享
- **连接数上限**: 超出上限返回 503
- **心跳保活**: 每 30s 发送 SSE comment，写入失败自动取消连接
- **Flusher 断言**: Writer 不支持流式时返回 500
- **Accept 校验**: 建议通过中间件实现（见 `example.go` 的 `SseAcceptMiddleware`），框架不内置
- **Last-Event-Id**: 提取请求头，通过 `w.LastEventID()` 暴露给业务做断线续传

## 完整示例

```go
func main() {
    // ... 初始化日志、数据库、缓存、IOC 注册 ...

    container := ioc.New()
    // ... container.Set(...) ...

    apiHandle := api.NewApiSvc(&config.Server, responseSvc, traceSvc, container)

    // 注册 API 路由
    apiHandle.Middleware(&TokenMiddleware{})
    apiHandle.Post("example/api", &ExampleApi{})

    // 注册 SSE 路由（共享 Engine）
    sseHandle := sse.NewSseSvc(apiHandle.Engine(), container)
    sseHandle.Middleware(&TokenMiddleware{})
    sseHandle.Get("example/sse", &ExampleSse{})

    // 启动服务
    apiHandle.Listen()
}
```

## 错误处理

Serve 返回 error 时，框架以 JSON 格式发送 SSE `error` 事件，不会中断流连接：

```
event: error
data: {"code":"500","message":"..."}
```

客户端可据此区分错误类型并做相应处理。

## 注意事项

1. **共享端口**: SSE 与 API 共用同一 Gin Engine，只需调用 `apiHandle.Listen()` 一次
2. **长连接**: SSE 连接会一直保持直到客户端断开或心跳失败，Serve 方法监听 `ctx.Done()` 感知断开
3. **写入**: 业务通过 `w.Send()` 写入，框架自动 Flush 并保证并发安全
4. **中间件复用**: SSE 中间件与 API 中间件共用 `IMiddleware` 接口，无需单独定义
5. **依赖注入**: 与 API 一样，SSE 实例通过 IOC 自动注入依赖

## 相关文档

| 文档 | 说明 |
|------|------|
| [路由与中间件](routes_middleware.md) | API 路由与中间件复用 |
| [日志组件](logging.md) | SSE 连接日志 |
