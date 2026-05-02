## ADDED Requirements

### Requirement: 进阶管理 HTTP API 面

在 **`/api/admin/v1`**（或文档化之等价前缀）下 MUST 增加以下能力对应的受 JWT 保护端点：**日志查询与 CSV 导出**、**限流规则 CRUD**、**CORS 配置读写与批量域名**、**IP 名单读写与批量操作**。未认证访问 MUST 返回 **401** 且 body 符合既有网关 JSON 错误约定。

#### Scenario: 导出接口需管理员权限

- **WHEN** 匿名客户端请求日志 CSV 导出 URL
- **THEN** MUST **不**返回文件内容且为 **401**

### Requirement: 请求链整合 IP 名单

在 **IP 限流** 与 **鉴权** 之间的精确顺序在 `design.md` 固定；IP 名单拒绝 MUST 在调用上游前发生，且 MUST 记录结构化日志（含 **`request_id`**），且 MUST NOT 记录完整 API Key。

#### Scenario: 名单拒绝不产生上游流量

- **WHEN** 请求被黑名单或白名单缺省拒绝
- **THEN** MUST **不**对上游发起 **`POST /v1/chat/completions`**
