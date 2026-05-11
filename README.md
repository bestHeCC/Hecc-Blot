# Hecc-Go-Core

Go后端核心框架，提供HTTP服务、日志管理、IoC容器、数据库和缓存操作功能。

## 技术栈

| 类别 | 技术 | 说明 | future       |
|------|------|------|--------------|
| HTTP | Gin | 高性能Web框架 | 支持SSE和WS     |
| 日志 | Zap + Lumberjack | 本地日志 + 阿里云SLS可选 | -            |
| 数据库 | GORM | MySQL ORM | 支持MongoDB和ES |
| 缓存 | go-redis + 本地缓存 | Redis + 内存缓存 | -            |
| 配置 | Viper | 配置管理 | -            |

## 目录结构

```
Hecc-Go-Core/
├── contract/             # 接口定义(面向接口编程)
│   ├── api/              # API服务接口
│   ├── cache/            # 缓存接口
│   ├── db/               # 数据库接口
│   ├── error/            # 错误定义接口
│   └── log/              # 日志接口
├── entity/               # 实体层
│   ├── api/              # API实体
│   └── config/           # 配置结构体
├── enum/                 # 枚举定义
│   ├── db/               # 数据库相关枚举
│   ├── env/              # 环境枚举(Dev/Test/Prod)
│   ├── response/         # 响应状态码枚举
│   └── trace/            # 链路追踪枚举
├── service/              # 服务层实现
│   ├── api/              # HTTP服务
│   ├── cache/            # 缓存服务(Redis/本地)
│   ├── db/               # 数据库服务
│   ├── error/            # 错误服务
│   ├── ioc/              # IoC依赖注入容器
│   └── log/              # 日志服务
└── example.go            # 使用示例
```

## 快速开始
完整示例可以参考[example.go](example.go)  
各组件使用示例可以参考单测，单测在service目录下对应组件目录中可以找到  
这里以[mysqlSvc_test.go](service/db/mysqlSvc_test.go)为例

### 1. 项目初始化

```go
package main

import (
    iCoreApi "core/contract/api"
    iCoreCache "core/contract/cache"
    iCoreDb "core/contract/db"
    iCoreLog "core/contract/log"
    coreConfig "core/entity/config"
    "core/service/api"
    "core/service/cache"
    "core/service/db"
    "core/service/ioc"
    "core/service/log"
    "github.com/spf13/viper"
)

func main() {
    var allErrors []error

    // 1. 加载配置
    config, err := initConf("/config.yaml")
    if err != nil {
        allErrors = append(allErrors, err)
    }

    // 2. 初始化日志服务
    logSvc, err := log.NewLogger(config)
    if err != nil {
        allErrors = append(allErrors, err)
    }

    // 3. 初始化数据库工厂
    dbFactory, dbCleanUp, err := db.NewDbFactory(config, logSvc)

    // 4. 初始化缓存工厂
    cacheFactory := cache.NewCacheFactory(&config.Cache)

    // 5. 初始化响应服务
    responseSvc := api.NewResponseSvc()

    // 错误统一处理
    if len(allErrors) > 0 {
        panic(fmt.Errorf("初始化错误：%v", allErrors))
    }

    // 程序退出时清理资源
    defer func() {
        if dbFactory != nil {
            dbCleanUp()
        }
    }()

    // 6. 注册服务到IoC容器
    ioc.Set(new(iCoreDb.IDbFactory), dbFactory)
    ioc.Set(new(iCoreLog.ILog), logSvc)
    ioc.Set(new(iCoreCache.ICacheFactory), cacheFactory)
    ioc.Set(new(iCoreApi.IResponse), responseSvc)

    // 7. 启动HTTP服务
    apiHandle := api.NewApiSvc(config, responseSvc)
    register(apiHandle)
    apiHandle.Listen()
}
```

### 2. 定义中间件

```go
// TokenMiddleware Token验证中间件
type TokenMiddleware struct {
    ResponseSvc iCoreApi.IResponse `inject:""`
}

func (t TokenMiddleware) Middleware() interface{} {
    return func(c *gin.Context) {
        u, err := strconv.ParseUint(c.GetHeader("id"), 10, 64)
        if err != nil {
            t.ResponseSvc.Regular(c, nil, errorSvc.NewError(response.TokenInvalid, err))
            c.Abort()
        }
        c.Set("id", u)
        c.Next()
    }
}
```

### 3. 定义Model

```go
type AccountModel struct {
    ID          int    `json:"id" gorm:"primaryKey"`
    AccountName string `json:"account_name"`
    Password    string `json:"password"`
    CreatedAt   int64  `json:"created_at"`
    UpdatedAt   int64  `json:"updated_at"`
    DeletedAt   int64  `json:"deleted_at"`
}

func (b AccountModel) TableName() string {
    return "account"
}

func (b AccountModel) GetID() int {
    return b.ID
}
```

### 4. 定义API

```go
type AddRequest struct {
    Name string `json:"name" binding:"required"`
}

// 自定义校验错误信息
func (a AddRequest) GetMessages() entityApi.Messages {
    return entityApi.Messages{
        "Name.required": "名字不能为空",
    }
}

type AddApi struct {
    // inject字段在前，请求参数在后
    DbFactory iCoreDb.IDbFactory `inject:""`
    LogSvc    iCoreLog.ILog      `inject:""`
    AddRequest
}

func (a AddApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
    newAccount := AccountModel{
        AccountName: a.Name,
    }

    a.LogSvc.Info(ctx, "add account", "account", newAccount)

    err := a.DbFactory.Build(ctx).Add(&newAccount)
    if err != nil {
        return nil, errorSvc.NewError(response.Fail, err)
    }

    return newAccount, nil
}
```

### 5. 注册路由

```go
func register(apiHandle iCoreApi.IApiHandle) {
    // 注册中间件
    apiHandle.Middleware(&ReplayMiddleware{}, &TokenMiddleware{})

    // 注册API（自动注入、自动校验、自动响应包装）
    apiHandle.Post("account/add", &AddApi{})
}
```

## 核心特性

### IoC容器
通过 `inject` 标签实现依赖注入，路由注册时自动注入。

### 自动特性
| 特性 | 说明 |
|------|------|
| 自动注入 | 路由注册时自动注入 `inject` 标签的字段 |
| 自动校验 | 请求参数自动根据 `binding` 标签校验 |
| 自动响应 | 返回值自动包装为 `{code: 0, msg: "success", data: {}}` |

### 数据库操作
```go
// 使用DbFactory操作数据库
a.DbFactory.Build(ctx).Add(&model)          // 添加
a.DbFactory.Build(ctx).Find(&results)       // 查询
a.DbFactory.Build(ctx).Save(&model)         // 更新
a.DbFactory.Build(ctx).Remove(&model)       // 删除
a.DbFactory.Build(ctx).Count(&model)        // 统计
a.DbFactory.Build(ctx).Order(&model)        // 排序
a.DbFactory.Build(ctx).Select(&model)       // 查询指定字段
a.DbFactory.Build(ctx).Offset(&model)       // 偏移量
a.DbFactory.Build(ctx).Limit(&model)        // 数据量
a.DbFactory.Build(ctx).Where(&model)        // 拼接where条件
```

## 许可证

MIT License
