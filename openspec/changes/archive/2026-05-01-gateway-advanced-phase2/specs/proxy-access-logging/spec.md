## ADDED Requirements

### Requirement: 代理访问日志字段

当 **`proxy_access_log`**（或 `gateway.yaml` 中等价键，见 `design.md`）启用时，网关 MUST 对每次完成的 **`POST /v1/chat/completions`** 代理事务写一条 **结构化日志**（JSON 行或 `zap` 结构化字段集合），且 MUST 至少包含：**`ts`**（RFC3339Nano UTC）、**`request_id`**、**`method`**、**`path`**、**`client_ip`**、**`upstream_id`**（或等价匿名标识）、**`upstream_host`**、**`status`**（回写客户端的 HTTP 状态码）、**`duration_ms`**（整数；起止时刻定义见 **`design.md`**，须覆盖非流式与 SSE）。

#### Scenario: 成功与上游错误均记录

- **WHEN** 上游返回 **200** 或 **4xx/5xx** 且 body 原样透传
- **THEN** 访问日志中 **`status`** 与客户端所见一致，且 **`duration_ms`** 非负

#### Scenario: 网关自身错误不写完整敏感头

- **WHEN** 请求在到达上游前失败（如 **401**、**429**）
- **THEN** 访问日志 MAY 省略 **`upstream_host`** 或置空，且 MUST NOT 记录 **`Authorization`** 明文或完整 API Key

### Requirement: 日志分级与输出

网关 MUST 支持配置 **`access_log_level`** 为 **`info`** 与 **`error`**（或 `design.md` 超集）：**`info`** MUST 记录所有完成的事务（含成功与上游业务错误码）；**`error`** MUST 仅记录网关内部错误、连接失败、超时及 **5xx** 透传（实现细则见 `design.md`）。输出 MUST 支持 **文件路径** 与/或 **stdout**，二者可同时启用或择一，MUST 在 README 说明。

#### Scenario: error 级别不记录纯 200

- **WHEN** **`access_log_level`** 为 **`error`** 且单次代理返回 **200**
- **THEN** 该请求 MUST **不**写入访问日志文件（或等价「无新增行」）

### Requirement: 请求头日志脱敏

若配置记录请求头，网关 MUST 仅记录 **`design.md`** 中 **允许名单** 内的头名，且对 **`Authorization`**、**`Cookie`**、**`X-API-Key`** 等敏感头 MUST **永不记录值**（可记录 **`present: true`** 布尔）。

#### Scenario: 带 Authorization 的成功请求

- **WHEN** 客户端携带 **`Authorization`** 且 **`info`** 级别开启头日志
- **THEN** 日志中 MUST **不**出现 Bearer 令牌子串
