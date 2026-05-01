## ADDED Requirements

### Requirement: 维度与阈值

网关 MUST 支持为 **`POST /v1/chat/completions`**（及 `design.md` 声明的其他受保护路径）配置 **每秒请求数（RPS）** 阈值，且 MUST 至少支持按 **客户端 IP**（`RemoteAddr` 或 `X-Forwarded-For` 首跳策略见 `design.md`）与 **API Key**（通过 **`api-key-management`** 解析出的稳定 **key_id** 或哈希）分别计数。两维度同时配置时，**拒绝条件** MUST 在 `design.md` 固定为 **任一超限即拒绝** 或 **取更严**，并文档化。

#### Scenario: IP 超限返回 429

- **WHEN** 某 IP 在滑动一秒窗口内请求次数超过 **`rps_per_ip`**
- **THEN** 网关 MUST 返回 **429** 与 **`gateway-backend`** 统一 JSON 错误（含 **`request_id`**），且 MUST **不**向上游发起该请求

#### Scenario: API Key 超限返回 429

- **WHEN** 某 API Key 超过 **`rps_per_key`**
- **THEN** 响应为 **429** 且 MUST **不**调用上游

### Requirement: 与鉴权及代理链顺序

**`rps_per_ip`** MUST 在 **鉴权之前** 执行（以便匿名刷接口仍被限制）；**`rps_per_key`** MUST 在 **鉴权成功之后**、发起上游请求之前执行。限流组件整体 MUST 位于 **反向代理之前**。若 **`design.md`** 声明 **OPTIONS** 预检豁免，则 **OPTIONS** 对 **per-key** 计数 MUST **不**递增或适用更低 **`rps_options`**（二选一并在 README 固定）。

#### Scenario: 未鉴权请求不计入 per-key

- **WHEN** 请求无有效 Bearer 且在鉴权阶段被拒绝
- **THEN** **`rps_per_key`** 计数器 MUST **不**对该 key 递增

### Requirement: 可观测性

限流拒绝 MUST 产生 **Zap** 日志，级别 **Warn** 或 **Info**（实现固定），且 MUST 包含 **`request_id`** 与拒绝原因枚举 **`RATE_LIMIT_IP`** / **`RATE_LIMIT_KEY`**。

#### Scenario: 日志不含密钥

- **WHEN** 因 per-key 限流拒绝
- **THEN** Zap 字段 MUST **不**包含 API Key 明文
