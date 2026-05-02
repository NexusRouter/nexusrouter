## Why

NexusRouter 管理控制台「系统设置」页当前将**运行中可见的配置键值列表**与**代理访问日志表单**上下堆叠，二者缺乏场景化关联；界面以 `proxy_access_log_*` 等后端键名为主标签，「热更新 / 需重启 / 只读」等术语缺少与生效方式、修改路径的对应说明，导致用户难以判断**能改什么、怎么改、改了会怎样**。在**不调整网关对 `gateway.yaml` 及管理 API 的语义**前提下，需对 `/settings` 做一次信息架构与交互重构，使运维与管理员能按意图完成任务并降低误操作风险。

## What Changes

- **页面架构**：由「扁平变量列表 + 独立日志表单」改为三大场景模块——**系统运行状态（只读聚合）**、**代理访问日志（可编辑表单）**、**高级配置（可选折叠）**；首屏一眼区分「看状态 / 改日志 / 查细节」。
- **术语平民化**：用户主文案统一为业务语言（如「HTTP 监听地址」「上游请求超时」）；后端键名（如 `http_listen_addr`）仅作次要信息或 Tooltip，与 `zh`/`en` 资源 **100% 同源**，禁止中英文各写一套不一致描述。
- **交互与联动**：日志配置保留现有 PUT 载荷与校验；保存成功后 **invalidate** 与运行状态区同源数据，使表单与只读区展示一致；对需重启/只读项提供**明确操作指引**（环境变量名、须重启等），而非仅展示枚举标签。
- **状态与风险透明**：每个配置项展示 **生效方式**（热更新 / 需重启进程 / 仅展示）及简短**风险提示**（如写盘、路径无效、服务不可用等场景引用现有 API 错误）；「写回 gateway.yaml」与保存按钮关系说明清楚。
- **非目标**：不修改 Go 侧 `GET/PUT /api/admin/v1/system/settings` 的请求与响应 schema、不新增字段需求（若未来后端补充 `max_size_mb` 等在 GET 中的展示，前端可再对齐）；不改变 Operator 只读与 Admin 可写规则。

## Capabilities

### New Capabilities

- `system-settings-page-ux`：约束 `/settings` 的场景化布局、术语与 i18n 一致性、日志表单与运行状态联动、生效方式与风险提示的呈现规则；与现有 `admin-system-settings` 所定义的 API 行为一致，仅增加**前端体验层**需求。

### Modified Capabilities

- （无）后端「系统设置」读写与 mutability 语义不变；本变更通过新增 `system-settings-page-ux` 规格承载 UI/UX，避免与 API 规格混淆。

## Impact

- **前端**：`web/dashboard/src/pages/SystemSettings.tsx`（或拆分出的子组件）、`web/dashboard/src/locales/zh.ts` 与 `en.ts`；可复用与网关策略页类似的 `TermHint`、卡片/折叠等模式（若项目已有）。
- **后端**：无变更预期；仍消费 `settings[]` 的 `key` / `value` / `mutability` / `hint` 与 `PUT` body 中的 `proxy_access_log`、`persist`。
- **配置**：继续兼容现有 `gateway.yaml` 与热更新流程；无数据迁移。
