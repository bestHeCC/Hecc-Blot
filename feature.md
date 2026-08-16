# 框架优化规划

各模块待优化项与生产就绪补充项。

---

## 待办

### SSE 背压控制（暂缓）

**位置:** `modules/sse/util/sse_writer.go` — 新增 `WriteSSEDrop`

**问题:** 生产速度 > 消费速度时，TCP 发送缓冲区填满导致 `Write()` 阻塞或 OOM。

**方案:** 提供非阻塞写入（channel + goroutine 异步模型），写入失败丢弃当前帧，业务层自行决定丢弃或降速策略。

---

## 框架优化点与风险

| # | 位置 | 问题 | 建议 |
|---|------|------|------|
| 1 | `modules/ioc` | `Container.values` 无锁，运行时并发 `Set` 有数据竞争 | ✅ 已约定：`docs/ioc_injection.md`「并发约定」+ `ioc_svc.go` 注释明确 Set 仅初始化阶段调用 |
| 2 | `modules/api` | `registerAPI` 每请求 `reflect.New` + `Inject` 反射开销 | 实例隔离要求每请求新实例，反射不可避免；可缓存类型元数据 |
| 3 | `modules/sse` | `sseWriter` 用 mutex 串行化心跳与业务写入，高频推送有锁竞争 | 4.1 异步模型可缓解 |
| 4 | `modules/log` | `sls_svc` 中 `_ = client.PutLogs(...)` 忽略错误，上传失败静默丢失 | 暂缓（决定不管） |
| 5 | `modules/api` | API 层无请求频率限流（防刷） | 可参考 SSE 信号量方案加限流中间件 |
| 6 | 各模块 | `sse/core/api/error/trace` 无测试，刚做的心跳/限流/优雅关闭无保护 | 补单测（`httptest` + 假 Flusher 即可） |
| 7 | `docs/` | 文档全部中文，英文用户仅有 README_EN.md | 后续翻译 docs |
| 8 | `example/config.yaml` | SLS 密钥明文（用户已知，自行处理） | 轮换 + 环境变量注入 |

---

## 模块发布规划

后续拆独立仓库时，按「核心 vs 扩展」划分：

### 默认提供（框架核心，随框架引入）

| 模块 | 职责 | 理由 |
|------|------|------|
| `core` | 契约 + 通用工具 | 所有模块与业务方都依赖的接口层 |
| `ioc` | 依赖注入容器 | 框架核心机制，零依赖 |
| `api` | HTTP 内核 + 响应包装 | Web 能力，业务必经 |
| `error` | 统一错误 | 与 api/业务紧密耦合的基础 |

### 拆独立仓库（可选，按需 `go get`）

| 模块 | 职责 | 不引用的场景 |
|------|------|-------------|
| `db` | 数据库（GORM） | 不用数据库的项目 |
| `cache` | 缓存（本地 + Redis） | 不用缓存的项目 |
| `log` | 日志（Zap + SLS） | 只用标准日志的项目 |
| `trace` | 链路追踪（OpenTelemetry） | 不需要追踪的项目 |
| `sse` | SSE 推送 | 不需要实时推送的项目 |

**依赖关系**：扩展模块只依赖 `core` 的接口（`contract/db`、`contract/log`、`contract/trace` 等），拆出后通过 `go get hecc-core` 引入；`api`/`sse` 额外依赖 `ioc` 接口。核心模块（core/ioc/api/error）互依，作为一个整体默认提供。
