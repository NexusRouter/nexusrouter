# admin-cors-management Specification

## Purpose
TBD - created by archiving change advanced-admin-management. Update Purpose after archive.
## Requirements
### Requirement: CORS 可视化与批量域名

管理控制台 MUST 允许查看与编辑 **`allow_origins`**、**`allow_methods`**、**`allow_headers`**（与 `gateway.yaml` 中 `cors` 段字段对齐）。MUST 支持通过**多行或逗号分隔文本**一次**批量添加**域名（去重、去空白）；提交后 MUST 触发运行时更新使 CORS 中间件使用新配置。

#### Scenario: 批量导入后列表无重复

- **WHEN** 管理员粘贴含重复与空行的域名列表并保存
- **THEN** 持久化后的 **`allow_origins`** MUST 无重复项且无空项

#### Scenario: CORS 关闭时不发送允许头逻辑不变

- **WHEN** `cors.enabled` 为 false
- **THEN** 网关行为与变更前一致（不注入 CORS 允许逻辑）

### Requirement: 实时生效与校验

保存 CORS 配置 MUST 经服务端校验（如方法名大写、头名合法性）；失败 MUST 不覆盖旧配置。成功 MUST 对后续预检与实际请求生效。

#### Scenario: 非法方法名被拒绝

- **WHEN** `allow_methods` 包含非标准 HTTP 方法 token
- **THEN** MUST 拒绝保存并提示错误列表

