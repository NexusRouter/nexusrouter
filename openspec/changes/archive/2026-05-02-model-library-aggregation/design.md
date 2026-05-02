## Context

NexusRouter 为 OpenAI 兼容 LLM 网关。本变更以**库表为唯一事实来源**定义厂商、逻辑模型、上游 HTTP 端点与可调用实例；**不要求**自旧 `model_catalog_entry` / `model_upstream_binding` 导入数据，**直接建库使用**四张表即可。

## Goals / Non-Goals

**Goals:**

- 四表 **`model_vendor`、`model_base`、`model_upstream`、`model_instance`** 的字段、类型、约束与下列「生产级 DDL」一致（实现可选用 GORM/SQL 迁移；SQLite 下 bigint 映射为 INTEGER 等细节以实现为准）。
- 网关对 **`model=gpt-3.5-turbo`** 只解析 **`model_base.model_code`**，在 **`model_instance.status=1`** 且关联 **`model_vendor` / `model_base` / `model_upstream`** 均为启用态（**`status=1`**）的集合中选实例；按 **`priority` 数值小优先**（1=高、2=中、3=低）、同档 **`weight`** 负载、**`is_official`** 辅助官方优先策略。
- 选中实例后，HTTP 转发目标为 **`model_upstream.base_url`**，鉴权使用该行 **`api_key`**（及 **`timeout` / `max_concurrent`** 等策略）；请求体中发往厂商的真实模型名为 **`model_instance.provider_model_code`**。
- 管理端可按厂商、按逻辑模型筛选；与现有 admin 鉴权一致。

**Non-Goals:**

- **历史数据迁移**、双读旧表、回滚到旧 schema。
- 不改变 OpenAI 兼容路径 **`/v1/chat/completions`**、**`/v1/models`**。
- 健康检查探针协议以实现为准。

## 生产级数据模型（DDL 摘要）

以下与实现字段名一致；`datetime` 以实现选用时间类型。

### 1. `model_vendor`

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK | 主键 |
| vendor_name | varchar(64) | NOT NULL | 厂商名称 |
| vendor_type | tinyint | NOT NULL | 1=官方原厂 2=第三方聚合 |
| vendor_code | varchar(32) | UNIQUE | openai/zhipu/oneapi 等 |
| logo | varchar(512) | NULL | 图标 URL |
| status | tinyint | DEFAULT 1 | 1=启用 0=禁用 |
| created_at | datetime | NOT NULL | |
| updated_at | datetime | NOT NULL | |

### 2. `model_base`

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK | 主键 |
| model_name | varchar(64) | NOT NULL | 展示名 |
| model_code | varchar(64) | UNIQUE | 全局逻辑 id，如 gpt-3.5-turbo |
| model_type | tinyint | NOT NULL | 1=对话 2=Embedding 3=图像 4=语音 |
| capability | json | NULL | 上下文窗口等 |
| sort | int | DEFAULT 0 | 排序 |
| status | tinyint | DEFAULT 1 | 1=启用 0=禁用 |
| created_at | datetime | NOT NULL | |
| updated_at | datetime | NOT NULL | |

### 3. `model_upstream`

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK | 主键 |
| vendor_id | bigint | NOT NULL | FK → model_vendor |
| upstream_name | varchar(64) | NOT NULL | 上游名称 |
| base_url | varchar(512) | NOT NULL | 接口根地址 |
| api_key | varchar(512) | NULL | 密钥 |
| timeout | int | DEFAULT 30 | 秒 |
| max_concurrent | int | DEFAULT 100 | 最大并发 |
| status | tinyint | DEFAULT 1 | 1=启用 0=禁用 |
| created_at | datetime | NOT NULL | |
| updated_at | datetime | NOT NULL | |

### 4. `model_instance`

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | bigint | PK | 主键 |
| base_model_id | bigint | NOT NULL | FK → model_base |
| vendor_id | bigint | NOT NULL | FK → model_vendor |
| upstream_id | bigint | NOT NULL | FK → **model_upstream.id**（上游服务行） |
| instance_name | varchar(64) | NOT NULL | 实例展示名 |
| provider_model_code | varchar(64) | NOT NULL | 厂商侧真实 model 名 |
| weight | int | DEFAULT 10 | 负载均衡权重 |
| priority | tinyint | DEFAULT 1 | 1=高 2=中 3=低（数值小优先） |
| is_official | tinyint | DEFAULT 0 | 1=官方 0=非官方 |
| status | tinyint | DEFAULT 1 | 1=启用 0=禁用 |
| created_at | datetime | NOT NULL | |
| updated_at | datetime | NOT NULL | |

**关系：** 一厂商多上游；一逻辑模型多实例；一实例指向**一条** `model_upstream`（字段名 **`upstream_id`** 表示 FK，勿与网关 YAML 混淆）。

## Decisions

| 决策 | 选择 | 理由 |
|------|------|------|
| 转发地址来源 | **`model_upstream.base_url` + `api_key`** | 与提供 DDL 一致，单库闭环 |
| 逻辑名 vs 厂商名 | 客户端 **`model`** 对齐 **`model_code`**；上游 JSON **`model`** 使用 **`provider_model_code`** | 调用方无感；厂商真实名落库 |
| 启停 | 统一用各表 **`status`**（1/0），不用单独 `enabled` 布尔 | 与 DDL 一致 |
| 优先级 | **`priority` ASC** 为主序；**`is_official`** 作同档tie-break（实现文档化） | 满足官方优先与多档降级 |
| 与静态网关配置关系 | **不并存**：OpenAI 兼容的 **`/v1/chat/completions`** / **`/v1/models`** 在启用本能力时，**仅以四表 + `model_instance` 选择结果**为转发与列表来源；**不得**与 **`gateway.yaml` 中的 Upstream 列表混用或并行二选一 | 单一事实来源，避免双配置冲突；实现按本变更 **`priority` / `weight` / `is_official` / `status`** 规则优先 |

## Risks / Trade-offs

- **[Risk] 运维习惯仍维护 `gateway.yaml` 上游** → **Mitigation**：文档明确「模型库模式」下转发不读 YAML 上游；非模型库路由（若有）仍以既有逻辑为准，二者**不按同一请求混用**。
- **[Risk] `api_key` 落库** → **Mitigation**：加密存储或密钥管理服务由实现定；日志与 API 回显遵循既有安全规范。

## Open Questions

- `max_concurrent` 在进程内如何与全局限流协同，以实现为准。
- 管理 API 子路径命名（`/vendors`、`/bases`、`/upstreams`、`/instances`）与 OpenAPI 同步即可。
