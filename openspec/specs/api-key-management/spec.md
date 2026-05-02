# api-key-management Specification

## Purpose

定义网关对 **`POST /v1/chat/completions`** 等受保护路由的 **API Key** 校验、密钥元数据（启用/禁用、过期）及与文件热加载相关的行为契约。
## Requirements
### Requirement: Bearer API Key 校验

网关 MUST 对受保护路由（至少包含 **`POST /v1/chat/completions`**）在调用上游前校验 **`Authorization`** 头：格式 MUST 为 **`Bearer <API_KEY>`**（`Bearer` 大小写不敏感，`<API_KEY>` 为去首尾空白后的令牌字符串）。校验 MUST 基于当前已加载的密钥集合（见「密钥元数据与存储」），且比对成功仅当：密钥记录存在、**`disabled`** 为 false、且（**`expires_at`** 为空或省略，或当前 UTC 时间 **`now`** 满足 **`now.Before(expires_at)`**——即 **`expires_at`** 为首次失效时刻，达到该时刻起令牌无效）。

#### Scenario: 缺少 Authorization

- **WHEN** 客户端对受保护路由发起请求且未携带 **`Authorization`** 头
- **THEN** 响应状态码为 **401**，body 为网关统一 JSON 错误（含 **`code`**、**`message`**、**`request_id`**），且 MUST **不调用上游**

#### Scenario: 非 Bearer 或格式错误

- **WHEN** **`Authorization`** 存在但不满足 **`Bearer <token>`** 可解析形式
- **THEN** 响应为 **401** 与统一 JSON 错误，且 MUST **不调用上游**

#### Scenario: 令牌不在库或已禁用

- **WHEN** Bearer 令牌与任一启用记录均不一致，或匹配记录的 **`disabled`** 为 true
- **THEN** 响应为 **401** 与统一 JSON 错误，且 MUST **不调用上游**

#### Scenario: 令牌已过期

- **WHEN** Bearer 令牌匹配某记录且 **`expires_at`** 已设置，但当前 UTC 时间 **`now`** 满足 **`!now.Before(expires_at)`**（含恰好在失效时刻）
- **THEN** 响应为 **401** 与统一 JSON 错误，且 MUST **不调用上游**

#### Scenario: 校验通过

- **WHEN** Bearer 令牌匹配启用且未过期的记录
- **THEN** 请求进入后续处理链（如反向代理），且 MUST **不**因鉴权再次返回 **401**

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

### Requirement: 与 OpenAI 兼容代理的衔接

**`POST /v1/chat/completions`** MUST 在 **`openai-chat-completions-proxy`** 所述转发逻辑之前链接本规范所述鉴权中间件顺序：**Request ID → 鉴权 → 代理**（中间件精确顺序以实现为准，MUST 在 `design.md` 固定）。

#### Scenario: 鉴权失败不产生上游流量

- **WHEN** 鉴权失败
- **THEN** 上游不得收到对应 **`POST /v1/chat/completions`** 请求（无 TCP 连接或无任何 HTTP 请求发出，以可观测层能判定为准）

### Requirement: 管理端对密钥文件的受控写入

在启用管理控制台且配置了 **`NEXUSROUTER_GATEWAY_KEYS_FILE`** 的前提下，网关 MAY 通过**受管理员权限保护**的 HTTP API 对同一 JSON 文件执行新增、更新（禁用/过期）、删除操作；每次成功写入后 MUST 触发与 **`POST /internal/reload-keys`** 等价的 **`KeyStore`** 重载语义，使 **`Bearer API Key 校验`** 要求立即适用于新集合。

#### Scenario: 管理 API 写入后与热加载一致

- **WHEN** 管理员通过 API 新增一条启用密钥并提交成功
- **THEN** 后续 **`POST /v1/chat/completions`** 使用对应 Bearer MUST 通过鉴权（在未过期且未禁用前提下）

#### Scenario: 并发写冲突可检测

- **WHEN** 两次写入基于同一基线版本发生冲突（若实现版本控制）
- **THEN** 后者 MUST 失败并提示刷新，而非静默丢更新

