# 链路追踪

框架基于 OpenTelemetry 实现分布式追踪，支持通过 OTLP 协议导出数据到 Jaeger 等追踪系统。

## 配置说明

在 `config.yaml` 中添加 trace 配置：

```yaml
trace:
  service_name: Hecc-Blot            # 服务名称
  endpoint: 127.0.0.1:4318              # OTLP 接收端点 (HTTP)
  sampler:
    type: always                        # 采样类型: always/never/probability
    ratio: 0.5                          # 采样比例 (probability 模式使用)
```

### 配置项说明

| 配置项 | 类型 | 说明 |
|--------|------|------|
| `service_name` | string | 服务名称，用于在追踪系统中标识 |
| `endpoint` | string | OTLP HTTP 接收端点地址 |
| `sampler.type` | string | 采样类型：`always`/`never`/`probability` |
| `sampler.ratio` | float | 采样比例，仅 probability 模式生效 (0-1) |

### 采样类型

| 类型 | 说明 |
|------|------|
| `always` | 采样所有请求，适用于开发环境 |
| `never` | 不采样任何请求，适用于调试场景 |
| `probability` | 按比例采样，`ratio` 指定采样率 (如 0.5 表示 50%) |


## 代码使用

### 初始化追踪服务

```go
import (
    iCoreTrace "github.com/bestHeCC/hecc-core/contract/trace"
    "github.com/bestHeCC/hecc-trace"
)

// 初始化
traceSvc, traceClearUp, err := trace.NewTraceSvc(&config.Trace)
if err != nil {
    allErrors = append(allErrors, err)
}

defer traceClearUp()
```

### 在业务中使用 Trace

框架提供 `ITrace` 接口，支持在业务代码中创建 Span 和记录追踪数据：

```go
type YourApi struct {
    // 通过 inject tag 自动注入
    TraceSvc iCoreTrace.ITrace `inject:""`
}

func (y YourApi) Call(ctx *gin.Context) (interface{}, error) {
    // 从 Context 获取当前活跃的 Span
    currentSpan := y.TraceSvc.FromContext(ctx)
    
    // 为当前 Span 添加自定义属性
    currentSpan.SetAttribute("user.id", 12345)
    currentSpan.SetAttribute("operation.type", "query")
    
    // 记录业务错误
    if err != nil {
        currentSpan.RecordError(err)
    }
    
    // 开启子 Span 追踪子操作
    subCtx, subSpan := y.TraceSvc.Start(ctx, "sub-operation",
        "sub.key", "sub-value",
    )
    defer subSpan.End()
    
    // 执行子操作
    result := doSomething(subCtx)
    
    subSpan.End()
    return result, nil
}
```

### Span 操作

| 方法 | 说明 |
|------|------|
| `SetAttribute(key, value)` | 设置 Span 属性，支持 string/int/int64/bool/float64 |
| `RecordError(err)` | 记录错误信息到当前 Span |
| `Name()` | 获取 Span 名称 |
| `End()` | 结束当前 Span |

## 全局中间件实现自动追踪

### HttpTraceMiddleware

框架提供默认 `HttpTraceMiddleware`，自动追踪所有 HTTP 请求。创建 API 服务时传入 `traceSvc` 即可开启链路追踪。

### 使用方式

```go
// 创建 API 服务时显式传入 traceSvc
apiHandle := api.NewApiSvc(&config.Server, responseSvc, traceSvc)

// 注册路由
register(apiHandle)

// 启动服务
apiHandle.Listen()
```

### 自动行为

`HttpTraceMiddleware` 会自动执行以下操作：

1. **链路上下文提取**：从请求头 `traceparent` 中提取分布式追踪上下文，实现跨服务链路关联
2. **创建请求 Span**：为每个 HTTP 请求创建 `http.request` Span，请求路径记录在 `http.url` 属性中
3. **提取 Trace ID**：从 Span 中获取 Trace ID
4. **响应头注入**：
   - `X-Trace-Id`: 当前请求的 Trace ID
   - `traceparent`: W3C Trace Context 格式的追踪上下文

## 日志集成

追踪服务与日志服务深度集成，自动将 TraceId 关联到日志中：

```go
func (y YourApi) Call(ctx *gin.Context) (interface{}, error) {
    // 日志会自动包含 traceId 字段
    y.LogSvc.Info(ctx, "执行查询", "table", "users")
    // 输出: {"level":"info","msg":"执行查询","table":"users","traceId":"4bf92f35..."}
}
```

## 上下文传递

### 在 HTTP 服务间传递

通过 HTTP 头传递 `traceparent`：

```go
// 发送方
req, _ := http.NewRequest("POST", "http://service-b/api", body)
req.Header.Set("traceparent", c.GetHeader("traceparent"))

// 接收方自动通过中间件提取
```

## 相关文档

| 文档 | 说明 |
|------|------|
| [配置说明](config.md) | Trace 配置项 |
| [日志组件](logging.md) | TraceId 自动关联日志 |
| [路由与中间件](routes_middleware.md) | HttpTraceMiddleware 自动追踪 |