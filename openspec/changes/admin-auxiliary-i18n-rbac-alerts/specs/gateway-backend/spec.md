## ADDED Requirements

### Requirement: 管理端系统设置 API

在 **`/api/admin/v1`** 下 MUST 增加受 JWT 保护的 **系统设置** 端点（路径以实现为准，如 `GET/PUT /api/admin/v1/system/settings`），语义 MUST 满足 `admin-system-settings` 能力规范；写入 MUST 经校验且失败时保留旧配置。

#### Scenario: 未认证访问设置

- **WHEN** 匿名请求系统设置 GET
- **THEN** 返回 **401**

### Requirement: 管理端 RBAC 强制校验

所有 **非只读** 管理写接口（含密钥、网关配置、安全策略、CORS、限流、IP 名单、系统设置写等，以实现维护的清单为准）MUST 在校验 JWT 后检查 **`role`**；当 `role=operator` 时 MUST 返回 **403**。只读 GET 清单 MUST 在 `design.md` 或实现代码注释中维护并与 `admin-rbac` 一致。

#### Scenario: 操作员 PUT 网关配置被拒绝

- **WHEN** `operator` 调用 `PUT /api/admin/v1/gateway/config`
- **THEN** 返回 **403**

### Requirement: 管理端告警状态 API

MUST 提供 `GET /api/admin/v1/alerts/status`（或等价路径）返回 `admin-runtime-alerts` 规范定义的状态与 `reasons`；实现 MUST 基于进程内指标与配置阈值计算，MUST NOT 在响应中泄露完整 API Key 或 `Authorization` 原文。

#### Scenario: 操作员可读告警

- **WHEN** `operator` 调用告警状态 GET
- **THEN** 返回 **200** 且 body 符合规范
