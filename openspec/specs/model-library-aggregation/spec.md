# model-library-aggregation Specification

## Purpose

定义模型库 **四表** 持久化模型、字段约束与实例选择语义；详细列定义以归档变更内 **`design.md`** 中 DDL 摘要为准（路径见 **`openspec/specs/model-library`** Purpose）。本规范不涵盖历史数据迁移。启用本能力时，**OpenAI 兼容 Chat / List models 与 `gateway.yaml` 中的 Upstream 配置不并存**，仅以四表及本文实例选择规则为唯一来源。

## Requirements

### Requirement: 模型厂商表 model_vendor

系统 SHALL 维护 **`model_vendor`**，包含字段：**`id`**（PK）、**`vendor_name`**、**`vendor_type`**（1=官方原厂 2=第三方聚合）、**`vendor_code`**（UNIQUE）、**`logo`**（可空）、**`status`**（1=启用 0=禁用）、**`created_at`**、**`updated_at`**。

#### Scenario: vendor_code 唯一

- **WHEN** 管理员提交与已有记录相同的 **`vendor_code`**
- **THEN** 创建失败并返回可机器读取的错误码

### Requirement: 基础逻辑模型表 model_base

系统 SHALL 维护 **`model_base`**，包含字段：**`id`**（PK）、**`model_name`**、**`model_code`**（UNIQUE，全局逻辑标识）、**`model_type`**（1=对话 2=Embedding 3=图像 4=语音）、**`capability`**（JSON，可空）、**`sort`**、**`status`**（1=启用 0=禁用）、**`created_at`**、**`updated_at`**。

#### Scenario: model_code 唯一

- **WHEN** 写入与已存在相同的 **`model_code`**（trim 策略实现文档化）
- **THEN** 违反唯一约束并返回错误

### Requirement: 模型上游服务表 model_upstream

系统 SHALL 维护 **`model_upstream`**，表示某厂商下一条可 HTTP 访问的上游端点，包含字段：**`id`**（PK）、**`vendor_id`**（FK → `model_vendor`）、**`upstream_name`**、**`base_url`**、**`api_key`**（可空）、**`timeout`**（默认 30）、**`max_concurrent`**（默认 100）、**`status`**（1=启用 0=禁用）、**`created_at`**、**`updated_at`**。网关转发 Chat 时 MUST 使用选中实例对应行的 **`base_url`** 与 **`api_key`**（若为空则行为以实现与错误处理为准且文档化）。

#### Scenario: 厂商维度一对多上游

- **WHEN** 同一 **`vendor_id`** 下存在多条 **`model_upstream`**
- **THEN** 每条独立寻址，互不覆盖 **`base_url`**

### Requirement: 模型实例表 model_instance

系统 SHALL 维护 **`model_instance`**，包含字段：**`id`**（PK）、**`base_model_id`**（FK → `model_base`）、**`vendor_id`**（FK → `model_vendor`）、**`upstream_id`**（FK → **`model_upstream.id`**）、**`instance_name`**、**`provider_model_code`**（厂商侧真实模型名）、**`weight`**（默认 10）、**`priority`**（1=高 2=中 3=低，**数值越小越优先**）、**`is_official`**（1/0）、**`status`**（1=启用 0=禁用）、**`created_at`**、**`updated_at`**。

#### Scenario: 停用实例不参与

- **WHEN** **`model_instance.status=0`** 或关联 **`model_base`/`model_vendor`/`model_upstream`** 任一为禁用
- **THEN** 该实例不参与 **GET `/v1/models`** 聚合，且不参与 Chat 路径实例选择

### Requirement: 关系基数不变量

数据 SHALL 满足：**一厂商**多条 **`model_upstream`**；**一逻辑模型**多条 **`model_instance`**；**一 `model_instance`** 精确引用 **一条 `model_upstream`**（**`upstream_id` 列为 FK**）。

#### Scenario: 审查

- **WHEN** 审查者检查 ER 或迁移
- **THEN** 可见上述 FK 与基数

### Requirement: 实例选择策略

在 **POST `/v1/chat/completions`** 上，对 **`model`** 解析 **`model_base.model_code`** 后，网关 SHALL 在 **`model_instance.status=1`** 且关联 **`model_base`、`model_vendor`、`model_upstream` 均为 `status=1`** 的实例集合中选择 **至多一条**：**先按 `priority` ASC**；同 **`priority`** 档内 **`is_official=1` 优先于 0**；同档再按 **`weight`** 做加权分配。选中后转发 HTTP 目标为 **`model_upstream.base_url`**，请求体中 **`model`** SHALL 改写为 **`provider_model_code`**（见 **`model-library`** 改写要求）。

#### Scenario: 高档无可用则降级

- **WHEN** **`priority=1`** 无可用实例
- **THEN** 网关 SHALL 尝试 **`priority=2`**、再 **`priority=3`**，直至无可用则返回 **`gateway-backend`** 一致错误

### Requirement: 上游探针（产品范围外）

系统 SHALL NOT 实现 **`model_upstream` / `model_instance`** 的周期性可达性探针，亦 SHALL NOT 因探针临时剔出选择池；与 **`model-library`** 中「上游探针与临时剔除」一致。实例选择 MUST 仅依据库内 **`status` 与 `priority` / `is_official` / `weight`**。

#### Scenario: 选择不依赖探针

- **WHEN** 网关为 Chat 路径选择 **`model_instance`**
- **THEN** MUST NOT 以探针可达性作为选择依据；仅使用库表字段与既定算法

### Requirement: 管理视角筛选

管理 API SHALL 支持按 **厂商**、按 **`model_code`** 列出实例；支持修改 **`model_instance.status`** 与各表 **`status`**。

#### Scenario: 按 model_code 列出实例

- **WHEN** 管理员按 **`model_code`** 查询
- **THEN** 返回实例及关联厂商、上游摘要字段
