## ADDED Requirements

### Requirement: 管理端对密钥文件的受控写入

在启用管理控制台且配置了 **`NEXUSROUTER_GATEWAY_KEYS_FILE`** 的前提下，网关 MAY 通过**受管理员权限保护**的 HTTP API 对同一 JSON 文件执行新增、更新（禁用/过期）、删除操作；每次成功写入后 MUST 触发与 **`POST /internal/reload-keys`** 等价的 **`KeyStore`** 重载语义，使 **`Bearer API Key 校验`** 要求立即适用于新集合。

#### Scenario: 管理 API 写入后与热加载一致

- **WHEN** 管理员通过 API 新增一条启用密钥并提交成功
- **THEN** 后续 **`POST /v1/chat/completions`** 使用对应 Bearer MUST 通过鉴权（在未过期且未禁用前提下）

#### Scenario: 并发写冲突可检测

- **WHEN** 两次写入基于同一基线版本发生冲突（若实现版本控制）
- **THEN** 后者 MUST 失败并提示刷新，而非静默丢更新
