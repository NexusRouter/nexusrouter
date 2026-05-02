## ADDED Requirements

### Requirement: 告警阈值配置

网关配置（如 `gateway.yaml` 或等价快照）MUST 支持 **`admin_alerts`**（或 `design.md` 最终命名）段，至少包含：**错误率阈值**（百分比或千分比，实现固定一种并在文档说明）、**评估窗口**（秒）、**连续命中周期数**（防抖）。缺省未配置时 MUST 等价于「告警功能关闭」且不报错。

#### Scenario: 未配置时不告警

- **WHEN** 未配置 `admin_alerts` 段
- **THEN** 管理告警 API 返回未启用状态且前端不展示红色告警条

### Requirement: 告警状态 API

管理 API MUST 提供 **GET** 端点返回当前告警状态：`ok` | `warning` | `critical`（或实现固定枚举），并包含 **`reasons`** 字符串数组（无敏感数据）。该端点 MUST 受 JWT 保护；`operator` MAY 读取。

#### Scenario: 超阈值返回 critical

- **WHEN** 过去窗口内网关错误率超过配置阈值且满足连续周期条件
- **THEN** GET 返回 `critical` 且 `reasons` 至少包含一条可读的摘要编码（如 `HIGH_ERROR_RATE`）

### Requirement: 管理面板视觉告警

仪表盘布局 MUST 在告警状态非 `ok` 时于 **顶栏或内容区上方** 展示显著视觉差异（颜色/图标）；`critical` MUST 比 `warning` 更突出。用户 MUST 能区分「数据加载失败」与「运行告警」两种状态。

#### Scenario: critical 时展示告警条

- **WHEN** GET 告警状态为 `critical`
- **THEN** 管理布局出现红色系告警条并展示摘要文案（走 i18n）
