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
