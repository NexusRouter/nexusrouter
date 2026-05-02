## MODIFIED Requirements

### Requirement: OpenAI 兼容模型列表端点

网关 SHALL 注册 **GET `/v1/models`**，且 MUST 在与 **POST `/v1/chat/completions`** 相同的网关入口鉴权链下执行（**Bearer** 优先，可选 **`X-API-Key`** 若项目仍声明兼容）。成功响应 MUST 为 JSON，且 **`object`** 字段为 **`list`**，**`data`** 为数组；**`data`** MUST 仅包含存在 **至少一条 `model_instance`（`status=1`）且关联 `model_base`、`model_vendor`、`model_upstream` 均为启用（`status=1`）** 的逻辑模型；数组元素 MUST 至少包含 **`id`**（等于 **`model_base.model_code`**）、**`object`**（值为 **`model`**）、**`created`**（Unix 秒，整数）、**`owned_by`**（字符串，可为 **`model_vendor.vendor_name`** 或占位）。

#### Scenario: 鉴权失败不列模型

- **WHEN** 请求未携带有效网关凭证
- **THEN** 响应 **401** 且**不**返回模型列表 body

#### Scenario: 形状可被客户端解析

- **WHEN** 客户端解析成功响应 JSON
- **THEN** 可读取 **`data`** 数组且每项含 **`id`**
