# Hecc-Blot

[![Go Version](https://img.shields.io/badge/Go-1.26.1-blue)](https://github.com/bestHeCC/Hecc-Blot)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)
[![Gitee Repo](https://img.shields.io/badge/Gitee-hecc--go--core-red)](https://gitee.com/bestHeCC/hecc-go-core)

Hecc-Blot is a lightweight Go backend framework built around interface-oriented design, providing dependency injection, route registration, parameter validation, and unified responses.

## Getting the Code

The project is mirrored on both GitHub and Gitee — clone from either:

```bash
# GitHub
git clone https://github.com/bestHeCC/Hecc-Blot.git

# Gitee
git clone https://gitee.com/bestHeCC/hecc-go-core.git
```

## Features

- **Interface-oriented**: all components interact through interface contracts, easy to replace and extend
- **Dependency injection**: reflection-based IOC container, auto-inject via the `inject` tag
- **Routing**: built on Gin, supports GET/POST route registration and middleware chains
- **Validation**: automatic binding and validation with customizable error messages
- **Unified response**: wraps return values into a `{code, message, data}` format
- **Multi-database**: supports MySQL and PostgreSQL with runtime switching
- **Transactions**: chainable transaction API
- **Two-tier cache**: in-memory cache + Redis, with expiry cleanup
- **Tracing**: OpenTelemetry-based, OTLP export to Jaeger
- **SSE**: Server-Sent Events sharing the API port, for real-time push
- **Replaceable**: every component can be swapped by implementing its interface and registering it with the IOC

## Quick Start

See [`example.go`](example.go) for a complete runnable example covering all features, organized by module.

```bash
go mod tidy
```

## Project Layout

```
├── modules/                # sub-modules (managed by go.work, monorepo multi-module)
│   ├── ioc/                # dependency injection container (github.com/bestHeCC/hecc-ioc)
│   ├── core/               # contract SDK (github.com/bestHeCC/hecc-core)
│   │   ├── contract/       # interface contracts (api/cache/db/error/ioc/log/sse/trace)
│   │   ├── entity/         # entities and config structs
│   │   ├── enum/           # enums (env/db/response/trace)
│   │   └── util/           # utilities (pagination, validation messages, context)
│   ├── api/                # HTTP core (github.com/bestHeCC/hecc-api)
│   ├── error/              # unified errors (github.com/bestHeCC/hecc-error)
│   ├── sse/                # SSE push (github.com/bestHeCC/hecc-sse)
│   ├── db/                 # MySQL / PostgreSQL (github.com/bestHeCC/hecc-db)
│   ├── cache/              # local + Redis cache (github.com/bestHeCC/hecc-cache)
│   ├── log/                # logging (github.com/bestHeCC/hecc-log)
│   └── trace/              # OpenTelemetry tracing (github.com/bestHeCC/hecc-trace)
├── docs/                   # documentation (currently in Chinese)
├── go.work                 # workspace config
├── example.go              # full usage example
└── README.md
```

## Documentation

> **Note**: the docs under `docs/` are currently written in Chinese.

### Example Walkthrough

`example.go` is divided into 11 sections and serves as living documentation:

| # | Section | Demonstrates | Details |
|---|---------|--------------|---------|
| 1 | Entry point | main() skeleton: init → IOC → routes → start | [Quick Start](docs/quick_start.md) |
| 2 | Config loading | viper reads config.yaml | [Config](docs/config.md) |
| 3 | Model definition | IDbModel interface, TableName, multiple models | [Database](docs/database.md) |
| 4 | Request & validation | binding tags, GetMessages() | [Routes & Middleware](docs/routes_middleware.md) |
| 5 | Middleware | Authorization check, inject injection | [Routes & Middleware](docs/routes_middleware.md) |
| 6 | Database CRUD | Add/Take/Find/Save/Remove/Count/transactions | [Database](docs/database.md) |
| 7 | Multi-database | MySQL ↔ PostgreSQL switching | [Database](docs/database.md) |
| 8 | Cache operations | Local/Redis read-write-delete, Hash, read-through | [Cache](docs/cache.md) |
| 9 | Tracing | Span/SetAttribute/RecordError/sub-span | [Tracing](docs/trace.md) |
| 10 | Pagination | offset + cursor pagination | [Pagination](docs/paginator.md) |
| 11 | SSE | ISse interface, heartbeat, Flusher assertion | [SSE](docs/sse.md) |

### Getting Started

| Doc | Description |
|-----|-------------|
| [Quick Start Guide](docs/quick_start.md) | full tutorial for building a project from scratch |
| [Config Reference](docs/config.md) | all config.yaml options |

### Core Mechanisms

| Doc | Description |
|-----|-------------|
| [Routes & Middleware](docs/routes_middleware.md) | route registration, middleware, auto-validation, response wrapping |
| [IOC Injection](docs/ioc_injection.md) | injection principles, rules, named injection |
| [Component Replacement](docs/component_replacement.md) | full examples of swapping log/db/cache components |

### Component Usage

| Doc | Description |
|-----|-------------|
| [Logging](docs/logging.md) | local file logging, Alibaba Cloud SLS |
| [Database](docs/database.md) | CRUD, transactions, multi-database, model definition |
| [Cache](docs/cache.md) | local cache, Redis, expiry cleanup, tracing integration |
| [Tracing](docs/trace.md) | OpenTelemetry integration, span operations, cross-service propagation |
| [SSE](docs/sse.md) | SSE usage, route registration, middleware reuse, error handling |
| [Pagination](docs/paginator.md) | offset/limit and cursor pagination |

## Component Overview

### IOC Container

Auto-inject dependencies via the `inject:""` tag — no manual wiring. → [IOC Injection](docs/ioc_injection.md)

### API Service

Routes automatically perform binding, validation and response wrapping. → [Routes & Middleware](docs/routes_middleware.md)

### Database Service

MySQL and PostgreSQL support, chainable queries, transactions. → [Database](docs/database.md)

### Cache Service

In-memory + Redis two-tier cache with Hash operations and read-through. → [Cache](docs/cache.md)

### Logging Service

Local file logging (Zap + lumberjack rotation) and Alibaba Cloud SLS. → [Logging](docs/logging.md)

### Tracing

OpenTelemetry-based, auto-traces HTTP requests and correlates logs. → [Tracing](docs/trace.md)

### SSE Real-time Push

Shares the API port and pushes from the server via the `ISse` interface. → [SSE](docs/sse.md)

### Pagination

Offset/limit and cursor pagination. → [Pagination](docs/paginator.md)

## Design Principles

1. **Dependency inversion**: high-level modules depend on abstractions, not concrete implementations
2. **Interface segregation**: each interface defines a single responsibility
3. **Open/closed principle**: open for extension, closed for modification

## Roadmap

See [feature.md](feature.md) for the framework's optimization plan.

## Thanks

If Hecc-Blot helps you, a ⭐️ is appreciated.

### Feedback & Contributing

- **Bug reports and feature requests**: open an [Issue](https://gitee.com/bestHeCC/hecc-go-core/issues)
- **Code contributions**: pull requests are welcome

### Credits

- [Gin](https://github.com/gin-gonic/gin) — high-performance Go web framework
- [GORM](https://github.com/go-gorm/gorm) — Go ORM library
- [Zap](https://github.com/uber-go/zap) — high-performance logging library
- [OpenTelemetry](https://opentelemetry.io/) — distributed tracing standard

## License

MIT License
