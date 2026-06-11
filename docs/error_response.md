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
