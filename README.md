# Hecc-Go-Core

Go核心框架，提供HTTP服务、日志管理、IoC容器和数据库操作功能。

## 技术栈

- **HTTP**: Gin
- **日志**: Zap + 阿里云SLS
- **MySQL**: GORM
- **Redis**: go-redis
- **MongoDB**: mongo-driver

## 目录结构

```
core/
├── contract/          # 接口定义目录
│   ├── logger.go      # 日志接口
│   ├── container.go   # IoC容器接口
│   ├── database.go    # 数据库接口(MySQL/Redis/MongoDB)
│   └── http_server.go # HTTP服务接口
├── enum/              # 枚举目录
│   └── log_type.go    # 日志类型枚举
├── service/           # 服务实现目录
│   ├── logger_service.go      # 日志服务实现
│   ├── container_service.go   # IoC容器实现
│   ├── mysql_service.go       # MySQL服务实现
│   ├── redis_service.go       # Redis服务实现
│   ├── mongodb_service.go     # MongoDB服务实现
│   └── http_server_service.go # HTTP服务实现
├── go.mod
└── README.md
```

## 功能特性

### 1. 日志服务
- 支持本地日志和阿里云SLS上报
- 支持多种日志级别(Debug/Info/Warn/Error/Panic/Fatal)
- 基于Zap实现高性能日志

### 2. IoC容器
- 支持依赖注入
- 支持单例和工厂模式
- 线程安全

### 3. 数据库支持
- MySQL: 基于GORM
- Redis: 基于go-redis
- MongoDB: 基于mongo-driver

### 4. HTTP服务
- 基于Gin框架
- 支持中间件
- 支持路由注册

## 使用示例

### 初始化日志服务

```go
import (
    "core/enum"
    "core/service"
)

logger, _ := service.NewLoggerService(service.LoggerConfig{
    LogType:     enum.LogTypeLocal,
    Level:       enum.LogLevelInfo,
    LocalPath:   "./logs/app.log",
})
logger.Info("Hello World")
```

### 初始化IoC容器

```go
container := service.NewContainerService()
container.Bind("logger", func() interface{} {
    logger, _ := service.NewLoggerService(service.LoggerConfig{
        LogType: enum.LogTypeLocal,
        Level:   enum.LogLevelInfo,
    })
    return logger
})
```

### 初始化MySQL

```go
mysql, _ := service.NewMySQLService(service.MySQLConfig{
    Host:     "localhost",
    Port:     3306,
    Username: "root",
    Password: "password",
    Database: "test",
    Charset:  "utf8mb4",
})
```

### 初始化Redis

```go
redis, _ := service.NewRedisService(service.RedisConfig{
    Host:     "localhost",
    Port:     6379,
    Password: "",
    DB:       0,
})
```

### 初始化MongoDB

```go
mongo, _ := service.NewMongoDBService(service.MongoDBConfig{
    Host:     "localhost",
    Port:     27017,
    Username: "",
    Password: "",
    Database: "test",
})
```

### 初始化HTTP服务

```go
server := service.NewHTTPServerService()
server.AddRoute(contract.Route{
    Method: http.MethodGet,
    Path:   "/hello",
    Handler: func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello World"))
    },
})
server.Start(":8080")
```

## 许可证

MIT License
