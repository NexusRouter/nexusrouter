# gateway-data-persistence Specification

## Purpose
TBD - created by archiving change gorm-sqlite-default-persistence. Update Purpose after archive.
## Requirements
### Requirement: 默认 SQLite 文件型持久化

当 **未** 配置实现文档所定义的 **数据库 DSN**（或等价主连接串）时，网关 MUST 使用 **GORM** 打开 **SQLite** 持久化，且默认数据库文件名为 **`gateway.db`**（完整路径 MAY 通过环境变量覆盖，MUST 在 README 中说明默认值相对进程工作目录或专用数据目录的规则）。

#### Scenario: 无 DSN 启动创建或打开库文件

- **WHEN** 进程启动且 DSN 为空或未设置
- **THEN** 网关成功打开或创建 SQLite 文件、完成 `AutoMigrate`（见下条），且不因缺少独立数据库服务而失败

### Requirement: Postgres 与 SQLite 双模式切换

当配置了非空 **Postgres** DSN 时，网关 MUST 使用同一套 GORM 模型与数据访问路径连接 Postgres，且 MUST NOT 要求修改业务包内调用方式即可替换 SQLite 模式（仅配置变更）。

#### Scenario: 设置 DSN 后使用 Postgres

- **WHEN** 启动时 DSN 指向有效 Postgres 且凭据与网络可达
- **THEN** 网关使用 Postgres Dialector 打开连接、执行 `AutoMigrate`，且持久化读写与 SQLite 模式行为一致（除方言性能差异外）

### Requirement: 启动时 AutoMigrate

网关 MUST 在应用接受业务流量之前，对全部已定义的持久化模型执行 **GORM `AutoMigrate`**，以创建或演进表结构；MUST NOT 依赖运维手写 **CREATE TABLE** 或提交 SQL 迁移文件作为本变更的必备步骤。

#### Scenario: 空库首次启动

- **WHEN** 数据库文件或 Postgres schema 中尚不存在所需表
- **THEN** 启动成功后相应表存在且进程可正常读写

### Requirement: 从文件态导入（升级兼容）

若启动时数据库中 **关键域数据为空**（实现 MUST 定义「空」的判定，如网关配置行数为零），且文件系统中仍存在先前约定的 **`gateway.yaml`** 或 **`NEXUSROUTER_GATEWAY_KEYS_FILE`** 等遗留源，网关 MUST 将该数据 **导入** 数据库并记录结构化日志；导入后 **真源** MUST 为数据库（文件 MAY 仅作备份或不再读取，行为 MUST 与 README 一致）。

#### Scenario: 仅有 YAML 与 keys 文件的旧部署首次升级

- **WHEN** 新二进制首次启动且 DB 为空、YAML/JSON 有效
- **THEN** 导入后管理功能与 API Key 校验与导入前语义等价（在字段可映射范围内），且后续持久化写入 DB

### Requirement: 依赖与实现约束

本能力 MUST 基于 **gorm.io/gorm** 实现；SQLite 模式 MUST 通过 **`gorm.io/driver/sqlite`**（或文档声明的等价官方驱动）注册；MUST NOT 为同一持久化域引入第二套 ORM 或并行抽象层。

#### Scenario: 模块依赖可解析

- **WHEN** 在 `services/gateway` 执行 `go mod tidy` 与 `go build ./...`
- **THEN** GORM 与 SQLite、Postgres 驱动均解析成功

