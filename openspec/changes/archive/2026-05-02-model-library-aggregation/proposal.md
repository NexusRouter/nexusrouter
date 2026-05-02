## Why

当前 `model-library` 能力以「目录项 + 单条上游绑定」为主，难以表达**同一逻辑模型名**在多家厂商、多条上游之间的冗余、优先级与权重分流；运维也无法按「厂商 / 逻辑模型 / 实例」三个维度统一管理。需要引入**生产级三层解耦**（厂商 → 逻辑模型 → 上游服务行 → 模型实例），使调用方始终只使用统一 `model` 标识（如 `gpt-3.5-turbo`），由网关在实例间按策略择优转发。

## What Changes

- 采用 **4 张核心表**（**无历史数据迁移**，全新建库即可），字段与类型以设计文档中的 DDL 为准：
  - **`model_vendor`**：厂商（名称、类型、唯一 `vendor_code`、logo、`status` 等）
  - **`model_base`**：逻辑模型（`model_name`、`model_code`、`model_type`、`capability` JSON、`sort`、`status` 等）
  - **`model_upstream`**：上游服务（`vendor_id`、`upstream_name`、`base_url`、`api_key`、`timeout`、`max_concurrent`、`status` 等）
  - **`model_instance`**：可调用实例（`base_model_id`、`vendor_id`、`upstream_id` → `model_upstream.id`、`provider_model_code`、`weight`、`priority`、`is_official`、`status` 等）
- 网关 **Chat Completions /v1/models**：按逻辑模型聚合；转发使用选中实例关联的 **`model_upstream.base_url` / `api_key`**；按 **priority → weight → is_official** 与 `status` 选择实例。
- 管理端：按厂商、按逻辑模型筛选；实例级启停；与现有 JWT/RBAC 一致。

## Capabilities

### New Capabilities

- `model-library-aggregation`：上述四表 DDL、关系不变量、实例选择与转发数据来源（库表内 `base_url`，不依赖旧目录表迁移）。

### Modified Capabilities

- `model-library`：持久化与绑定语义与四表 DDL 对齐；管理 CRUD、同步与公开列表数据来源对齐新表。
- `gateway-backend`：Chat 路径按 `model_code` 选实例，从 **`model_upstream`** 行解析转发地址与凭证。
- `openai-chat-completions-proxy`：`GET /v1/models` 与 Chat 路径下模型语义与聚合实例一致。
- `dashboard-frontend`：模型库 UI 支持厂商 / 逻辑模型 / 上游 / 实例视图与操作。

## Impact

- **后端**（`services/gateway`）：GORM 实体、迁移（仅新建四表）、repository、管理 API、Chat 转发链（**HTTP 客户端目标为 `model_upstream.base_url`**，请求体 **`model`** 改写为 **`provider_model_code`**）。
- **前端**（`web/dashboard`）：模型库页面与表单字段对齐四表。
- **数据库**：**仅新增四张表**；不考虑自 `model_catalog_entry` / `model_upstream_binding` 的迁移。
- **与 `gateway.yaml`：** **不并存**——启用本模型库聚合路径时，**`/v1/chat/completions` 与 `/v1/models` 仅以四表及实例选择逻辑为准**，不与 YAML 中的 Upstream 配置混用；选择顺序仍按本变更的 **`priority` → `weight` → `is_official`** 及 **`status`**。
