## Why

NexusRouter 作为面向 LLM 调用的 API 网关，需要尽快提供与生态工具链兼容的 **OpenAI Chat Completions** 入口，使客户端可零改造或低改造接入；由网关在边缘完成 **反向代理**、鉴权、可观测与上游隔离。与此同时，需对外提供与 [OpenAI API 概览与约定](https://developers.openai.com/api/reference/overview) 对齐的 **可机读契约（OpenAPI 3.0）**，由 **swaggo/swag** 从代码注解自动生成，并配套 **Swagger UI**，避免文档与实现漂移；**测试驱动** 先固定文档与代理的关键行为，再实现代码。

## What Changes

- 在网关暴露 **标准 OpenAI 兼容路径** `POST /v1/chat/completions`，将请求 **反向代理** 至可配置的上游服务（基址 URL 可配置）。
- **完整透传**：将客户端请求头（在合理安全策略下，见设计）与请求体转发至上游；将上游响应 **状态码、响应头（可配置过滤）、响应体** 原样回传给客户端（流式 `text/event-stream` 与 JSON 均需支持）。
- **内置鉴权**：在命中该路径前执行接口级鉴权（如 API Key / Bearer），未通过则返回统一 JSON 错误且不转发上游。
- **异常与 Panic 捕获**：网络错误、上游超时、非法请求、处理中 panic 等均被捕获，返回 **统一结构的 JSON 错误响应**，并写入 Zap 日志；不向客户端泄露内部栈信息（开发模式可配置详细程度）。
- **OpenAPI 3.0 + swag + Swagger UI**：以 **TDD** 为序增加对 OAS3 暴露、路径与安全声明、UI 入口的测试；实现 **`swag init`** 生成链（必要时 2→3 转换或升级 swag），对外提供 **`/openapi.yaml`**（及可选 **`/openapi.json`**），Swagger UI **加载 OAS3**；非生产默认开启 UI、生产可配置关闭（见设计）。

## Capabilities

### New Capabilities

- `openai-chat-completions-proxy`：**运行时** OpenAI 兼容 `/v1/chat/completions` 反向代理、透传、上游配置、鉴权拦截、统一错误响应；以及 **契约与文档面** OpenAPI 3.0 暴露、swag 生成流程、Swagger UI 与测试驱动验收（同一能力名下统一交付）。

### Modified Capabilities

- （无）与现有 `gateway-backend` 中的通用错误 JSON、Zap、分层目录等要求对齐，不修改其既有条款。

## Impact

- **代码**：`services/gateway` 下 `cmd/api`、`internal/router`、`internal/handler` 或 `internal/proxy`、`internal/config`、`internal/middleware`、Wire 装配；可选 `docs/` 与嵌入 `openapi.yaml`。
- **配置**：上游基址、超时、鉴权、`forward_client_authorization`、**文档 UI 开关** 等。
- **依赖**：`net/http/httputil.ReverseProxy`；`github.com/swaggo/swag`、`github.com/swaggo/gin-swagger`、`github.com/swaggo/files`（或等价）；若 OAS3 经转换，可能增加可复现 CLI（Makefile/Docker/npm）。
- **测试**：`httptest` 双端代理用例；OpenAPI/UI 解析与路由测试；CI 含 `swag init` 或生成物漂移校验。
