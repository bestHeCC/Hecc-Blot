# 配置说明

框架使用 `config.yaml` 作为配置文件，通过 Viper 加载。以下为全部配置项说明。

## 完整示例

```yaml
server:
  port: "9500"        # 服务端口
  env: dev            # 运行环境：dev / test / product
  name: Hecc-Blot     # 服务名称
  enable_trace: true  # 是否开启链路追踪
  read_timeout: 30    # 读取超时（秒），0 默认 30
  write_timeout: 30   # 写入超时（秒），0 默认 30
  idle_timeout: 60    # 空闲超时（秒），0 默认 60
  body_size_limit: 10485760  # 请求体大小限制（字节），0 默认 10MB

db:
  mysql:
    username: root
    password: "123456"
    ip: 127.0.0.1
    port: 3306
    db_name: test
    max_idle_conn: 10          # 最大空闲连接数
    max_open_conn: 100         # 最大打开连接数
    conn_max_lifetime: 3600    # 连接最大生命周期（秒）
    slow_threshold: 200        # 慢查询阈值（毫秒），0 表示不记录

  postgres:
    username: postgres
    password: "123456"
    ip: 127.0.0.1
    port: 5432
    db_name: test
    max_idle_conn: 10
    max_open_conn: 100
    conn_max_lifetime: 3600
    slow_threshold: 200

cache:
  local:
    enable: true
    clear_interval: 3600       # 过期缓存清理间隔（秒）

  redis:
    addr: "127.0.0.1:6379"     # Redis 地址
    password: ""               # Redis 密码
    db: 0                      # Redis DB 编号
    pool_size: 100             # 连接池大小

log:
  local:
    enable: true
    root_dir: runtime/logs     # 日志根目录
    max_size: 100              # 单个日志文件最大大小（MB）
    max_backups: 30            # 最大保留文件数
    max_age: 7                 # 日志保留天数
    compress: true             # 是否压缩旧日志

  sls:                         # 阿里云日志服务（可选）
    enable: false
    endpoint: ""
    access_key: ""
    secret_key: ""
    secret_token: ""
    project: ""
    log_store: ""

trace:
  service_name: Hecc-Blot      # 服务名称
  endpoint: 127.0.0.1:4318     # OTLP HTTP 端点
  sampler:
    type: always               # always / never / probability
    ratio: 0.5                 # probability 模式的采样比例 (0-1)
```

## 配置项详解

### server

| 配置项 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `port` | string | 是 | 服务监听端口 |
| `env` | string | 是 | 运行环境，映射到 Gin 模式 |
| `name` | string | 否 | 服务名称 |
| `enable_trace` | bool | 否 | 是否开启 HTTP 链路追踪中间件 |
| `read_timeout` | int | 否 | 读取请求超时（秒），0 默认 30 |
| `write_timeout` | int | 否 | 写入响应超时（秒），0 默认 30 |
| `idle_timeout` | int | 否 | 空闲连接超时（秒），0 默认 60 |
| `body_size_limit` | int | 否 | 请求体最大字节数，0 默认 10MB |

env 映射关系：

| 配置值 | Gin 模式 | 说明 |
|--------|----------|------|
| `dev` | DebugMode | 输出详细调试日志 |
| `test` | TestMode | 测试模式 |
| `product` | ReleaseMode | 生产模式，关闭调试输出 |

### db.mysql / db.postgres

数据库配置按 IP 是否为空判断是否启用——IP 为空字符串则跳过该数据库的初始化。

| 配置项 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `username` | string | 是 | 数据库用户名 |
| `password` | string | 是 | 数据库密码 |
| `ip` | string | 是 | 数据库 IP，空则跳过 |
| `port` | int | 是 | 数据库端口 |
| `db_name` | string | 是 | 数据库名 |
| `max_idle_conn` | int | 否 | 最大空闲连接数 |
| `max_open_conn` | int | 否 | 最大打开连接数 |
| `conn_max_lifetime` | int | 否 | 连接最大生命周期（秒） |
| `slow_threshold` | int | 否 | 慢查询阈值（毫秒） |

### cache.local

| 配置项 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `enable` | bool | 否 | 是否启用 |
| `clear_interval` | int | 否 | 过期缓存清理间隔（秒），≤0 不启动清理 |

### cache.redis

| 配置项 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `addr` | string | 是 | Redis 地址，如 `127.0.0.1:6379` |
| `password` | string | 否 | Redis 密码 |
| `db` | int | 否 | Redis DB 编号 |
| `pool_size` | int | 否 | 连接池大小 |

### log.local

| 配置项 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `enable` | bool | 否 | 是否启用 |
| `root_dir` | string | 是 | 日志文件根目录，不存在则自动创建 |
| `max_size` | int | 否 | 单个日志文件最大大小（MB） |
| `max_backups` | int | 否 | 最大保留旧文件数 |
| `max_age` | int | 否 | 日志保留天数 |
| `compress` | bool | 否 | 是否压缩旧日志文件 |

日志按级别分文件输出：

| 文件 | 级别 |
|------|------|
| `debug.log` | Debug |
| `info.log` | Info |
| `warn.log` | Warn |
| `error.log` | Error |
| `panic.log` | Panic |

### log.sls

阿里云日志服务（可选），enable 为 false 则不初始化。

| 配置项 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `enable` | bool | 否 | 是否启用 |
| `endpoint` | string | 是 | SLS 端点地址 |
| `access_key` | string | 是 | 阿里云 AccessKey |
| `secret_key` | string | 是 | 阿里云 SecretKey |
| `secret_token` | string | 否 | 阿里云 STS Token |
| `project` | string | 是 | SLS Project 名称 |
| `log_store` | string | 是 | SLS LogStore 名称 |

### trace

| 配置项 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `service_name` | string | 是 | 在追踪系统中的服务名称 |
| `endpoint` | string | 是 | OTLP HTTP 端点地址 |
| `sampler.type` | string | 否 | 采样类型 |
| `sampler.ratio` | float | 否 | 采样比例（probability 模式） |

采样类型：

| 值 | 说明 |
|------|------|
| `always` | 全量采样 |
| `never` | 不采样 |
| `probability` | 按比例采样 |

## 相关文档

| 文档 | 说明 |
|------|------|
| [快速开始](quick_start.md) | 从零搭建项目 |
| [数据库组件](database.md) | 使用 db 配置 |
| [缓存组件](cache.md) | 使用 cache 配置 |
| [日志组件](logging.md) | 使用 log 配置 |
| [链路追踪](trace.md) | 使用 trace 配置 |
