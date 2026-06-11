# 数据库组件

Hecc-Blot 数据库组件基于 GORM，同时支持 MySQL 和 PostgreSQL，提供链式查询、事务、多数据库动态切换。

## 接口定义

```go
type IDbFactory interface {
    Build(ctx context.Context, v ...dbEnum.Value) IDb
    SetDefault(t dbEnum.Value)
}

type IDb interface {
    Add(entry IDbModel) error
    Remove(entry IDbModel) error
    Query(entry IDbModel) IDb
    Save(entry IDbModel) error
    Count() (int64, error)
    Order(fields ...string) IDb
    Select(args ...interface{}) IDb
    Offset(v int) IDb
    Limit(v int) IDb
    Where(args ...interface{}) IDb
    Take(dst interface{}) error
    Find(dst interface{}) error
    WithContext(ctx context.Context)
    Begin() IDb
    Commit() error
    Rollback()
    GetInstance() any
}
```

## 初始化

```go
dbFactory, clearUp, err := db.NewDbFactory(&config.Db, logSvc)
if err != nil {
    panic(err)
}
defer clearUp()

// 注册到 IOC
ioc.Set(new(iCoreDb.IDbFactory), dbFactory)
```

## Model 定义

Model 需嵌入 GORM Model 并实现 `IDbModel` 接口：

```go
type AccountModel struct {
    ID          int    `json:"id" gorm:"primaryKey"`
    AccountName string `json:"account_name"`
    Password    string `json:"password"`
    CreatedAt   int    `json:"created_at"`
    UpdatedAt   int    `json:"updated_at"`
    DeletedAt   int    `json:"deleted_at"`
}

func (a AccountModel) TableName() string {
    return "account"
}

func (a AccountModel) GetID() int {
    return a.ID
}
```

`GetID()` 方法用于框架获取主键，`TableName()` 指定表名。

## CRUD 操作

### Add — 添加记录

```go
newAccount := AccountModel{AccountName: "test"}
err := mysqlSvc.Add(&newAccount)
// newAccount.ID 会被自动填充
```

### Find — 查询多条

```go
data := new(make([]AccountModel, 0))
err := mysqlSvc.
    Where("id >= ? AND id <= ?", 1, 8).
    Find(data)
```

### Find 分页

```go
data := new(make([]AccountModel, 0))
err := mysqlSvc.
    Where("id >= ?", 1).
    Offset(0).
    Limit(10).
    Find(data)
```

### Take — 查询单条

```go
data := AccountModel{}
err := mysqlSvc.
    Where("id = ?", 1).
    Take(&data)
```

### Select — 指定字段

```go
data := AccountModel{}
err := mysqlSvc.
    Select("id, account_name").
    Where("id = ?", 1).
    Take(&data)
```

### Save — 更新记录

```go
updateData := AccountModel{AccountName: "updated"}
err := mysqlSvc.Where("id = ?", 1).Save(&updateData)
```

### Count — 统计

```go
count, err := mysqlSvc.Query(&AccountModel{}).
    Where("id >= ?", 1).
    Count()
```

### Order — 排序

```go
data := new(make([]AccountModel, 0))
err := mysqlSvc.
    Select("id, account_name").
    Where("id >= ?", 1).
    Order("id DESC").
    Find(data)
```

### Remove — 删除

```go
err := mysqlSvc.Where("id = ?", 1).Remove(&AccountModel{})
```

## 事务

`Begin()` 返回一个新的 `IDb` 事务实例，原始实例不受影响。事务内的所有操作应在返回的实例上执行。

```go
func (a AddApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
    db := a.DbFactory.Build(ctx)

    // 开启事务
    tx := db.Begin()

    // 事务内操作——都在 tx 上执行
    newAccount := AccountModel{AccountName: "test"}
    if err := tx.Add(&newAccount); err != nil {
        tx.Rollback()
        return nil, errorSvc.NewError(response.Fail, err)
    }

    // 更新也在事务内
    updateData := AccountModel{Password: "new-password"}
    if err := tx.Where("id = ?", newAccount.ID).Save(&updateData); err != nil {
        tx.Rollback()
        return nil, errorSvc.NewError(response.Fail, err)
    }

    // 提交事务
    if err := tx.Commit(); err != nil {
        return nil, errorSvc.NewError(response.Fail, err)
    }

    return newAccount, nil
}
```

**注意事项：**

- `Begin()` 返回新实例，必须在返回的 `tx` 上操作，不能用原始的 `db`
- `Rollback()` 和 `Commit()` 只在事务实例上调用
- `Commit()` 或 `Rollback()` 后事务实例不应再使用
- 原始 `db` 实例不受事务影响，事务结束后仍可继续使用

## 多数据库切换

框架支持同时配置 MySQL 和 PostgreSQL：

```go
// 初始化工厂（自动连接所有已配置的数据库）
dbFactory, clearUp, err := db.NewDbFactory(&config.Db, logSvc)

// 设置默认数据库（默认使用 MySQL）
dbFactory.SetDefault(dbEnum.Postgres)

// 不带参数使用默认数据库
db := dbFactory.Build(ctx)

// 运行时指定数据库类型
mysqlDB := dbFactory.Build(ctx, dbEnum.Mysql)
pgDB := dbFactory.Build(ctx, dbEnum.Postgres)
```

## 在 API 中使用

通过 IOC 注入 `IDbFactory`，每次请求从 Context 获取数据库实例：

```go
type ListApi struct {
    DbFactory iCoreDb.IDbFactory `inject:""`
}

func (a ListApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
    db := a.DbFactory.Build(ctx)
    data := new(make([]AccountModel, 0))
    if err := db.Where("id >= ?", 1).Find(data); err != nil {
        return nil, errorSvc.NewError(response.Fail, err)
    }
    return data, nil
}
```

## 相关文档

| 文档 | 说明 |
|------|------|
| [配置说明](config.md) | 数据库连接配置项 |
| [IOC 注入](ioc_injection.md) | 注入 IDbFactory |
| [缓存组件](cache.md) | 缓存与数据库配合的读穿透模式 |
