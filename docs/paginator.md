# 分页组件

Hecc-Blot 提供两种分页模式：**Offset/Limit 分页**和**游标分页**，位于 `util/paginator.go`。

## Offset/Limit 分页

传统页码分页，适用于需要显示总页数、支持跳页的管理后台场景。

### 结构体

```go
type PageOpts struct {
    Page     int
    PageSize int
}

type Paginator[T any] struct {
    List     []T   `json:"list"`
    Page     int   `json:"page"`
    PageSize int   `json:"pageSize"`
    Total    int64 `json:"total"`
}
```

### 使用示例

```go
type PageRequest struct {
    Page     int `json:"page" form:"page"`
    PageSize int `json:"pageSize" form:"pageSize"`
}

type PageListApi struct {
    DbFactory iCoreDb.IDbFactory `inject:""`
    PageRequest
}

func (e PageListApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
    opts := util.PageOpts{Page: e.Page, PageSize: e.PageSize}
    db := e.DbFactory.Build(ctx).Query(AccountModel{})

    // 查总数
    total, err := db.Count()
    if err != nil {
        return nil, errorSvc.NewError(response.Fail, err)
    }

    // 查当前页
    var list []AccountModel
    offset := (opts.Page - 1) * opts.PageSize
    err = db.Order("id desc").Limit(opts.PageSize).Offset(offset).Find(&list)
    if err != nil {
        return nil, errorSvc.NewError(response.Fail, err)
    }

    return util.NewPage(list, total, opts), nil
}
```

### 请求

```
GET /example/page?page=1&pageSize=10
```

### 响应

```json
{
    "code": 10000,
    "message": "请求成功",
    "data": {
        "list": [
            {"id": 10, "account_name": "user10", ...},
            {"id": 9,  "account_name": "user9",  ...}
        ],
        "page": 1,
        "pageSize": 10,
        "total": 100
    }
}
```

### 注意事项

- `Page` 不传时默认为 1，`PageSize` 不传时默认为 10
- `list` 为空时返回 `[]` 而非 `null`
- 翻页越深性能越差，数据量大时考虑游标分页

---

## 游标分页

基于游标（通常是主键 ID 或时间戳）的分页，适用于无限滚动、实时数据流等场景。

### 结构体

```go
type CursorOpts struct {
    Cursor   any // 上一页最后一条记录的游标值，首页传 nil
    PageSize int
}

type Cursor[T any] struct {
    List       []T  `json:"list"`
    NextCursor any  `json:"nextCursor"`
    HasMore    bool `json:"hasMore"`
    PageSize   int  `json:"pageSize"`
}
```

### 核心约定

**查询时多取一条（`pageSize + 1`）**，`NewCursor` 自动判断 `hasMore` 并截断多余数据。

```
查询 LIMIT = pageSize + 1
  ├── len(list) > pageSize → hasMore = true, 取前 pageSize 条, nextCursor = 最后一条的游标
  └── len(list) ≤ pageSize → hasMore = false, nextCursor = nil
```

### 使用示例

```go
type CursorRequest struct {
    Cursor   int `json:"cursor" form:"cursor"`
    PageSize int `json:"pageSize" form:"pageSize"`
}

type CursorListApi struct {
    DbFactory iCoreDb.IDbFactory `inject:""`
    CursorRequest
}

func (e CursorListApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
    pageSize := e.PageSize
    cursor := e.Cursor

    db := e.DbFactory.Build(ctx).Query(AccountModel{})

    // 多查一条用于判断 hasMore
    var list []AccountModel
    err := db.Where("id > ?", cursor).Order("id asc").Limit(pageSize + 1).Find(&list)
    if err != nil {
        return nil, errorSvc.NewError(response.Fail, err)
    }

    return util.NewCursor(list, pageSize, func(item *AccountModel) any {
        return item.ID
    }), nil
}
```

### 请求

```
首次请求（无 cursor）:
  GET /example/cursor?pageSize=10

翻页（带上上次返回的 nextCursor）:
  GET /example/cursor?pageSize=10&cursor=10
```

### 响应

```json
{
    "code": 10000,
    "message": "请求成功",
    "data": {
        "list": [
            {"id": 1,  "account_name": "user1",  ...},
            {"id": 2,  "account_name": "user2",  ...}
        ],
        "nextCursor": 10,
        "hasMore": true,
        "pageSize": 10
    }
}
```

最后页：

```json
{
    "data": {
        "list": [...],
        "nextCursor": null,
        "hasMore": false,
        "pageSize": 10
    }
}
```

### 客户端翻页流程

```
1. 首次请求无 cursor → 拿到 nextCursor=10, hasMore=true
2. 用 cursor=10 请求下一页 → 拿到 nextCursor=20, hasMore=true
3. 用 cursor=20 请求下一页 → 拿到 nextCursor=null, hasMore=false → 停止
```

---

## 对比

| | Offset/Limit | 游标 |
|--|-------------|------|
| 入参 | `page` + `pageSize` | `cursor` + `pageSize` |
| SQL | `OFFSET x LIMIT y` | `WHERE id > cursor LIMIT y+1` |
| 深翻页性能 | 递减（需扫描前 N 行） | 稳定（走索引） |
| 总页数 | 支持 | 不支持 |
| 跳页 | 支持 | 不支持 |
| 新增数据影响 | 可能重复或漏数据 | 不受影响 |
| 适用场景 | 管理后台、报表 | 信息流、无限滚动 |

## 相关文档

| 文档 | 说明 |
|------|------|
| [数据库组件](database.md) | Offset/Limit 和 WHERE 游标查询 |
| [路由与中间件](routes_middleware.md) | 注册分页 API |
