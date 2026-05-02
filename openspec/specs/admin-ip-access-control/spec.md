# admin-ip-access-control Specification

## Purpose
TBD - created by archiving change advanced-admin-management. Update Purpose after archive.
## Requirements
### Requirement: IP 访问模式与条目

网关 MUST 支持 **`ip_access.mode`** 取值：**`off`**、**`allowlist`**、**`denylist`**；并维护 **`ip_access.cidrs`** 字符串列表（IPv4/IPv6 单地址或 **CIDR** 表示法，以实现校验为准）。**白名单**模式下：未匹配任何 CIDR 的客户端 IP MUST 被拒绝访问至少 **`POST /v1/chat/completions`**（响应为网关统一 JSON 错误，状态码 **403** 或 **403/429** 由实现固定并文档化）。**黑名单**模式下：匹配 CIDR 的 IP MUST 被拒绝（同上）。**off** 模式：不执行名单逻辑。

#### Scenario: 白名单仅允许指定网段

- **WHEN** `mode=allowlist` 且列表仅包含 `203.0.113.0/24`，请求来自 `198.51.100.1`
- **THEN** MUST 拒绝该请求且不调用上游

#### Scenario: 黑名单拦截命中 IP

- **WHEN** `mode=denylist` 且列表包含客户端 IP 所在 CIDR
- **THEN** MUST 拒绝该请求

### Requirement: 批量管理与互斥

管理 API MUST 支持 **批量追加** 与 **批量删除** CIDR 字符串；**MUST NOT** 同时处于 allowlist 与 denylist 有效叠加状态——以 `mode` 单选为准。变更 MUST 实时生效（与 `design.md` 写盘/重载策略一致）。

#### Scenario: 无效 CIDR 被拒绝

- **WHEN** 批量提交中含非法 CIDR 文本
- **THEN** 整批保存失败或返回部分失败列表（实现二选一并文档化），且旧配置保留

