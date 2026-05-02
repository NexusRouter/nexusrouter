# admin-upstream-management Specification

## Purpose
TBD - created by archiving change core-admin-management. Update Purpose after archive.
## Requirements
### Requirement: 上游列表的可视化维护

管理控制台 MUST 提供上游条目的**列表与表单**，支持 **`id`**、**`base_url`**、**`weight`**（及与现有 `gateway.yaml` 模型一致的字段）的**新增、编辑、删除**。提交变更 MUST 触发运行时更新，使后续代理选择使用新快照（见「实时生效」）。

#### Scenario: 新增合法上游后可选中

- **WHEN** 管理员新增一条通过校验的上游并保存
- **THEN** 该上游出现在列表中，且运行时快照包含该条目（以查询接口或后续请求行为可验证）

#### Scenario: 删除上游后不可再被选

- **WHEN** 管理员删除某上游且该条目不再出现在配置中
- **THEN** **`default_upstream_id`** 与 **`active_upstream_id`** MUST NOT 引用已删除的 **`id`**；若冲突 MUST 拒绝保存并提示

### Requirement: 默认上游

控制台 MUST 支持设置 **`default_upstream_id`**（或运行时等价字段），且 MUST 与 **`upstream-target-management`** 语义一致：默认上游 MUST 为已存在 **`id`**。

#### Scenario: 设置不存在的默认被拒绝

- **WHEN** 管理员将默认上游设为不存在的 **`id`**
- **THEN** 保存失败并返回可理解错误，且运行时状态不变

### Requirement: 实时生效与当前生效上游可观测

上游配置变更 MUST **实时生效**于网关运行时（无需单独重启进程）。界面 MUST 展示**当前生效**的上游解析结果：至少包括**当前 `active_upstream_id` pin**（若有）及解析到的 **base_url** 或 **`id`**（与 `design.md` 展示层级一致）。

#### Scenario: 修改 active 后立即影响新请求

- **WHEN** 管理员将 pin 的上游切换为另一合法 **`id`**
- **THEN** 新发起的受保护业务请求 MUST 使用新上游（在无其他覆盖策略时）

