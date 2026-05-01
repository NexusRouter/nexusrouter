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

网关 MUST 支持从 **`NEXUSROUTER_GATEWAY_KEYS_FILE`**（或实现文档中等价配置键）指向的 **JSON 文件**加载密钥记录。每条记录 MUST 至少包含：**`id`**（字符串，唯一）、**`secret`**（字符串，与 Bearer 令牌比对）、**`disabled`**（布尔）、**`expires_at`**（可空，RFC3339/RFC3339Nano 字符串；空表示不过期）。进程 MUST 在启动时加载；MUST 支持 **`SIGHUP`** 触发热加载（Unix）；MAY 支持受令牌保护的 **`POST /internal/reload-keys`** 作为无信号环境的等价机制（若实现，MUST 文档化认证方式）。

#### Scenario: 文件缺失或 JSON 无效

- **WHEN** 启动时路径已配置但文件不可读或 JSON 无法解析为对象数组
- **THEN** MUST 启动失败 **或**（若实现固定为降级策略）拒绝所有受保护路由并记录 Zap 致命/错误日志；实现 MUST 在 `design.md` 与 README 中二选一并一致

#### Scenario: 热加载后新密钥生效

- **WHEN** 运维更新 JSON 文件并触发 **`SIGHUP`**（或调用重载接口）
- **THEN** 后续请求使用新密钥集合进行校验，且进行中请求不得崩溃进程

### Requirement: 与 OpenAI 兼容代理的衔接

**`POST /v1/chat/completions`** MUST 在 **`openai-chat-completions-proxy`** 所述转发逻辑之前链接本规范所述鉴权中间件顺序：**Request ID → 鉴权 → 代理**（中间件精确顺序以实现为准，MUST 在 `design.md` 固定）。

#### Scenario: 鉴权失败不产生上游流量

- **WHEN** 鉴权失败
- **THEN** 上游不得收到对应 **`POST /v1/chat/completions`** 请求（无 TCP 连接或无任何 HTTP 请求发出，以可观测层能判定为准）
