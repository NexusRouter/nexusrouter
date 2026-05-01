# gateway-backend Specification

## Purpose
TBD - created by archiving change project-initialization. Update Purpose after archive.
## Requirements
### Requirement: Go 与 Web 框架版本

`services/gateway` MUST 使用 **Go 1.24.x** 工具链（`go` directive），并依赖 **github.com/gin-gonic/gin v1.10.0** 作为 HTTP 框架。

#### Scenario: 模块解析

- **WHEN** 在模块根执行 `go list -m github.com/gin-gonic/gin`
- **THEN** 输出主版本为 **v1.10.0**（允许补丁后缀以模块代理为准）

### Requirement: 后端依赖矩阵

模块 MUST 声明并可通过构建解析以下依赖（主模块或间接，以 `go.mod` 为准，版本约束与规范一致）：**gorm.io/gorm v1.25.x**、**gorm.io/driver/postgres v1.5.x**、**github.com/redis/go-redis/v9 v9.x**、**github.com/google/wire v0.6.x**、**github.com/spf13/viper v1.19.x**、**go.uber.org/zap v1.27.x**、**github.com/golang-jwt/jwt/v5 v5.x**、**github.com/go-playground/validator/v10 v10.x**、**github.com/golang-migrate/migrate/v4 v4.x**、**github.com/swaggo/swag v1.16.x**、**github.com/stretchr/testify v1.10.x**。

#### Scenario: 编译通过

- **WHEN** 在 `services/gateway` 执行 `go build ./...`
- **THEN** 以零退出码完成

### Requirement: 应用入口与监听端口

HTTP 服务 MUST 自 **`cmd/api`**（或等价的 `cmd/api/main.go`）启动，并监听 **8080** 端口（可配置但默认值 MUST 为 8080）。

#### Scenario: 本地启动

- **WHEN** 开发者运行 API 入口程序（如 `go run ./cmd/api`）
- **THEN** 进程在 **8080** 上接受连接（与现有健康检查路径可并存或迁移）

### Requirement: 分层目录结构

`services/gateway` MUST 具备以下目录：**`cmd/api/`**、**`internal/config/`**、**`internal/router/`**、**`internal/handler/`**、**`internal/service/`**、**`internal/repository/`**。

#### Scenario: 目录存在性检查

- **WHEN** 审查者列出上述路径
- **THEN** 各目录均存在且包含至少占位文件（如 `.gitkeep` 或包注释）以避免空包被工具链忽略

### Requirement: Zap 日志集成

服务启动路径 MUST 初始化 **zap** 日志（生产或开发合理配置），后续请求处理中的错误 MUST 具备统一记录到 Zap 的路径（骨架级即可）。

#### Scenario: 启动日志

- **WHEN** API 进程启动
- **THEN** Zap 向标准输出或配置目标写入至少一条表明服务已启动的日志

### Requirement: Wire 依赖注入

项目 MUST 采用 **google/wire** 组织依赖注入，并提供可生成的 `wire.go` 与生成物（如 `wire_gen.go`）或等价文档化生成步骤。

#### Scenario: 生成注入代码

- **WHEN** 开发者执行 `wire ./...`（在指定包路径）
- **THEN** 生成成功且无待提交的手改冲突（或 CI 文档化步骤通过）

### Requirement: 本地热重载工具

文档或脚本 MUST 说明使用 **github.com/air-verse/air** 进行本地热重载的安装方式（如 `go install github.com/air-verse/air@latest`）及配置文件位置。

#### Scenario: 文档可发现

- **WHEN** 新开发者阅读 `services/gateway` 下 README 或 `openspec/project.md` 链接
- **THEN** 能找到 air 的启动命令说明

### Requirement: 错误响应与日志约定

网关在处理 **自身生成的错误响应**（不包含上游透传的原始错误体）时 MUST 使用 **JSON**，且 body MUST 为对象并 MUST 包含：**`code`**（字符串，机器可读、稳定枚举）、**`message`**（字符串，人类可读）、**`request_id`**（字符串，与响应头 **`X-Request-ID`** 一致；若客户端请求已携带 **`X-Request-ID`**，MUST 使用该值）。错误路径 MUST 记录到 **Zap**，且日志 MUST 包含 **`request_id`** 与 **`code`**（或等价结构化字段），且 MUST NOT 记录完整 API Key 或 **`Authorization`** 头原文。

#### Scenario: 未知路由携带请求 ID

- **WHEN** 客户端请求未注册路由且请求头包含 **`X-Request-ID: probe-1`**
- **THEN** 响应为 JSON，其 **`request_id`** 为 **`probe-1`**，响应头 **`X-Request-ID`** 为 **`probe-1`**，且 Zap 记录该 **404**（或等价）事件并带相同 **`request_id`**

#### Scenario: 未带请求 ID 时服务端生成

- **WHEN** 客户端请求触发网关错误且未携带 **`X-Request-ID`**
- **THEN** 响应头 **`X-Request-ID`** 非空，JSON body 中 **`request_id`** 与该头一致，且 Zap 含相同 **`request_id`**

#### Scenario: Panic 恢复响应含请求 ID

- **WHEN** 处理链触发 panic 并由恢复中间件转换为 **500** JSON
- **THEN** 该 JSON MUST 包含 **`request_id`** 且与 **`X-Request-ID`** 响应头一致

### Requirement: 请求 ID 中间件全局启用

HTTP 引擎 MUST 在业务路由之前注册 **`RequestID`**（或等价命名）中间件：若请求已带 **`X-Request-ID`**，MUST 透传该值；否则 MUST 生成并写入响应头 **`X-Request-ID`**；且 MUST 将当前请求 ID 存入 Gin 上下文供后续处理器与错误封装读取。

#### Scenario: 下游处理器可读 request_id

- **WHEN** 任意已注册业务处理器执行
- **THEN** 其可通过上下文读取与 **`X-Request-ID`** 一致的请求 ID 字符串

### Requirement: 文档与提交语言

后端源码中的 **注释** MUST 使用 **中文**；导出标识符与包名 MUST 使用 **英文**；Git commit message MUST 使用 **中文**。

#### Scenario: 代码审查抽样

- **WHEN** 审查者打开 `cmd/api` 与 `internal` 下主要文件
- **THEN** 注释为中文且 API 命名为英文

