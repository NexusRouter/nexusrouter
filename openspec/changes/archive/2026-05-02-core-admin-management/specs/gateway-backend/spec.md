## ADDED Requirements

### Requirement: 管理端 HTTP API 与鉴权

`services/gateway` MUST 暴露一组**管理用途**的 HTTP API（路径前缀在实现中固定，如 **`/admin` 或 `/api/admin`**），用于支撑控制台的登录、指标查询、上游配置与 API Key 管理。除登录/重置密码等公开端点外，其余管理端点 MUST 要求**有效管理员权限令牌**（或等价会话），未携带或无效时 MUST 返回 **401** 且 body 符合 **`gateway-backend`** 既有 JSON 错误约定（含 **`code`**、**`message`**、**`request_id`**）。

#### Scenario: 无令牌访问受保护管理 API

- **WHEN** 客户端调用需认证的管理 API 且不携带有效凭证
- **THEN** 响应为 **401**，且 body 为统一 JSON 错误格式

### Requirement: 网关运行指标对外查询

网关 MUST 提供管理端可消费的**指标查询接口**（或单一聚合端点），使 **`admin-dashboard-metrics`** 所需之在线状态、请求量、成功率、平均耗时、今日/昨日对比与错误 **`code`** 聚合可被前端拉取或订阅。指标覆盖范围 MUST 在 `design.md` 声明；若某指标暂不可用 MUST 明确返回占位或省略策略而非伪造数据。

#### Scenario: 指标端点需管理员权限

- **WHEN** 未认证客户端请求指标端点
- **THEN** MUST **不**返回可识别业务的详细统计（**401**）

### Requirement: 上游与密钥管理 API 与运行时一致

管理 API 对 **`gateway.yaml`** 与 **API Key JSON** 的写入 MUST 与 `design.md` 一致：写后 MUST 使 **`KeyStore`** / **`runtime.Store`** 进入与磁盘一致的有效状态；失败时 MUST 保留先前快照并返回错误。

#### Scenario: 写文件失败不损坏旧快照

- **WHEN** 持久化配置写入失败
- **THEN** 运行时继续服务先前有效配置，且响应指示失败原因
