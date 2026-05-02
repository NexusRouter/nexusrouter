# model-library Specification

## Purpose
TBD - created by archiving change model-library. Update Purpose after archive.
## Requirements
### Requirement: 模型目录项持久化

系统 SHALL 在关系型存储中维护 **`model_catalog_entry`**（或等价命名），至少包含：**稳定字符串主键 `id`**（与 OpenAI `model` 标识兼容）、**`display_name`**、**`created_at`/`updated_at`**；MAY 包含 **`owned_by`**、**`metadata` JSON**（扩展字段如 `context_length`）。

#### Scenario: 创建目录项

- **WHEN** 管理员通过管理 API 提交合法的新模型 id 与展示名
- **THEN** 记录被持久化且可通过列表 API 检索

#### Scenario: id 唯一

- **WHEN** 管理员提交与已存在记录相同的 `id`
- **THEN** 创建失败并返回可机器读取的错误码（不与成功混淆）

### Requirement: 上游绑定

系统 SHALL 支持 **`model_upstream_binding`**，将目录项 **`catalog_entry_id`** 与运行时配置中的 **`upstream_id`**（字符串，与 `gateway` 上游列表一致）关联，并包含 **`enabled`** 布尔与 **`priority`** 整型；MAY 包含 **`actual_model`**（非空时表示转发到上游使用的模型名字符串）。

#### Scenario: 停用绑定

- **WHEN** 管理员将某绑定的 `enabled` 设为 false
- **THEN** 该绑定不参与公开模型列表聚合，且不参与（若已实现）模型级转发

### Requirement: 管理端 CRUD API

系统 SHALL 在 **`/api/admin/v1`** 下提供模型库 REST：**列表（分页或游标）**、**创建目录项**、**更新目录项**、**删除目录项**、**创建/更新/删除绑定**；所有写操作 MUST 受与现有管理端一致的 **JWT 与角色** 约束（操作员只读策略与现有一致）。

#### Scenario: 未授权拒绝写

- **WHEN** 请求无有效管理 JWT 调用写接口
- **THEN** 响应为 **401**，且 body 符合网关统一错误约定

### Requirement: 上游模型同步

系统 SHALL 提供管理端操作：**对指定 `upstream_id` 调用该上游的 HTTP GET `{base}/v1/models`**，使用配置中的上游凭证或单次请求提供的凭证（凭证 MUST NOT 以明文写入持久化审计日志）；将返回中的模型 id 与目录进行**合并或批量导入建议**（具体 UX 由控制台实现，服务端至少提供原始列表或导入预览数据结构）。

#### Scenario: 上游不可达

- **WHEN** 同步请求连接失败或超时
- **THEN** 返回 **4xx/5xx** 与明确 `code`，且不部分写入未定义状态（事务或空操作）

### Requirement: 公开列表的数据来源

**GET `/v1/models`** 返回的条目 MUST 仅来自：**存在至少一条 `enabled=true` 的绑定**且其 **`upstream_id`** 在当前运行时快照中存在；目录项元数据用于填充 `id` 与展示相关字段。

#### Scenario: 空库

- **WHEN** 无任何启用绑定
- **THEN** **GET `/v1/models`** 响应为 **`object: list`** 且 **`data` 为空数组**（与 OpenAI 空列表行为一致）

### Requirement: 安全与密钥

模型同步与上游 HTTP 调用 MUST NOT 将 Bearer 密钥写入 Zap 或返回给客户端；管理 API MUST NOT 在 JSON 中回显用户粘贴的临时密钥（若需回显状态仅返回布尔成功）。

#### Scenario: 日志无密钥

- **WHEN** 同步失败并记录 Zap
- **THEN** 日志中无 `Authorization` 原文或完整 Bearer 串

### Requirement: Chat Completions 请求体 `model` 改写（首版必做）

在 **POST `/v1/chat/completions`** 转发路径上，网关 SHALL 在 **已确定当前命中的 `upstream_id`**（与运行时 **`Picker`** 一致）后，尝试将请求体解析为 JSON **对象**；若成功且存在字符串字段 **`model`**，则 SHALL 查询启用绑定 **`catalog_entry_id = trim(model)`** 且 **`upstream_id`** 等于当前命中上游。若绑定存在且 **`actual_model`** 非空（trim 后），则 SHALL 将转发给上游的 JSON 中 **`model`** 替换为该值；若绑定存在且 **`actual_model`** 为空，则 SHALL 不更改 **`model`** 字段值；若绑定不存在，则 SHALL 不更改 body。若 body 无法解析为 JSON 对象，则 SHALL **原样转发**原始字节流。

#### Scenario: actual_model 生效

- **WHEN** 客户端 body 含 **`"model":"logical-id"`**，且存在启用绑定 **`logical-id` → 当前上游**，**`actual_model`** 为 **`upstream-real`**
- **THEN** 上游 HTTP 请求体解析后 **`model`** 为 **`upstream-real`**

#### Scenario: 无绑定不改写

- **WHEN** 无匹配绑定
- **THEN** 上游收到的 body 与客户端一致（除 JSON 键重排等无害差异外）

