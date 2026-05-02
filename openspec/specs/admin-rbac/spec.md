# admin-rbac Specification

## Purpose
TBD - created by archiving change admin-auxiliary-i18n-rbac-alerts. Update Purpose after archive.
## Requirements
### Requirement: 角色枚举与 JWT 声明

管理登录签发的访问令牌 MUST 在可验证载荷中包含 **`role`** 字段，取值为 **`admin`** 或 **`operator`**。系统 MUST 至少支持上述两种角色；`admin` MUST 具备全部管理 API 与仪表盘写能力，`operator` MUST 为只读子集（见下条）。

#### Scenario: 管理员令牌含 role

- **WHEN** 配置的管理员账号登录成功
- **THEN** 解码后的 JWT（或等价会话载荷）含 `role=admin`

### Requirement: 操作员权限边界

`operator` MUST 允许：**GET** 类仪表盘数据（如指标摘要、网关快照只读、日志查询；只读范围以 `design.md` 为准）。`operator` MUST NOT：**创建/修改/删除** API Key、**写入** 网关 YAML 段（上游、CORS、限流、IP 名单等）、**修改** 系统设置写接口。违反时服务端 MUST 返回 **403**。

#### Scenario: 操作员调用写密钥接口被拒绝

- **WHEN** `role=operator` 的用户请求 `POST /api/admin/v1/keys`
- **THEN** 响应为 **403** 且不得修改密钥文件

#### Scenario: 操作员查看指标成功

- **WHEN** `role=operator` 的用户请求 `GET /api/admin/v1/metrics/summary`
- **THEN** 响应为 **200**

### Requirement: 前端按角色隐藏或禁用

仪表盘 MUST 根据 `role` 隐藏或禁用无权的菜单项与表单提交入口；即使用户篡改前端请求，**服务端校验仍为最终权威**。

#### Scenario: 操作员不展示写上游按钮

- **WHEN** `role=operator` 的用户打开上游页
- **THEN** 不得出现可触发写盘保存的提交控件（或控件为禁用且附说明）

