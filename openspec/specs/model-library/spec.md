# model-library Specification

## Purpose

本能力采用 **四表架构**，用关系表表达厂商、逻辑模型、上游 HTTP 端点与可路由实例：

| 表 | 职责 |
|----|------|
| **`model_vendor`** | 厂商维度（名称、类型、`vendor_code` 等） |
| **`model_base`** | 逻辑模型目录（全局 **`model_code`**、展示名、类型与能力等） |
| **`model_upstream`** | 某厂商下一条可访问的上游（**`base_url`、`api_key`**、超时与并发等） |
| **`model_instance`** | 将逻辑模型绑定到一条上游，并承载 **`provider_model_code`**、**`priority` / `weight` / `is_official`**、启用态等 |

字段、索引、外键（含 **`model_instance.upstream_id` → `model_upstream.id`**）与基数以 **`openspec/changes/archive/2026-05-02-model-library-aggregation/design.md`** 中 DDL 为准；部署与行为说明亦可参见 **`services/gateway/README.md`**（「模型库（四表聚合，无历史迁移）」等节）。

**废弃模型（非本能力）**：历史表 **`model_catalog_entry`**、**`model_upstream_binding`** 及「仅依赖 **`gateway.yaml`** 上游列表完成 Chat/List」的旧链路，**不是**本规范的数据模型。启用四表聚合时，**GET `/v1/models`** 与 **POST `/v1/chat/completions`** 仅以四表及实例选择结果为事实来源，**不得**与 YAML 声明的上游在同一请求中混选。**不提供**自旧 schema 的自动数据迁移。

## Requirements

### Requirement: 厂商（model_vendor）

系统 SHALL 维护 **`model_vendor`**，字段与 **`openspec/changes/archive/2026-05-02-model-library-aggregation/design.md`** 中 **`model_vendor` DDL** 一致（含 **`vendor_name`、`vendor_type`、`vendor_code`（UNIQUE）、`logo`、`status`**、时间戳等）。

#### Scenario: vendor_code 唯一

- **WHEN** 管理员提交与已有记录相同的 **`vendor_code`**
- **THEN** 创建失败并返回可机器读取的错误码（不与成功混淆）

### Requirement: 逻辑模型目录（model_base）

系统 SHALL 以 **`model_base`** 作为逻辑模型目录，字段与上述归档 **`design.md`** 中 **`model_base` DDL** 一致：**`model_name`**、**`model_code`**（UNIQUE）、**`model_type`**、**`capability`**、**`sort`**、**`status`** 等。**不考虑**自旧 **`model_catalog_entry`** 的迁移。

#### Scenario: 创建逻辑模型

- **WHEN** 管理员提交合法的新 **`model_code`** 与 **`model_name`**
- **THEN** 记录被持久化且可通过列表 API 检索

#### Scenario: model_code 唯一

- **WHEN** 管理员提交与已存在记录相同的 **`model_code`**
- **THEN** 创建失败并返回可机器读取的错误码（不与成功混淆）

### Requirement: 上游端点与路由实例（model_upstream / model_instance）

系统 SHALL 以 **`model_upstream`** 描述可 HTTP 访问的上游端点（**`base_url`、`api_key`**、**`timeout` / `max_concurrent`**、**`status`** 等）；以 **`model_instance`** 表达可调用路径：**`model_instance`** 关联 **`base_model_id`、`vendor_id`、`upstream_id`（→ `model_upstream.id`）**，并包含 **`provider_model_code`、`weight`、`priority`、`is_official`、`status`**。基数与关系见归档 **`openspec/changes/archive/2026-05-02-model-library-aggregation/design.md`** 中四表说明（一厂商多上游；一逻辑模型多实例；一实例指向一条 **`model_upstream`**）。

#### Scenario: 停用实例

- **WHEN** 管理员将 **`model_instance.status`** 设为 **0**
- **THEN** 该实例不参与公开模型列表聚合，且不参与 Chat 路径转发

### Requirement: 管理端 CRUD API

系统 SHALL 在 **`/api/admin/v1`** 下提供模型库 REST（路径以实现为准）：**`model_vendor`、`model_base`、`model_upstream`、`model_instance`** 的列表（分页或游标）、创建、更新、删除；并保留**对指定 `model_upstream`（按主键或 `base_url`）调用 HTTP GET `{base_url}/v1/models`** 的同步能力。所有写操作 MUST 受与现有管理端一致的 **JWT 与角色** 约束（操作员只读策略与现有一致）。

#### Scenario: 未授权拒绝写

- **WHEN** 请求无有效管理 JWT 调用写接口
- **THEN** 响应为 **401**，且 body 符合网关统一错误约定

### Requirement: 上游模型同步

系统 SHALL 提供管理端操作：**对指定 `model_upstream` 使用其 `base_url` 与 `api_key` 调用 HTTP GET `{base_url}/v1/models`**（路径前缀以实现为准）；凭证 MUST NOT 以明文写入持久化审计日志；将返回中的模型 id 与 **`model_base` / 实例导入建议** 合并或预览。

#### Scenario: 上游不可达

- **WHEN** 同步请求连接失败或超时
- **THEN** 返回 **4xx/5xx** 与明确 `code`，且不部分写入未定义状态（事务或空操作）

### Requirement: 公开列表的数据来源

**GET `/v1/models`** 返回的条目 MUST 仅来自：存在至少一条 **`model_instance`** 满足 **`status=1`**，且 **`model_base`、`model_vendor`、其 `model_upstream` 均为 `status=1`**；**`data[].id`** MUST 为 **`model_base.model_code`**；**`owned_by`** MAY 来自 **`model_vendor.vendor_name`** 或占位。

#### Scenario: 空库

- **WHEN** 无任何满足条件的实例
- **THEN** **GET `/v1/models`** 响应为 **`object: list`** 且 **`data` 为空数组**（与 OpenAI 空列表行为一致）

### Requirement: 安全与密钥

模型同步与上游 HTTP 调用 MUST NOT 将 Bearer 密钥写入 Zap 或返回给客户端；管理 API MUST NOT 在 JSON 中回显用户粘贴的临时密钥或完整 **`api_key`**（若需状态仅返回布尔或掩码）。

#### Scenario: 日志无密钥

- **WHEN** 同步失败并记录 Zap
- **THEN** 日志中无 `Authorization` 原文、完整 Bearer 串或完整 **`api_key`**

### Requirement: Chat Completions 请求体 `model` 改写

在 **POST `/v1/chat/completions`** 转发路径上，网关 SHALL 在 **已按四表实例选择策略选中 `model_instance` 及其 `model_upstream`** 后，尝试将请求体解析为 JSON **对象**；若成功且存在字符串字段 **`model`**，且命中 **`model_base.model_code`** 对应该实例，则 SHALL 将转发给上游 HTTP 的 JSON 中 **`model`** 设为 **`provider_model_code`**。若未走模型库选择或未命中实例，则 SHALL 不更改 **`model`**（除非其他路由另有规则）。若 body 无法解析为 JSON 对象，则 SHALL **原样转发**原始字节流。

#### Scenario: provider_model_code 生效

- **WHEN** 客户端 body 含 **`"model":"gpt-3.5-turbo"`**，选中实例 **`provider_model_code`** 为 **`gpt-3.5-turbo-0125`**
- **THEN** 发往 **`model_upstream.base_url`** 的请求体中 **`model`** 为 **`gpt-3.5-turbo-0125`**

#### Scenario: 无匹配不改写

- **WHEN** 无匹配实例或未走模型库路径
- **THEN** 上游收到的 body 与客户端一致（除 JSON 键重排等无害差异外）

### Requirement: 上游探针与临时剔除（产品范围外，不实现）

对 **`model_upstream` / `model_instance`** 的**周期性可达性探针**、失败时**临时剔出选择池**等行为 **已确定为产品范围外，不实现**（对应实现任务 **2.4** 关闭，非「仅首版延后」）。实例选择仅依据库内 **`status` 与 `priority` / `is_official` / `weight`** 等已持久化字段及既定算法；**不**依据运行时探活结果。

若未来单独立项引入探针，MUST 另起规范修订，且 MUST 遵守本文 **安全与密钥** 条款（日志与响应中不得泄露 **`api_key`**）。

#### Scenario: 验收不包含探针剔除

- **WHEN** 审查者验收模型库与网关转发
- **THEN** MUST NOT 将「探针剔除不可达上游」列为必测项
