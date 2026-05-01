## ADDED Requirements

### Requirement: 健康检查路径与成功语义

网关 MUST 注册 **GET `/health`**，且该路径 MUST **不经过** 需 API Key 的鉴权中间件（以便负载均衡与合成监控无密钥探活）。

#### Scenario: 成功响应结构

- **WHEN** 客户端对 **`/health`** 发起 **GET**
- **THEN** 响应状态码为 **200**，`Content-Type` 为 **`application/json`**，且 JSON 对象 MUST 至少包含：**`status`**（字符串，进程正常接受流量时为 **`ok`**）、**`version`**（字符串，构建或发布标识；未注入构建信息时允许为约定占位如 **`dev`**）、**`server_time`**（字符串，服务端当前 UTC 时间，MUST 符合 **RFC3339** 或 **RFC3339Nano**）

#### Scenario: 监控可解析

- **WHEN** 外部监控系统轮询 **`GET /health`**
- **THEN** 可在不解析上游 LLM 协议的前提下，仅凭 HTTP 200 与 JSON 字段判断进程存活与时间漂移（由监控侧策略使用 **`server_time`**）

### Requirement: 与健康路由并存的代理路由

**`GET /health`** MUST 与 **`POST /v1/chat/completions`** 及其他已注册路由并存，且不得独占根路径。

#### Scenario: Chat 与健康互不影响

- **WHEN** 客户端交替请求 **`/health`** 与 **`/v1/chat/completions`**
- **THEN** 二者分别命中各自处理器，且 **`/health`** 不因未带 Bearer 而被 **401**
