---
alwaysApply: true
scene: git_message
---

# Git 提交信息规范

## 概述

本项目采用 **Conventional Commits** 规范，确保提交信息清晰、一致，便于自动化版本管理和 Changelog 生成。所有提交信息均使用中文。

## 提交信息格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

## 各部分说明

### 1. Type（类型）

| 类型 | 说明 |
|-----|------|
| `feat` | 新增功能 (feature) |
| `fix` | 修复 bug |
| `docs` | 文档更新 |
| `style` | 代码格式（不影响代码运行的变动） |
| `refactor` | 重构（既不新增功能，也不是修复 bug） |
| `perf` | 性能优化 |
| `test` | 测试相关 |
| `chore` | 构建过程或辅助工具的变动 |
| `revert` | 撤销之前的提交 |
| `build` | 构建系统或外部依赖的变动 |
| `ci` | CI/CD 配置或脚本的变动 |

### 2. Scope（范围）

可选字段，用于说明提交影响的模块或文件范围，如：

- `log` - 日志模块
- `db` - 数据库模块
- `api` - API 模块
- `cache` - 缓存模块
- `ioc` - IOC 容器模块

### 3. Subject（主题）

- 简短描述，不超过 50 个字符
- 使用祈使句，首字母小写
- 结尾不加句号

### 4. Body（正文）

可选字段，详细描述提交的内容：

- 可以有多行
- 使用空行与 subject 分隔
- 说明修改的原因和具体变更

### 5. Footer（脚注）

可选字段，用于关联 issue 或 breaking change：

- `BREAKING CHANGE`: 破坏性变更说明
- `Fixes #issueNumber`: 关联修复的 issue
- `Refs #issueNumber`: 关联相关 issue

## 示例

### 新增功能

```
feat(log): 添加 SLS 日志支持

- 新增 slsSvc.go 实现阿里云日志服务
- 支持日志级别配置
- 添加单元测试

Refs #123
```

### 修复 Bug

```
fix(db): 修复 MySQL 连接泄漏问题

- 使用 defer 确保连接正确关闭
- 添加连接池配置参数校验

Fixes #456
```

### 文档更新

```
docs(readme): 更新快速开始指南

- 添加配置文件示例
- 补充 API 定义说明
```

### 代码重构

```
refactor(ioc): 优化注入逻辑

- 优化反射性能
- 支持匿名结构体嵌套注入
- 修复指针类型注入问题
```

### 性能优化

```
perf(cache): 优化本地缓存命中率

- 使用 LRU 算法替换简单缓存
- 添加缓存预热机制
```

### 破坏性变更

```
feat(api): 重构响应格式

BREAKING CHANGE: 统一响应格式从 {code, msg, data} 改为 {code, message, data}

- 更新 ResponseSvc 实现
- 修改响应码枚举定义
- 更新相关测试用例
```

## 提交信息检查清单

- [ ] 使用了正确的 type
- [ ] subject 不超过 50 字符
- [ ] subject 使用祈使句、首字母小写
- [ ] body 清晰描述了变更原因和内容
- [ ] 关联了相关 issue（如有）
- [ ] 没有无关的变更

## 工具支持

推荐使用以下工具辅助规范提交：

1. **commitlint**: 提交信息校验工具
2. **cz-cli**: 交互式提交信息生成工具
3. **standard-version**: 自动生成 Changelog 和版本号

## 参考资料

- [Conventional Commits 官方规范](https://www.conventionalcommits.org/)
- [Angular 提交规范](https://github.com/angular/angular/blob/main/CONTRIBUTING.md#commit)
