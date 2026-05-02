# api-key-management Specification Delta

## MODIFIED Requirements

### Requirement: 密钥元数据与存储

网关 MUST 将 API Key 记录持久化在 **GORM 所管理的数据库**中（默认 SQLite 文件、可选 Postgres，见 **`gateway-data-persistence`** 规范）；每条记录 MUST 至少包含：**`id`**（字符串，唯一）、**`secret`**（字符串，与 Bearer 令牌比对）、**`disabled`**（布尔）、**`expires_at`**（可空，UTC 时刻；空表示不过期）。进程 MUST 在启动时从数据库加载至内存（或等价缓存）以供鉴权；MUST 支持 **`SIGHUP`**（Unix）触发自数据库的重新加载；MAY 支持受令牌保护的 **`POST /internal/reload-keys`**，其在 DB 模式下的语义 MUST 为重新自数据库拉取密钥集合（MUST 文档化）。为升级兼容，当数据库中密钥集合为空且 **`NEXUSROUTER_GATEWAY_KEYS_FILE`** 指向有效 JSON 文件时，网关 MUST 将该文件内容导入数据库并随后以数据库为真源（导入失败策略 MUST 与 `design.md` 及 README 一致）。

#### Scenario: 数据库可用且含有效密钥行

- **WHEN** 启动时数据库已存在至少一条可解析的密钥记录
- **THEN** 鉴权使用数据库中的记录，且受保护路由行为符合 Bearer 校验规范

#### Scenario: 自数据库热加载

- **WHEN** 运维在数据库中更新密钥记录后触发 **`SIGHUP`**（或文档化的重载接口）
- **THEN** 后续请求使用更新后的密钥集合，且进行中请求不得崩溃进程

#### Scenario: 空库且存在遗留 JSON 文件

- **WHEN** 启动时数据库无密钥行且 **`NEXUSROUTER_GATEWAY_KEYS_FILE`** 配置为可读有效 JSON 数组
- **THEN** 密钥被导入数据库且后续鉴权以导入后数据为准

#### Scenario: 数据库不可用

- **WHEN** 启动时无法打开 DSN 或 SQLite 文件或 `AutoMigrate` 失败
- **THEN** MUST 启动失败并记录 Zap 错误；MUST **不**进入半启动且误接受流量之状态
