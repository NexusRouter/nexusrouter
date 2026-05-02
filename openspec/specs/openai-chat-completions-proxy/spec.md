# openai-chat-completions-proxy Specification

## Purpose

NexusRouter 网关对 **OpenAI 兼容 Chat Completions** 的反向代理、鉴权与错误处理（归档自 change `openai-chat-completions-proxy`）。

## 实现说明（当前代码库）

- 网关 **不**通过 HTTP 提供机读 OpenAPI / Swagger：**无** **`GET /openapi.json`** / **`GET /openapi.yaml`**、**无** Swagger UI、**无** swag 生成物或仓库内嵌 OpenAPI 文档包。
- 产品 **不**要求交付或维护 **OpenAPI 3.0** 规范文档：无义务在仓库提交 **`openapi.yaml`/`openapi.json`**、无 **`swag init`** 等生成链路作为交付物。
- 下文凡涉及 **OpenAPI 机读暴露**、**嵌入文档**、**swag**、**Swagger UI**、**OpenAPI 3.0 文档** 的 **MUST/Scenario**，均视为历史归档；与上段冲突时以上段为准。

## Requirements
### Requirement: OpenAI 兼容路径与 HTTP 方法

网关 MUST 注册 **POST `/v1/chat/completions`**，且仅对该路径提供本规范所述的反向代理行为（与其他路由并存）。

#### Scenario: 仅 POST 命中代理

- **WHEN** 客户端对 `/v1/chat/completions` 发起 **POST** 且通过鉴权（见下文）
- **THEN** 请求被转发至配置的上游目标

#### Scenario: 非 POST 拒绝

- **WHEN** 客户端对 `/v1/chat/completions` 使用 **GET/PUT/DELETE** 等非常用方法（除 **OPTIONS** 若启用 CORS 预检外）
- **THEN** 响应状态码为 **405**，且 body 为网关统一 JSON 错误（非上游体）

### Requirement: 上游目标可配置

网关 MUST 通过配置支持 **一个或多个** 上游 **基址 URL**（每项含 scheme 与 host，可选 path 前缀）；对 **POST `/v1/chat/completions`** 的转发目标 MUST 为：按 **`design.md`** 固定的选择策略从列表中选出某一基址后，与 OpenAI 标准路径 **`/v1/chat/completions`** 按 **RFC 3986** 合并得到的绝对 URL。配置 MAY 同时保留单一环境变量形式以兼容旧部署；**优先级与键名 MUST 在 `design.md` 与 README 文档化**。

#### Scenario: 多上游轮询

- **WHEN** 配置中至少存在 **两个** 合法上游基址且策略为轮询
- **THEN** 连续多次成功的 **`POST /v1/chat/completions`** 请求所命中的上游主机（不含 path 细节时至少 host）MUST 在统计意义上轮换（测试可用固定种子或钩子验证索引递增语义）

#### Scenario: 配置缺失时拒绝启动或拒绝转发

- **WHEN** 进程启动时 **所有** 上游基址均未设置或均非法
- **THEN** MUST 启动失败 **或** 对该路径返回 **503** 统一 JSON（二者择一并在实现中一致）；禁止向空 host 发起转发

#### Scenario: 合法配置下转发

- **WHEN** 所选上游基址为 `https://api.example.com` 且客户端请求 **POST `/v1/chat/completions`**
- **THEN** 上游收到的请求 URL MUST 指向 `https://api.example.com/v1/chat/completions`（若基址带 path 前缀，MUST 与 RFC 3986 路径合并规则一致）

### Requirement: 上游出站 HTTP 客户端代理

网关 MUST 支持通过**环境变量**（键名 **`NEXUSROUTER_UPSTREAM_HTTP_PROXY`**，值为绝对代理 URL，首尾空白裁剪）为 **Chat 反向代理** 使用的 **`http.Transport`** 显式设置 **`Proxy`**（如 **`http://host:port`**）。当该变量**未设置**或值为**空**时，MUST NOT 覆盖 **`Transport.Proxy`** 的默认语义（即沿用 **`net/http`** 对 **`HTTP_PROXY`/`HTTPS_PROXY`/`NO_PROXY`** 等环境变量的默认行为）。当值为**非空**但无法解析为含 **scheme** 与 **host** 的合法 URL 时，MUST 忽略该值（等价于未设置显式代理），且 MUST NOT 因解析失败而中断进程启动。

#### Scenario: 设置显式代理 URL

- **WHEN** 环境变量 **`NEXUSROUTER_UPSTREAM_HTTP_PROXY`** 为合法的 **`http://`** 或 **`https://`** 代理地址
- **THEN** 发往 Chat 上游的客户端连接 MUST 经该 **`Proxy`** 函数解析出的代理发起（与 **`net/http.ProxyURL`** 语义一致）

#### Scenario: 未设置时保留环境代理默认

- **WHEN** **`NEXUSROUTER_UPSTREAM_HTTP_PROXY`** 未设置或为空
- **THEN** Chat 上游 **`http.Transport`** 的 **`Proxy`** 字段 MUST 为 **`nil`**，以保留标准库默认代理解析（含仅通过 **`HTTP_PROXY`** 等配置的场景）

### Requirement: 请求头与请求体透传

在通过鉴权后，网关 MUST 将客户端请求体转发至上游；**默认**在语义上与客户端 body 一致，**但**允许下文 **「流式缺省 Accept 补充」**、**「流式 stream_options 合并」**、以及 **`design.md`** / 实现中明示的 **model 改写**（模型库/绑定路径）等对 JSON 或头的**有文档出站改写**。网关 MUST **复制**客户端请求头至上游请求，**但** MUST 剔除 **`design.md`** 所列 **hop-by-hop** 头及代理层重算头；**`Host`** MUST 设置为目标上游主机；且 MUST 在出站至上游前移除 **`Accept-Encoding`**（无论客户端是否声明），以免将客户端对**响应体**压缩算法的协商带给上游。除上述清单与 **`Accept-Encoding`** 外 **不得** 无故丢弃业务语义头（例如 **`OpenAI-Organization`**、**`Idempotency-Key`**、**`X-Request-ID`** 等，若客户端提供则保留，除非与安全策略冲突并在 `design.md` 明示）。

#### Scenario: JSON 体原样到达上游

- **WHEN** 客户端发送带合法 JSON 的 `Content-Type: application/json` 的 body，且**未触发**本节上文允许的出站改写（例如非流式、未启用合并、无 model 改写）
- **THEN** 上游读取的 body 字节序列与客户端一致

#### Scenario: Gzip 压缩体在上游表现为明文 JSON

- **WHEN** 客户端对 **`POST /v1/chat/completions`** 使用 **`Content-Encoding: gzip`** 发送 gzip 压缩后的合法 Chat Completions JSON，且通过鉴权
- **THEN** 上游收到的 body 字节序列与**解压后**的 JSON 一致（与客户端以未压缩方式发送同一 JSON 时一致），且上游请求 MUST NOT 仍携带表示 gzip 载荷的 **`Content-Encoding`**（与解压后 body 一致）

#### Scenario: 流式请求体声明透传

- **WHEN** 客户端 `Accept: text/event-stream` 且 body 仍为合法 chat completions JSON
- **THEN** 上游收到的 `Accept` 与 body 与客户端一致（hop-by-hop 及 `design.md` 声明的安全例外除外）

#### Scenario: 自定义业务头保留

- **WHEN** 客户端在剔除名单之外携带非常见但合法的业务请求头 **`X-Custom-Client: foo`**
- **THEN** 上游请求 MUST 包含 **`X-Custom-Client: foo`**（大小写按 Go `http.Header` 规范）

#### Scenario: 出站不转发 Accept-Encoding

- **WHEN** 客户端在 **`POST /v1/chat/completions`** 上设置非空 **`Accept-Encoding`**（例如常见浏览器的 **`gzip, deflate, br`**）
- **THEN** 上游收到的 HTTP 请求 MUST NOT 含 **`Accept-Encoding`** 头（按大小写不敏感判断）

### Requirement: Chat Completions 请求体 max_tokens 校验

当请求体为可解析的 JSON **对象**且顶层含 **`max_tokens`** 字段（且值不为 JSON **`null`**）时，网关 MUST 在校验通过鉴权之后、向上游转发之前校验该字段：值 MUST 为 JSON **数字**且为**非负整数**，且 MUST NOT 大于 **`MaxInt32/2`**（与常见 OpenAI 兼容实现对非法 `max_tokens` 的拒绝语义一致）。不满足时 MUST 响应 **400**，body 为网关统一 JSON，**`code`** 为 **`INVALID_REQUEST`**，且 MUST NOT 向上游发起该次转发。

#### Scenario: 合法 max_tokens 可转发

- **WHEN** body 为对象且 **`max_tokens`** 为介于 **0** 与 **`MaxInt32/2`**（含）之间的整数
- **THEN** 不因此项校验单独拒绝请求（仍受其它规则约束）

#### Scenario: 非法 max_tokens 拒绝且不命中上游

- **WHEN** **`max_tokens`** 为负数、大于 **`MaxInt32/2`**、非整数数字、或非数字类型（如字符串、布尔）
- **THEN** 响应为 **400** 统一 JSON，**`code`** 为 **`INVALID_REQUEST`**，且上游 MUST NOT 收到该请求

### Requirement: Chat Completions 请求体 messages 必填与非空

当请求体为可解析的 JSON **对象**时，网关 MUST 在校验通过鉴权之后、向上游转发之前校验顶层 **`messages`**：字段 MUST 存在；值 MUST 为 JSON **数组**且长度 **≥ 1**（**`null`**、非数组类型、或空数组均不满足）。不满足时 MUST 响应 **400**，body 为网关统一 JSON，**`code`** 为 **`INVALID_REQUEST`**，且 MUST NOT 向上游发起该次转发。当请求体无法解为 JSON 对象时，MUST NOT 仅因本规则拒绝（与其它校验及上游行为一致）。

#### Scenario: 含至少一条消息的 messages 不因本规则单独拒绝

- **WHEN** body 为对象且 **`messages`** 为非空数组（元素可为任意合法 JSON 值）
- **THEN** 不因此项校验单独拒绝请求（仍受其它规则约束）

#### Scenario: 缺省、null、空数组或非数组 messages 拒绝且不命中上游

- **WHEN** 对象为 **`{}`**、缺少 **`messages`**、**`messages`** 为 **`null`**、**`messages`** 为 **`[]`**、或 **`messages`** 非数组
- **THEN** 响应为 **400** 统一 JSON，**`code`** 为 **`INVALID_REQUEST`**，且上游 MUST NOT 收到该请求

### Requirement: 流式请求缺省 Accept 补充

对 **`POST /v1/chat/completions`**，当请求体 JSON **顶层**字段 **`stream`** 为布尔 **`true`**，且入站请求的 **`Accept`** 头缺失或仅含首尾空白时，网关 MUST 在发往上游的出站请求中设置 **`Accept: text/event-stream`**。当 **`Accept`** 含非空白值时，MUST NOT 修改该 **`Accept`**。当 **`stream`** 不为布尔 **`true`**（缺省、**`false`** 或不可解析为布尔 **`true`**）时，MUST NOT 为上述目的增设或改写 **`Accept`**。

#### Scenario: stream 为 true 且无 Accept 时向上游补充

- **WHEN** 请求体为合法 JSON 且含 **`"stream":true`**，且入站请求未设置非空 **`Accept`**
- **THEN** 上游收到的请求 MUST 含 **`Accept: text/event-stream`**

#### Scenario: 客户端已声明 Accept 时不覆盖

- **WHEN** 请求体含 **`"stream":true`**，且客户端 **`Accept`** 为非空白值（例如 **`application/json`**）
- **THEN** 上游收到的 **`Accept`** 与客户端一致

#### Scenario: 非流式不补充 Accept

- **WHEN** 请求体无 **`stream`** 或为 **`false`**，且客户端未设置 **`Accept`**
- **THEN** 上游请求 MUST NOT 因本规则单独出现 **`Accept: text/event-stream`**（与入站一致或保持缺省）

### Requirement: 流式 stream_options.include_usage 合并

当环境变量 **`NEXUSROUTER_CHAT_STREAM_INCLUDE_USAGE`** **未设置**时，网关 MUST 视为 **`true`**（与实现中「未显式关闭则启用」一致）。当其为 **`false`** 时，MUST NOT 为本规则修改请求 JSON。

当上述有效值为 **`true`**，且请求体为合法 JSON 对象、顶层 **`stream`** 为布尔 **`true`**，且 **`stream_options`** 缺失或为 JSON 对象时，网关 MUST 在发往上游的 body 中合并或写入 **`stream_options`**，并设置 **`include_usage`** 为 **`true`**，且 MUST 保留 **`stream_options`** 内已有其它键（**`include_usage`** 以 **`true`** 为准覆盖同名字段）。当 **`stream_options`** 已存在且**不是** JSON 对象时，MUST NOT 修改 body（避免破坏非标准载荷）。

#### Scenario: 流式且启用合并时上游含 include_usage

- **WHEN** **`NEXUSROUTER_CHAT_STREAM_INCLUDE_USAGE`** 未设置或为 **`true`**，且 body 含 **`"stream":true`** 且 **`stream_options`** 缺失或可被合并
- **THEN** 上游收到的 JSON MUST 在 **`stream_options`** 内含 **`"include_usage":true`**

#### Scenario: 显式关闭合并不改写 body

- **WHEN** **`NEXUSROUTER_CHAT_STREAM_INCLUDE_USAGE`** 为 **`false`**，且 body 含 **`"stream":true`**
- **THEN** 上游收到的 body 中 **`stream_options`** 与客户端一致（未因本规则新增或覆盖 **`include_usage`**）

### Requirement: 响应原样回包

当上游返回响应时，网关 MUST 将上游 **HTTP 状态码** 与 **响应体** 回传至客户端。上游响应头 MUST 在剔除 hop-by-hop 及由代理层重算的头之后尽可能保留（至少保留 **`Content-Type`**；流式时保留 **`Cache-Control`** / **`text/event-stream`** 语义所需头）。

#### Scenario: 上游 JSON 成功响应

- **WHEN** 上游返回 **200** 与 `application/json` body
- **THEN** 客户端收到的状态码为 **200** 且 body 字节序列与上游一致

#### Scenario: 上游 SSE 流式响应

- **WHEN** 上游返回 **200** 且 `Content-Type: text/event-stream` 并逐步写入 chunk
- **THEN** 客户端收到的状态码与 content-type 与上游一致，且数据以流式可达（不要求单 chunk 对齐，但 MUST 不在网关内整包缓冲完整流）；且响应头 MUST 包含 **`X-Accel-Buffering: no`**（网关对 SSE 回写客户端时附加，以降低常见反向代理对响应体的缓冲）；且当上游 **`Cache-Control`** 缺省或仅空白时，网关 MUST 在回写前补充 **`Cache-Control: no-cache`**，若上游已提供非空 **`Cache-Control`** 则 MUST NOT 覆盖；且在按 hop-by-hop 规则剔除上游 **`Connection`** 等头之后，若即将回写至客户端的响应中 **`Connection`** 仍缺省或仅空白，网关 MUST 补充 **`Connection: keep-alive`**

#### Scenario: 上游业务错误原样

- **WHEN** 上游返回 **4xx/5xx** 及自有 JSON 错误体
- **THEN** 客户端收到相同状态码与相同 body（hop-by-hop 头除外）

### Requirement: 内置接口鉴权拦截

在转发前，网关 MUST 对 **POST `/v1/chat/completions`** 执行鉴权，且鉴权行为 MUST 满足 **`api-key-management`** 规范：**MUST** 支持并优先记录 **`Authorization: Bearer <API_KEY>`** 为网关入口凭证；**`X-API-Key`** MAY 继续支持以实现向后兼容，若支持 MUST 在 README 明示为可选。未通过鉴权 MUST **不调用上游**。

#### Scenario: 缺少凭证

- **WHEN** 请求未携带满足 Bearer 形式的网关凭证（且未命中可选的 `X-API-Key` 兼容路径，若启用）
- **THEN** 响应为 **401** 且统一 JSON 错误（含 **`request_id`**），且上游无对应请求

#### Scenario: 凭证错误

- **WHEN** 凭证与已加载密钥集合不一致或记录已禁用/过期
- **THEN** 响应为 **401** 且统一 JSON 错误，且上游无对应请求

#### Scenario: 凭证正确

- **WHEN** 凭证正确且未过期且未禁用
- **THEN** 请求进入反向代理转发流程

### Requirement: 异常捕获与统一错误响应

网关 MUST 捕获：上游连接失败、TLS 错误、超时、`ReverseProxy` 内部错误、以及处理链中的 **panic**，并返回 **JSON** 错误体；该 JSON MUST 符合 **`gateway-backend`** 对网关错误体的字段约定（至少含 **`code`**、**`message`**、**`request_id`**，且 **`request_id`** 与 **`X-Request-ID`** 响应头一致）；同时 MUST 写入 **Zap** 日志（含错误类型与 **`request_id`**，不含完整客户端密钥）。对 **`POST /v1/chat/completions`** 转发路径内由本路径 **`recover`** 消费的 **panic**，该条 **Zap** 日志 MUST 另含 **`stack`** 字段（字符串，为捕获时刻的 **Go** 调用栈文本，与引擎级 **`ZapRecovery`** 的栈记录语义一致），且 MUST NOT 将栈或请求体写入回写客户端的 **JSON**。

#### Scenario: 上游不可达

- **WHEN** 上游 TCP/HTTP 连接失败
- **THEN** 客户端收到 **502**（或 **503**，实现固定一种）与统一 JSON，且 Zap 有错误记录

#### Scenario: 超时

- **WHEN** 上游在配置的超时时间内未返回响应头（或非流式未结束）
- **THEN** 客户端收到 **504** 与统一 JSON

#### Scenario: Panic 不崩溃进程

- **WHEN** 转发路径发生 panic
- **THEN** 进程不退出，客户端收到 **500** 与统一 JSON（无栈信息），且 Zap 记录 panic 摘要（含 **`stack`** 字段，且 **`stack`** 非空）

### Requirement: 无 OpenAPI 3.0 文档与无网关机读端点

网关 MUST NOT 注册 **`GET /openapi.json`**、**`GET /openapi.yaml`** 或 **Swagger UI**；MUST NOT 将 **OpenAPI 3.0** 生成物或内嵌文档包作为必选交付物。下文历史条款中凡要求提供 **OpenAPI 3.0** 文档、嵌入 OpenAPI、swag 生成或文档 UI 的 **MUST**，均以文首 **实现说明** 为准，**不适用**当前实现。

#### Scenario: OpenAPI 路径未注册

- **WHEN** 客户端请求 **GET `/openapi.json`** 或 **GET `/openapi.yaml`**
- **THEN** 响应为 **404**（或未匹配路由之等价行为）

### Requirement: OpenAI 兼容模型列表端点

网关 SHALL 注册 **GET `/v1/models`**，且 MUST 在与 **POST `/v1/chat/completions`** 相同的网关入口鉴权链下执行（**Bearer** 优先，可选 **`X-API-Key`** 若项目仍声明兼容）。成功响应 MUST 为 JSON，且 **`object`** 字段为 **`list`**，**`data`** 为数组；**`data`** MUST 仅包含存在 **至少一条 `model_instance`（`status=1`）且关联 `model_base`、`model_vendor`、`model_upstream` 均为启用（`status=1`）** 的逻辑模型；数组元素 MUST 至少包含 **`id`**（等于 **`model_base.model_code`**）、**`object`**（值为 **`model`**）、**`created`**（Unix 秒，整数）、**`owned_by`**（字符串，可为 **`model_vendor.vendor_name`** 或占位）。每项 SHALL 另含非空 **`permission`** 数组（元素为 **`model_permission`** 对象，字段集合与常见 OpenAI 列表形状一致）及 **`root`** 字符串（等于该项 **`id`**）；无父模型时 **`parent`** 宜省略。

#### Scenario: 鉴权失败不列模型

- **WHEN** 请求未携带有效网关凭证
- **THEN** 响应 **401** 且**不**返回模型列表 body

#### Scenario: 形状可被客户端解析

- **WHEN** 客户端解析成功响应 JSON
- **THEN** 可读取 **`data`** 数组且每项含 **`id`**

### Requirement: 可选模型检索端点

网关 MAY 注册 **GET `/v1/models/:model`**；若模型 id 不可用，SHALL 返回 **404** 或 OpenAI 常见错误包装（与实现选定一致，且在全仓库唯一）。成功响应为单模型对象时，SHALL 与列表项使用相同的 **`permission`**、**`root`**（及可省略的 **`parent`**）约定。

#### Scenario: 未知模型

- **WHEN** 请求不存在的 **`model`** id
- **THEN** 响应状态与 body 与「不存在」语义一致且不含上游密钥信息

### Requirement: 方法拒绝一致性

对 **`/v1/models`** 的 **POST/PUT/DELETE**（若未实现）SHALL 返回 **405** 与统一 JSON 错误，与 **`/v1/chat/completions`** 的非常用方法策略一致。

#### Scenario: POST 模型列表被拒绝

- **WHEN** 客户端对 **`/v1/models`** 发起 **POST**
- **THEN** **405** 且 body 为网关统一错误格式

### Requirement: 模型路径 OPTIONS

网关 MUST 注册 **OPTIONS `/v1/models`** 与 **OPTIONS `/v1/models/:model`**，且 MUST **不**要求入口鉴权；响应 MUST 为 **204 No Content** 且无 JSON body，与 **`OPTIONS /v1/chat/completions`** 一致，以便在未命中引擎级 CORS 处理（例如无 **`Origin`**）时仍可对模型列表路径返回成功状态。

#### Scenario: OPTIONS 列表路径

- **WHEN** 客户端对 **`/v1/models`** 发起 **OPTIONS**
- **THEN** 响应 **204** 且无 body

#### Scenario: OPTIONS 单模型路径

- **WHEN** 客户端对 **`/v1/models/{id}`** 发起 **OPTIONS**
- **THEN** 响应 **204** 且无 body

