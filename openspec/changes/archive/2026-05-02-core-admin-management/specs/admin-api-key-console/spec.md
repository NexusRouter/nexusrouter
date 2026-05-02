## ADDED Requirements

### Requirement: API Key 列表与脱敏

管理控制台 MUST 列出当前密钥库中的全部 API Key 记录，展示字段至少包括：**脱敏后的密钥**、**状态（启用/禁用）**、**有效期**、**创建时间**（若底层存储无创建时间，实现 MUST 定义等价字段或文档化为「未知」且不得伪造）。**脱敏**规则 MUST 确保完整 **`secret`** 不可从列表接口恢复。

#### Scenario: 列表从不返回完整 secret

- **WHEN** 客户端调用管理端 Key 列表 API
- **THEN** 响应体 MUST NOT 包含完整明文 **`secret`**

### Requirement: 单条生命周期操作

控制台 MUST 支持 **新增**、**禁用**、**删除** 单条 API Key；操作结果 MUST 反映在后续 **`POST /v1/chat/completions`** 鉴权行为中（禁用/删除后 MUST 401，新增且启用后 MUST 可通过校验）。

#### Scenario: 禁用后立即拒绝

- **WHEN** 管理员将某 Key 设为禁用并生效
- **THEN** 使用该 Key 的 Bearer 请求 MUST 收到 **401** 且不调用上游

### Requirement: 批量操作

控制台 MUST 支持对多条 API Key 的**批量禁用**与**批量删除**（或等价批量状态变更）；操作 MUST 具备**全部成功或明确部分失败**的反馈（实现可选择事务或逐条错误列表）。

#### Scenario: 批量禁用至少两条

- **WHEN** 管理员选择至少两条 Key 并执行批量禁用
- **THEN** 所选 Key 在生效后均为禁用状态
