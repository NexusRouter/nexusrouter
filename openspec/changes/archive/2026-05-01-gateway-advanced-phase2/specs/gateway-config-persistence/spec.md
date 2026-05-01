## ADDED Requirements

### Requirement: 单一配置文件入口

网关 MUST 支持从 **`NEXUSROUTER_GATEWAY_CONFIG_FILE`**（或 `design.md` 最终键名）指向的 **本地文件**加载 **聚合配置**，其 MUST 至少覆盖：`upstream` 列表与路由字段、`proxy_access_log`、`rate_limit`、`cors` 中与各子规范对应的块。文件格式 MUST 为 **`design.md`** 固定的一种（YAML 或 JSON）。

#### Scenario: 文件不存在时回退

- **WHEN** 环境变量未设置路径或文件不存在
- **THEN** 网关 MUST 以仅环境变量/遗留行为启动（与变更前兼容），且 README MUST 说明该回退

#### Scenario: 文件语法错误

- **WHEN** 启动时文件存在但解析失败
- **THEN** MUST 启动失败 **或** 拒绝加载并记录致命日志（与 `design.md` 一致）

### Requirement: 热更新

当文件内容变更时，网关 MUST 支持 **`SIGHUP`** 触发的 **原子重载**（Unix）；MAY 支持 **`POST /internal/reload-config`**（管理 Bearer）。重载失败 MUST **保留**先前有效配置快照，且 MUST 写 **Zap Error**。

#### Scenario: 热加载后限流阈值变更

- **WHEN** 运维将 **`rps_per_ip`** 从 **100** 改为 **10** 并重载
- **THEN** 新请求 MUST 按 **10** 评判（允许实现上最多一秒的过渡窗口，若有则上限 **1s** 且文档化）

### Requirement: 与环境变量的优先级

若同时存在 **文件配置项** 与 **环境变量**，网关 MUST 采用 **`design.md` 与 README** 声明的优先级（推荐：**文件为基线，显式 `env_override` 字段为真时 env 覆盖**）。

#### Scenario: 文档可验证

- **WHEN** 审查者阅读 `services/gateway/README.md`
- **THEN** 能找到优先级决策表或决策树
