## ADDED Requirements

### Requirement: 系统设置可读字段

管理 API MUST 提供受 JWT 保护的 **系统设置只读视图**，至少包含：**HTTP 监听端口**（当前进程生效值）、**上游请求超时**、**代理访问日志路径**（若启用）及与运行相关的其它已在 `design.md` 列出的键。响应 MUST 标注每个字段的 **`mutability`** 枚举之一：`hot_reload`（可通过现有热更新路径生效）、`restart_required`（需重启进程）、`read_only`（仅展示）。

#### Scenario: 管理员读取设置

- **WHEN** `role=admin` 的用户请求系统设置 GET 接口
- **THEN** 返回 200 且 body 含上述字段及 mutability 元数据

### Requirement: 系统设置写入与边界

对标记为 `hot_reload` 的字段，管理员 MUST 可通过 PUT/PATCH 更新并触发与现网一致的 **持久化 + Reload**（若适用）。对 `restart_required` 字段，API MAY 接受写入配置文件或模板，但 MUST 在响应中返回 **`restart_required: true`** 且 MUST NOT 声称已在线生效除非实现进程内重绑端口（非目标）。`operator` 角色 MUST NOT 调用写入接口（见 `admin-rbac`）。

#### Scenario: 操作员禁止保存设置

- **WHEN** `role=operator` 的用户调用系统设置写入接口
- **THEN** 返回 **403** 且 body 符合网关 JSON 错误约定

#### Scenario: 热更新类字段保存成功

- **WHEN** `role=admin` 的用户更新某 `hot_reload` 字段且校验通过
- **THEN** 新请求已按新配置生效或返回明确的成功载荷说明 Reload 已完成
