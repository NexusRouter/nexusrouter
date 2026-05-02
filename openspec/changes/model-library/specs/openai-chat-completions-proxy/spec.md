# openai-chat-completions-proxy Specification（变更 `model-library` — 增量）

## ADDED Requirements

### Requirement: OpenAI 兼容模型列表端点

网关 SHALL 注册 **GET `/v1/models`**，且 MUST 在与 **POST `/v1/chat/completions`** 相同的网关入口鉴权链下执行（**Bearer** 优先，可选 **`X-API-Key`** 若项目仍声明兼容）。成功响应 MUST 为 JSON，且 **`object`** 字段为 **`list`**，**`data`** 为数组；数组元素 MUST 至少包含 **`id`**、**`object`**（值为 **`model`**）、**`created`**（Unix 秒，整数）、**`owned_by`**（字符串，可为占位）。

#### Scenario: 鉴权失败不列模型

- **WHEN** 请求未携带有效网关凭证
- **THEN** 响应 **401** 且**不**返回模型列表 body

#### Scenario: 形状可被客户端解析

- **WHEN** 客户端解析成功响应 JSON
- **THEN** 可读取 **`data`** 数组且每项含 **`id`**

### Requirement: 可选模型检索端点

网关 MAY 注册 **GET `/v1/models/:model`**；若模型 id 不可用，SHALL 返回 **404** 或 OpenAI 常见错误包装（与实现选定一致，且在全仓库唯一）。

#### Scenario: 未知模型

- **WHEN** 请求不存在的 **`model`** id
- **THEN** 响应状态与 body 与「不存在」语义一致且不含上游密钥信息

### Requirement: OpenAPI 文档更新

嵌入的 OpenAPI 3.0 文档 MUST 在 **`paths`** 中描述 **GET `/v1/models`**（及若实现则 **GET `/v1/models/{model}`**），且 MUST 引用与 Chat 相同的 **`Bearer`** 安全方案。

#### Scenario: 文档含 list models

- **WHEN** 审查者打开 **`openapi.yaml`** 或 Swagger UI
- **THEN** 可见 **GET `/v1/models`** 与鉴权说明

### Requirement: 方法拒绝一致性

对 **`/v1/models`** 的 **POST/PUT/DELETE**（若未实现）SHALL 返回 **405** 与统一 JSON 错误，与 **`/v1/chat/completions`** 的非常用方法策略一致。

#### Scenario: POST 模型列表被拒绝

- **WHEN** 客户端对 **`/v1/models`** 发起 **POST**
- **THEN** **405** 且 body 为网关统一错误格式

## ADDED Requirements（模型请求体改写）

### Requirement: Chat 路径上 model 字段可按模型库绑定改写

对 **POST `/v1/chat/completions`**，在 **`openai-chat-completions-proxy`** 既有「透传」语义下，允许唯一例外：当 **`model-library`** 规范中的绑定要求将 **`model`** 改写为 **`actual_model`** 时，上游收到的 JSON **`model`** 值 SHALL 为改写后的值；其余字段 SHALL 仍按对象级透传（不因改写而丢弃字段）。若 body 非 JSON 对象或解析失败，SHALL 保持字节流原样转发。

#### Scenario: 改写后上游可见

- **WHEN** 绑定要求改写且 body 为合法 JSON 对象
- **THEN** 上游请求体中 **`model`** 与客户端原始值可不同，且其它键保留
