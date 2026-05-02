## Why

当前网关将路由快照、限流等写入 **YAML**、API Key 写入 **JSON 文件**、管理员凭据依赖 **环境变量**，多路径持久化增加运维心智负担，且不利于原子事务与后续水平扩展。项目已在 `go.mod` 中预留 **GORM / Postgres**，但尚未落地；现在以 **零外部依赖** 的默认路径先把结构化数据统一进数据库，可立即获得开箱即用的持久化，并与远期 **Postgres 生产化** 对齐。

## What Changes

- 引入 **GORM + SQLite 驱动**；当未配置数据库连接串时，默认使用进程可写目录下的 **`gateway.db`**（SQLite 文件），自动 **`AutoMigrate`** 建表，**不**要求手写 SQL 迁移文件或单独部署数据库服务。
- 将网关核心业务数据（路由/上游/CORS/限流/IP 策略等快照、API Key 元数据、管理员账号等，以设计为准）从 **YAML/JSON/纯环境变量文件模型** 迁移为 **GORM 模型读写**；保留通过 **单一 DSN 环境变量** 切换至 **Postgres** 的能力，业务层以接口/仓储抽象数据库方言，**切换连接配置即可**，无需分叉两套业务实现。
- 定义从现有文件持久化到 DB 的 **启动迁移或兼容读取** 策略（见 `design.md`），避免已有部署静默丢数据。
- 文档更新：默认行为、可选 DSN、与 `NEXUSROUTER_GATEWAY_CONFIG_FILE` / `NEXUSROUTER_GATEWAY_KEYS_FILE` 的优先级或弃用路径（若部分废弃则标 **BREAKING** 并给迁移说明）。

## Capabilities

### New Capabilities

- `gateway-data-persistence`: 基于 GORM 的网关结构化持久化；无 DSN 时默认 SQLite 文件 `gateway.db`；有 DSN 时使用 Postgres（或与驱动一致的目标库）；启动 `AutoMigrate`；与单一连接配置切换策略。

### Modified Capabilities

- `gateway-backend`: 补充/调整「持久化与配置来源」相关需求（例如依赖矩阵增加 `gorm.io/driver/sqlite`，以及运行时数据来自 DB 的约定）。
- `api-key-management`: 「密钥元数据与存储」从**仅 JSON 文件**扩展为**以 GORM 持久化为主**（及可选的文件回填/迁移期兼容），并更新热加载与 `SIGHUP` 行为在 DB 模式下的等价语义。

## Impact

- **代码**: `services/gateway`（`internal/config`、`internal/runtime`、`internal/keystore`、`internal/repository`、`cmd/api`、Wire）、可能的管理 API 与 README。
- **依赖**: 新增 `gorm.io/driver/sqlite`（或官方推荐 SQLite 驱动），保留现有 Postgres 驱动。
- **运维**: 默认多一个 `gateway.db` 文件；生产可改 DSN 指向 Postgres；需说明备份与文件权限。
- **规范**: `openspec/specs/gateway-backend`、`openspec/specs/api-key-management` 行为条目的增量更新；新增 `openspec/specs/gateway-data-persistence`。
