## MODIFIED Requirements

### Requirement: Zap 日志集成

服务启动路径 MUST 初始化 **zap** 日志（生产或开发合理配置），后续请求处理中的错误 MUST 具备统一记录到 Zap 的路径。若启用 **`proxy-access-logging`** 所定义的访问日志文件输出，网关 MAY 使用 **独立 zapcore** 或等价机制，且 MUST 确保 **访问日志字段定义** 以 **`proxy-access-logging`** 为准；应用主日志与访问日志 MUST **不**重复写入完整 **`Authorization`** 或 API Key 明文。

#### Scenario: 启动日志

- **WHEN** API 进程启动
- **THEN** Zap 向标准输出或配置目标写入至少一条表明服务已启动的日志

#### Scenario: 访问日志与 Zap 错误协同

- **WHEN** 上游连接失败且访问日志处于 **`info`** 级别
- **THEN** 访问日志行 MUST 记录该次失败事务，且 Zap **Error/Warn** MUST 含 **`request_id`** 与错误类型枚举

## ADDED Requirements

### Requirement: 与限流及访问日志的错误协同

当 **`http-rate-limiting`** 返回 **429** 或 **`proxy-access-logging`** 写入失败（磁盘满等）时，网关 MUST 仍返回 **`gateway-backend`** 统一 JSON 错误（若适用）或按 **`design.md`** 降级；此类路径 MUST 写 Zap **Error** 且含 **`request_id`**。

#### Scenario: 429 使用统一错误体

- **WHEN** per-IP 或 per-Key 限流触发
- **THEN** 响应 JSON MUST 包含 **`code`**、**`message`**、**`request_id`**，且 **`code`** MUST 为稳定枚举（如 **`RATE_LIMITED`**）
