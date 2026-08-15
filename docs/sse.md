# SSE 服务

Hecc-Blot 支持 Server-Sent Events（SSE），用于向客户端主动推送实时数据。SSE 服务与 API 服务共享同一个 Gin Engine 和端口。

## 接口定义

```go
// modules/core/contract/sse/sse.go
type ISse interface {
    Serve(ctx *gin.Context) error
}

// modules/core/contract/sse/sse_handler.go
type ISseHandle interface {
    Get(apiPath string, sse ISse)
    Middleware(middlewares ...iCoreApi.IMiddleware) ISseHandle
}
```

SSE 复用 `modules/core/contract/api` 中的 `IMiddleware` 接口，无需单独定义中间件接口。

## 初始化

SSE 服务与 API 服务共享 Engine，创建时传入 API 的 Engine：

```go
// 创建 API 处理器
apiHandle := api.NewApiSvc(&config.Server, responseSvc, traceSvc)

// 创建 SSE 处理器，共享 API 的 Engine
sseHandle := sse.NewSseSvc(apiHandle.Engine())

// 启动服务（仅 API 调用 Listen，SSE 共享同一端口）
apiHandle.Listen()
```

## 定义 SSE 端点

实现 `ISse` 接口：

```go
type ExampleSse struct {
    LogSvc iCoreLog.ILog `inject:""`
}

func (e ExampleSse) Serve(ctx *gin.Context) error {
    e.LogSvc.Info(ctx, "sse start")

    writer := ctx.Writer
    clientCtx := ctx.Request.Context()

    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-clientCtx.Done():
            return nil
        case <-ticker.C:
            msg := fmt.Sprintf("data: 当前服务器时间：%s\n\n", time.Now().Format(time.RFC3339))
            if _, err := writer.WriteString(msg); err != nil {
                return err
            }
            writer.Flush()
        }
    }
}
```

## 注册路由

```go
func registerSse(sseHandle iCoreSse.ISseHandle) {
    sseHandle.Middleware(&ReplayMiddleware{}, &TokenMiddleware{})
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

## 完整示例

```go
func main() {
    // ... 初始化日志、数据库、缓存、IOC 注册 ...

    apiHandle := api.NewApiSvc(&config.Server, responseSvc, traceSvc)

    // 注册 API 路由
    apiHandle.Middleware(&TokenMiddleware{})
    apiHandle.Post("example/api", &ExampleApi{})

    // 注册 SSE 路由（共享 Engine）
    sseHandle := sse.NewSseSvc(apiHandle.Engine())
    sseHandle.Middleware(&TokenMiddleware{})
    sseHandle.Get("example/sse", &ExampleSse{})

    // 启动服务
    apiHandle.Listen()
}
```

## SSE 响应头

框架自动设置以下 SSE 响应头，并立即 Flush 确保客户端及时识别流类型：

- `Content-Type: text/event-stream`
- `Cache-Control: no-cache`
- `Connection: keep-alive`

设置头后立即调用 `c.Writer.Flush()`，客户端无需等待第一个事件即可确认连接类型。

## 错误处理

Serve 返回 error 时，框架通过 SSE `error` 事件发送错误信息，不会中断流连接：

```go
if err := sseInstance.Serve(c); err != nil {
    c.SSEvent("error", err.Error())
}
```

## 注意事项

1. **共享端口**: SSE 与 API 共用同一 Gin Engine，只需调用 `apiHandle.Listen()` 一次
2. **长连接**: SSE 连接会一直保持直到客户端断开，Serve 方法需要监听 `ctx.Request.Context().Done()` 来感知断开
3. **流刷新**: 每次 Write 后需立即 `writer.Flush()` 确保数据即时推送到客户端
4. **中间件复用**: SSE 中间件与 API 中间件共用 `IMiddleware` 接口，无需单独定义
5. **依赖注入**: 与 API 一样，SSE 实例通过 IOC 自动注入依赖

## 相关文档

| 文档 | 说明 |
|------|------|
| [路由与中间件](routes_middleware.md) | API 路由与中间件复用 |
| [日志组件](logging.md) | SSE 连接日志 |
