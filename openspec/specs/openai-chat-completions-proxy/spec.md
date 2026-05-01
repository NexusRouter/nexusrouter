# openai-chat-completions-proxy Specification

## Purpose

NexusRouter 网关对 **OpenAI 兼容 Chat Completions** 的反向代理、鉴权与错误处理，以及 **OpenAPI 3.0** 契约暴露、**swaggo/swag** 生成物与 **Swagger UI** 的交付要求（归档自 change `openai-chat-completions-proxy`）。

## Requirements

### Requirement: OpenAI 兼容路径与 HTTP 方法

网关 MUST 注册 **POST `/v1/chat/completions`**，且仅对该路径提供本规范所述的反向代理行为（与其他路由并存）。

#### Scenario: 仅 POST 命中代理

- **WHEN** 客户端对 `/v1/chat/completions` 发起 **POST** 且通过鉴权（见下文）
- **THEN** 请求被转发至配置的上游目标

#### Scenario: 非 POST 拒绝

- **WHEN** 客户端对 `/v1/chat/completions` 使用 **GET/PUT/DELETE** 等非常用方法（除 **OPTIONS** 若启用 CORS 预检外）
- **THEN** 响应状态码为 **405** 或 **404**（实现择一但 MUST 在 `design.md`/代码注释中固定），且 body 为网关统一 JSON 错误（非上游体）

### Requirement: 上游目标可配置

网关 MUST 通过配置（Viper 键或环境变量，具体键名以实现为准但 MUST 文档化）指定上游 **基址 URL**（含 scheme 与 host，可选 path 前缀）；对 **POST `/v1/chat/completions`** 的转发目标 MUST 等价于「该基址与 OpenAI 标准路径解析后的绝对 URL」。

#### Scenario: 配置缺失时拒绝启动或拒绝转发

- **WHEN** 进程启动时上游基址未设置或非法
- **THEN** MUST 启动失败 **或** 对该路径返回 **503** 统一 JSON（二者择一并在实现中一致）；禁止向空 host 发起转发

#### Scenario: 合法配置下转发

- **WHEN** 上游基址为 `https://api.example.com` 且客户端请求 **POST `/v1/chat/completions`**
- **THEN** 上游收到的请求 URL MUST 指向 `https://api.example.com/v1/chat/completions`（若基址带 path 前缀，MUST 与 RFC 3986 路径合并规则一致）

### Requirement: 请求头与请求体透传

在通过鉴权后，网关 MUST 将客户端请求体 **原样字节流** 转发至上游（不解析、不修改 JSON 结构）。网关 MUST 转发与 Chat Completions 相关的常见语义头，至少包括：**`Content-Type`**、**`Accept`**、**`User-Agent`**（若客户端提供）；其他头 MAY 按 `design.md` 中的 hop-by-hop 剔除规则处理。

#### Scenario: JSON 体原样到达上游

- **WHEN** 客户端发送带合法 JSON 的 `Content-Type: application/json` 的 body
- **THEN** 上游读取的 body 字节序列与客户端一致

#### Scenario: 流式请求体声明透传

- **WHEN** 客户端 `Accept: text/event-stream` 且 body 仍为合法 chat completions JSON
- **THEN** 上游收到的 `Accept` 与 body 与客户端一致（除非 `design.md` 声明剥离特定安全头）

### Requirement: 响应原样回包

当上游返回响应时，网关 MUST 将上游 **HTTP 状态码** 与 **响应体** 回传至客户端。上游响应头 MUST 在剔除 hop-by-hop 及由代理层重算的头之后尽可能保留（至少保留 **`Content-Type`**；流式时保留 **`Cache-Control`** / **`text/event-stream`** 语义所需头）。

#### Scenario: 上游 JSON 成功响应

- **WHEN** 上游返回 **200** 与 `application/json` body
- **THEN** 客户端收到的状态码为 **200** 且 body 字节序列与上游一致

#### Scenario: 上游 SSE 流式响应

- **WHEN** 上游返回 **200** 且 `Content-Type: text/event-stream` 并逐步写入 chunk
- **THEN** 客户端收到的状态码与 content-type 与上游一致，且数据以流式可达（不要求单 chunk 对齐，但 MUST 不在网关内整包缓冲完整流）

#### Scenario: 上游业务错误原样

- **WHEN** 上游返回 **4xx/5xx** 及自有 JSON 错误体
- **THEN** 客户端收到相同状态码与相同 body（hop-by-hop 头除外）

### Requirement: 内置接口鉴权拦截

在转发前，网关 MUST 对 **POST `/v1/chat/completions`** 执行鉴权：至少支持 **`Authorization: Bearer <token>`** 与/或 **`X-API-Key`**（具体支持的头集合以实现为准，MUST 文档化），并与配置中的允许值比对；未通过 MUST **不调用上游**。

#### Scenario: 缺少凭证

- **WHEN** 请求未携带配置的鉴权凭证
- **THEN** 响应为 **401** 且统一 JSON 错误，且上游无对应请求

#### Scenario: 凭证错误

- **WHEN** 凭证与配置不一致
- **THEN** 响应为 **401** 且统一 JSON 错误

#### Scenario: 凭证正确

- **WHEN** 凭证正确
- **THEN** 请求进入反向代理转发流程

### Requirement: 异常捕获与统一错误响应

网关 MUST 捕获：上游连接失败、TLS 错误、超时、`ReverseProxy` 内部错误、以及处理链中的 **panic**，并返回 **JSON** 错误体（结构符合 `gateway-backend` 统一错误约定：含人类可读 `message` 与机器可读 `code` 或等价字段）；同时 MUST 写入 **Zap** 日志（含错误类型，不含完整客户端密钥）。

#### Scenario: 上游不可达

- **WHEN** 上游 TCP/HTTP 连接失败
- **THEN** 客户端收到 **502**（或 **503**，实现固定一种）与统一 JSON，且 Zap 有错误记录

#### Scenario: 超时

- **WHEN** 上游在配置的超时时间内未返回响应头（或非流式未结束）
- **THEN** 客户端收到 **504** 与统一 JSON

#### Scenario: Panic 不崩溃进程

- **WHEN** 转发路径发生 panic
- **THEN** 进程不退出，客户端收到 **500** 与统一 JSON（无栈信息），且 Zap 记录 panic 摘要

### Requirement: OpenAPI 3.0 规范可机读暴露

网关 MUST 通过 **HTTP GET** 提供至少一份 **OpenAPI 3.0** 文档（`application/json` 或 `application/yaml`），且文档根字段 **`openapi`** MUST 以 **`3.0.`** 为前缀（例如 `3.0.3`）。

#### Scenario: 获取 OpenAPI JSON 或 YAML

- **WHEN** 客户端请求文档约定路径（如 **GET `/openapi.yaml`** 或 **GET `/openapi.json`**，以实现注册为准）
- **THEN** 响应状态码为 **200**，`Content-Type` 与正文格式一致，且正文可被标准 OpenAPI 3.0 解析器解析

#### Scenario: 版本字段合法

- **WHEN** 解析该文档根对象
- **THEN** 存在 **`openapi`** 字符串且以 **`3.0.`** 为前缀（OpenAPI 3.0.x）

### Requirement: OpenAPI 中 Chat Completions 路径与动词

文档 MUST 在 **`paths`** 下描述 **`/v1/chat/completions`** 的 **`post`** 操作，且 **`operationId`** 或 **`summary`** 之一 MUST 明示「Chat Completions」语义（英文或中英文均可识别）。

#### Scenario: 路径存在

- **WHEN** 审查者检索文档 `paths["/v1/chat/completions"]`
- **THEN** 存在 **`post`** 对象且包含 **请求体**（`requestBody`）与 **至少一个 2xx 响应** 声明

### Requirement: OpenAPI 与 OpenAI 概览对齐的认证说明

文档 MUST 声明与 [OpenAI API Overview — Authentication](https://developers.openai.com/api/reference/overview) 一致的 **HTTP Bearer** 认证方式：在 **`components.securitySchemes`** 中定义 **`type: http`、`scheme: bearer`**（或 OAS3 等价写法），且 **`/v1/chat/completions` POST** MUST 引用该安全要求（全局 `security` 或操作级 `security` 均可）。

#### Scenario: Bearer 方案存在且被操作引用

- **WHEN** 解析 `components.securitySchemes` 与 `paths./v1/chat/completions.post.security`（或根级 `security`）
- **THEN** 客户端可识别需携带 **`Authorization: Bearer`** 凭证

### Requirement: OpenAPI 对外参考链接

文档 MUST 在 **`info.description`** 或 **`externalDocs`** 中包含指向 **`https://developers.openai.com/api/reference/overview`** 的 URL，以便读者对照官方概览。

#### Scenario: 链接可访问性（静态检查）

- **WHEN** 对文档文本执行字符串检查（测试或 CI）
- **THEN** 包含上述 **https** 完整 URL 子串

### Requirement: swaggo/swag 自动生成

项目 MUST 使用 **`github.com/swaggo/swag`** 从 Go 源码注释生成文档源码与中间产物（`docs.go` 及至少一种 JSON/YAML）；生成命令 MUST 在 **`services/gateway`** 的 README 或 Makefile/`go generate` 中可发现。

#### Scenario: 可复现生成

- **WHEN** 开发者在网关模块根执行文档化生成命令（如 `make docs` 或 `go generate ./...`）
- **THEN** 以零退出码完成且产出文件集与仓库策略一致（见 `design.md`）

### Requirement: Swagger UI

当配置允许（见 `design.md`）时，网关 MUST 提供 **Swagger UI**，且 UI MUST 加载本服务提供的 **OpenAPI 3.0** 文档 URL（而非错误地固定仅支持已废弃的 Swagger 2 且与仓库 OAS3 不一致的地址）。

#### Scenario: UI 可访问

- **WHEN** 配置为开启文档 UI 且请求 **GET `/swagger/index.html`**（或实现注册的等价入口）
- **THEN** 响应 **200**，且 HTML 中引用之 spec URL 指向上述 OAS3 文档路径

### Requirement: 测试驱动验收（文档与契约）

针对「OpenAPI 3.0 暴露」「OpenAPI 中 Chat 路径存在」「Bearer 安全」「Swagger UI 入口」的自动化测试 MUST **先于** 使上述行为通过的生产代码合并（同一 PR 或严格有序提交：测试提交可红，紧随实现提交转绿）；合并到主分支时 **MUST 全绿**。

#### Scenario: CI 执行文档相关测试

- **WHEN** CI 运行 `go test`（含 `-count=1` 若项目约定）
- **THEN** 包含对 OpenAPI 与 UI 的测试包且通过
