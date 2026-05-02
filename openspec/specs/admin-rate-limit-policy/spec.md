# admin-rate-limit-policy Specification

## Purpose
TBD - created by archiving change advanced-admin-management. Update Purpose after archive.
## Requirements
### Requirement: 多条限流规则模型

网关 MUST 支持在运行时配置中声明**零条或多条**限流规则；每条规则 MUST 至少包含：**稳定 `id`**、**维度**（`ip` 或 `api_key_fp`）、**RPS 阈值**、**突发 `burst`（可选，默认由实现推导）**、**`enabled` 布尔**、**`priority` 整数**（越大越优先）、**`match_path_prefix`**（空字符串表示匹配所有路径）。匹配时 MUST 自高优先级向低遍历，**首条维度命中且路径匹配**的规则生效；若均不命中，MUST 回退至既有全局 **`rps_per_ip` / `rps_per_key`**（若其大于 0）。

#### Scenario: 高优先级规则覆盖低优先级

- **WHEN** 存在两条同维度规则且路径均匹配，优先级不同
- **THEN** 仅**较高 `priority`** 的阈值应用于该请求

#### Scenario: 禁用规则不参与匹配

- **WHEN** 某规则 `enabled` 为 false
- **THEN** 该规则 MUST 被跳过且不影响其它规则或全局回退

### Requirement: 管理端规则 CRUD 与实时生效

管理控制台 MUST 支持规则的**新增、编辑、删除**；变更 MUST 在成功持久化后 **Reload** 或等价机制下对**新请求**立即生效，且失败时 MUST 保留旧快照。

#### Scenario: 非法 burst 或 rps 被拒绝

- **WHEN** 提交非正数 RPS 或 burst 小于 1
- **THEN** 保存失败并返回校验错误，且运行时配置不变

