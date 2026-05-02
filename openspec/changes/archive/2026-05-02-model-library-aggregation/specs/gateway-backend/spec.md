## ADDED Requirements

### Requirement: Chat 路径逻辑模型解析与模型实例选择

在 **POST `/v1/chat/completions`** 处理链中，当启用模型库聚合时，网关 SHALL 从 JSON body 读取 **`model`**，解析 **`model_base.model_code`**，在 **`model_instance`** 上按 **`model-library-aggregation`** 选择策略（各表 **`status=1`**、**`priority`、`weight`、`is_official`**、可选健康）选中 **一条实例**，使用其关联 **`model_upstream`** 的 **`base_url`** 作为转发根、**`api_key`** 作为上游鉴权、**`timeout`/`max_concurrent`** 参与客户端/队列策略（以实现为准）。**本路径 MUST NOT 再使用 `gateway.yaml` 中声明的 Upstream 列表参与同一请求的寻址或回退**（与静态配置**不并存**，单一事实来源为四表）。若无可用实例，SHALL 返回与 **`gateway-backend`** 错误约定一致的响应，且 MUST NOT 泄露内部表名或 **`api_key`**。

#### Scenario: 选择结果可追踪

- **WHEN** 请求成功转发
- **THEN** 结构化日志 MAY 记录 **`model_instance.id`**，且 MUST NOT 记录 **`api_key`**

#### Scenario: 无可用实例

- **WHEN** 逻辑模型存在但无 **`status=1`** 的可用链
- **THEN** 响应为 **4xx**（实现选定）且 **`code`** 稳定可枚举

### Requirement: 模型库管理路由覆盖聚合资源

`services/gateway` SHALL 在 **`/api/admin/v1/model-library/**`**（或文档化之等价前缀）下注册足以支撑 **`model_vendor`、`model_base`、`model_upstream`、`model_instance`** 的管理端点；路由 MUST 复用现有 **`adminJWTMiddleware`** 与写保护策略，且错误体符合既有 JSON 约定。

#### Scenario: 路由可发现

- **WHEN** 审查者检索路由注册表或 OpenAPI
- **THEN** 可见厂商/逻辑模型/上游/实例相关路径与 HTTP 方法映射
