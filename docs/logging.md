# 日志组件

Hecc-Blot 日志组件基于 Zap + lumberjack，支持本地文件日志和阿里云 SLS 日志服务。

## 接口定义

```go
type ILog interface {
    Debug(ctx context.Context, msg string, fields ...interface{})
    Info(ctx context.Context, msg string, fields ...interface{})
    Warn(ctx context.Context, msg string, fields ...interface{})
    Error(ctx context.Context, msg string, fields ...interface{})
}
```

## 初始化

```go
logSvc, err := log.NewLogger(&config.Log)
if err != nil {
    panic(err)
}

// 注册到 IOC 容器
container := ioc.New()
container.Set(new(iCoreLog.ILog), logSvc)
```

## 本地日志

基于 Zap + lumberjack 实现日志文件自动滚动，配置 enable 为 true 后自动启用

### 配置

```yaml
log:
  local:
    enable: true
    root_dir: runtime/logs
    max_size: 100       # 单文件最大 100MB
    max_backups: 30     # 最多保留 30 个旧文件
    max_age: 7          # 日志保留 7 天
    compress: true      # 压缩旧文件
```

### 日志级别与文件

日志按级别分文件输出在 `root_dir` 下：

| 文件 | 级别 | 说明 |
|------|------|------|
| `debug.log` | Debug | 调试信息 |
| `info.log` | Info | 一般信息 |
| `warn.log` | Warn | 警告信息 |
| `error.log` | Error | 错误信息（含 stacktrace） |
| `panic.log` | Panic | 恐慌信息（含 stacktrace） |

### 使用

```go
// 基本日志
a.LogSvc.Debug(ctx, "processing request", "user_id", 123)
a.LogSvc.Info(ctx, "user login success", "ip", c.ClientIP())
a.LogSvc.Warn(ctx, "rate limit approaching", "count", 950)
a.LogSvc.Error(ctx, "database connection failed", zap.Error(err))

// fields 支持 zapcore.Field 类型
a.LogSvc.Info(ctx, "order created",
    zap.String("order_id", orderID),
    zap.Int("amount", 100),
    zap.Duration("elapsed", elapsed),
)
```

### TraceId 自动关联

框架从 Context 中自动提取 TraceId，日志输出会自动包含 `traceId` 和 `caller` 字段：

```json
{
    "level": "info",
    "msg": "user login",
    "traceId": "4bf92f3577b34da6a3ce929d0e0e4736",
    "caller": "example/example.go:209",
    "user_id": 123
}
```

- `traceId`: 从 Context 自动提取，关联链路追踪
- `caller`: Zap 自动记录调用日志的文件和行号

## SLS 日志

阿里云日志服务（SLS）支持，配置 enable 为 true 后自动启用：

```yaml
log:
  sls:
    enable: true
    endpoint: "cn-hangzhou.log.aliyuncs.com"
    access_key_id: "your-access-key"
    access_key_secret: "your-secret"
    project: "your-project"
    logstore: "your-logstore"
```

## 注意
日志组件实例时在本地日志或阿里SLS日志服务中二选一，通过相应的配置enable进行启用，若都开启，则优先使用阿里SLS

## 在 API 中使用

通过 IOC 注入 `ILog`：

```go
type AddApi struct {
    LogSvc    iCoreLog.ILog       `inject:""`
    DbFactory iCoreDb.IDbFactory  `inject:""`
    AddRequest
}

func (a AddApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
    a.LogSvc.Info(ctx, "add account request", "name", a.AccountName)

    err := a.DbFactory.Build(ctx).Add(&data)
    if err != nil {
        a.LogSvc.Error(ctx, "add account failed", zap.Error(err))
        return nil, errorSvc.NewError(response.Fail, err)
    }

    a.LogSvc.Info(ctx, "add account success", "id", data.ID)
    return data, nil
}
```

## 自定义日志实现

如需替换为其他日志库（如 logrus），只需实现 `ILog` 接口并注册到 IOC：

```go
type LogrusLogSvc struct {
    logger *logrus.Logger
}

func (l LogrusLogSvc) Debug(ctx context.Context, msg string, fields ...interface{}) {
    l.logger.Debug(msg, fields...)
}
// ... 实现 Info / Warn / Error

container := ioc.New()
container.Set(new(iCoreLog.ILog), &LogrusLogSvc{logger: logrus.New()})
```

## 相关文档

| 文档 | 说明 |
|------|------|
| [配置说明](config.md) | 日志配置项 |
| [链路追踪](trace.md) | TraceId 与日志关联 |
| [组件替换](component_replacement.md) | 替换为 logrus 等第三方库 |
