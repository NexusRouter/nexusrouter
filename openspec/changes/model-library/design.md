## Context

NexusRouter 当前以 **`runtime.Snapshot`** 中的上游列表与 **`Picker`** 选择转发目标；**未**按请求体中的 `model` 字段做细粒度分流（与多通道矩阵型网关不同）。模型库首版聚焦：**目录真相源在数据库**、**对客户端声明可用模型集合**、**为后续「按模型选路 / 重写 model」打基础**。  
设计参考了常见 LLM 聚合网关中的实践（静态模型清单、通道—模型能力表、模型别名、OpenAI 式 list models），但实现与文档**独立**，不绑定任何外部代码库。

## Goals / Non-Goals

**Goals:**

- 管理端可维护「逻辑模型 id」及展示元数据（名称、可选分组标签、可选上下文长度等）。
- 将逻辑模型**绑定**到已存在的上游 **`upstream_id`**（与 `gateway.yaml`/DB 中一致）；支持启用/停用绑定。
- 支持 **alias**：客户端请求的 `model` 可映射为实际上游模型名（与上游配置中的 `model_mapping` 可并存；冲突时以设计优先级解析，建议：**请求级 alias 优先生效**或**显式文档化顺序**）。
- **GET `/v1/models`**：仅返回当前网关密钥有权感知的、已启用绑定所对应的模型列表（首版可与「全部已启用模型」一致，细粒度按密钥过滤留给后续）。
- **同步**：对指定上游基址发起 **GET `/v1/models`**（使用配置中的 `UpstreamAPIKey` 或同步请求中一次性传入的密钥，**响应不落库**），将返回的 `id` 与目录合并或批量建议导入。

**Non-Goals（首版）：**

- 替换现有上游选择策略为「按模型全自动选路」（可后续变更）。
- 完整计费 / 配额 / 用户分组矩阵（独立计费引擎不在本设计展开）。
- 在网关内实现全量供应商适配表维护（可采用**可选**内置种子 JSON，以版本号迭代）。

## Decisions

### D1: 数据模型（3NF 为主）

| 实体 | 说明 |
|------|------|
| `model_catalog_entry` | 逻辑模型：`id`（主键，稳定字符串，如 `gpt-4o`）、`display_name`、`owned_by`（可选）、`metadata`（JSON 扩展）、`created_at` / `updated_at` |
| `model_upstream_binding` | `catalog_entry_id`、`upstream_id`（字符串，对齐 snapshot）、`enabled`、`priority`（整数，备后续多绑定竞争）、可选 `actual_model`（别名指向的上游模型名，空则与 `catalog_entry.id` 相同） |
| 可选：`model_sync_job` 仅内存或审计表记录最后一次同步时间与错误 | |

**理由**：目录与绑定分离，符合 `database-design-standards` 的归一化；展示冗余字段通过 JOIN 或轻量视图查询。

### D2: 公开 API 行为

- **GET `/v1/models`**：鉴权与 **POST `/v1/chat/completions`** 相同；响应 `object: list`，`data[]` 元素含 `id`、`object: model`、`created`、`owned_by`（可来自绑定或目录）。
- **GET `/v1/models/:model`**：若 id 存在且可用则 **200**；否则 OpenAI 风格 **error** JSON（与现有网关错误包装策略一致，或沿用 `RetrieveModel` 常见形态）。
- **未配置任何启用模型时**：返回空列表或 **503**（择一并在实现中固定，规范中写死一种）。

### D3: 同步实现

- 服务端使用 `net/http` 对 **`{upstream_base}/v1/models`** 发起 GET，`Authorization: Bearer <key>` 来自环境变量或管理端表单仅用于该次请求。
- 超时、TLS、最大 body 限制与 Zap 日志（无密钥原文）。

### D4: 控制台

- 新路由如 `/model-library`，侧栏一项；表格 + 抽屉编辑；同步按钮选择上游（下拉来自 `gateway/snapshot` 或等价 API）。

### D5: 与 chat 代理的关系（首版必做）

- **`POST /v1/chat/completions`** 在 **已选定上游 `upstream_id`**（与 **`Picker`** 结果一致）后，对 **可解析为 JSON 对象** 的请求体读取 **`model`** 字符串。
- 若数据库中存在 **启用** 的绑定 **`(catalog_entry_id = 请求 model, upstream_id = 当前上游)`**：
  - 当 **`actual_model` 非空**：上游收到的 body 中 **`model`** MUST 替换为该值（trim 后）。
  - 当 **`actual_model` 为空或未设置**：保持客户端 **`model`** 为目录项 id（通常无需改写字节）。
- 若 **无匹配绑定**：请求体 **原样** 转发（与无模型库时一致）。
- 若 body **不是合法 JSON 对象**（或缺少 **`model`**）：不进行改写，**原样** 转发字节流（兼容非标准客户端）。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 同步上游返回模型数过大 | 限制导入条数、分页拉取（若上游支持）、仅增量建议 |
| 与现有上游 `model_mapping` 双重映射混淆 | 文档与控制台明示优先级；单测覆盖 |
| 公开 list 泄露未授权模型 | 首版绑定即授权；后续按 API Key 维度过滤 |

## Migration Plan

1. 迁移添加表；默认空目录，不影响现有转发。
2. 部署后管理员在控制台导入或同步。
3. 回滚：迁移 down 或 feature flag 关闭 **GET `/v1/models`** 注册。

## Open Questions

- 是否内置一份「常用模型种子」JSON 版本化随版本发布（减少冷启动空库）。
- API Key 级模型可见性是否与 `api-key-management` 扩展字段挂钩（后续）。
