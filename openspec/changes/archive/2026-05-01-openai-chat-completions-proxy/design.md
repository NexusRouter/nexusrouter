## Context

网关当前仅提供健康检查等骨架路由（`internal/router`），尚无面向 LLM 的兼容入口。本变更在 **Gin** 上增加 `POST /v1/chat/completions` 的 **反向代理**，将流量转发至可配置的上游，并在网关内完成鉴权与错误统一。同一变更内交付 **OpenAPI 3.0** 文档与 **Swagger UI**，使 swag 注解与真实 handler 共用同一路径与语义。约束：与现有 **Zap**、**JSON 错误**、**Wire**、**Viper 配置** 及中文注释规范对齐。

仓库已在 `go.mod` 中依赖 **swag v1.16.x**（CLI 描述为 **Swagger 2.0**）。用户要求 **OpenAPI 3.0** 与 **swaggo/swag** 自动生成并存，需在实现中择定 **原生 OAS3**（如升级 **swag v2**）或 **Swagger 2 → OpenAPI 3** 的固定转换步骤（见下文决策）。

## Goals / Non-Goals

**Goals（代理）：**

- 使用 Go 标准库 **`net/http/httputil.ReverseProxy`**（或等价 `Director`/`ModifyResponse` 组合）实现反向代理。
- **方法固定为 POST**；路径固定为 `/v1/chat/completions`。
- **透传**：默认转发客户端 `Content-Type`、`Authorization`（鉴权通过后是否剥离见决策）、`Accept`、`User-Agent`（若客户端提供）；请求体 **字节级透传**。
- **响应**：上游状态码与 body 原样返回；**Hop-by-hop** 头按 RFC 7230 由代理层处理。
- **流式**：上游 `text/event-stream` 时禁用整包缓冲、及时 **Flush**。
- **鉴权**：`ReverseProxy` 前中间件；Bearer / API Key；失败 **401** 且不转发。
- **异常**：连接失败、超时、panic 等 → 统一 JSON + **Zap**（脱敏）。

**Goals（文档 / TDD）：**

- **TDD**：先写测试（见 `tasks.md`），覆盖 OAS3 GET、**`openapi: 3.0.*`**、**`paths./v1/chat/completions.post`**、**Bearer** `securitySchemes`、**overview URL**、**Swagger UI** 入口；再实现路由与生成物。代理侧可采用「先测后实现」的同等原则于关键路径。
- **对齐 OpenAI 参考**：`info.description` 或 `externalDocs` 链接 [API Overview](https://developers.openai.com/api/reference/overview)；请求/响应模型与 [Chat Completions](https://developers.openai.com/docs/api-reference/chat) 对齐到 **网关实际支持的子集**（过大时 `description` 标明兼容子集）。
- **Swagger UI**：`gin-swagger` + `files`（或等价），UI **必须加载本服务提供的 OpenAPI 3** URL。
- **生成**：`Makefile` 或 `go generate` 文档化 `swag init`（`-g`、`-d`、`--parseInternal` 等）。

**Non-Goals:**

- 不实现 `/v1/models` 等管理类 API（除非后续变更）。
- 不在网关内做计费、配额限流、模型名改写。
- 不要求解析 SSE 做审计；不承诺与 OpenAI 云端 **字节级** 相同的全量 JSON Schema。

## Decisions

### 代理

1. **`httputil.ReverseProxy`**：成熟、支持 `ModifyResponse` 与可配置 `Transport`。
2. **上游**：单一 **`UPSTREAM_BASE_URL`**（或 `upstream.base_url`），与 `/v1/chat/completions` 解析为绝对 URL。
3. **`Authorization` 转发**：默认 **移除** 客户端 `Authorization` 发往上游，改注网关 **`upstream.api_key`** 等；**`forward_client_authorization: true`** 可覆盖。
4. **超时**：`ResponseHeaderTimeout` + Context 超时（默认如 **120s**）；流式首字节/读空闲见实现迭代。
5. **网关错误 JSON**：与 `gateway-backend` 一致（`message`、`code` 等）；**上游返回的错误体原样透传**，不包一层网关 JSON。

### 文档与 OpenAPI 3.0

6. **OAS3 来源**：优先 **swag 原生输出 OpenAPI 3.x**（如 **swag v2**）；否则 **方案 B**：`swag init` 产 Swagger2 + **`swagger2openapi`** / OpenAPI Generator 固定 target 生成 **`openapi.yaml`（3.0）**。
7. **生成物入库**：提交 **`docs/docs.go`**、中间 JSON/YAML（若存在）及 **`openapi.yaml`（3.0）**，使 `go test` 无 swag CLI 亦可校验；CI 可 `git diff --exit-code` 防漂移。
8. **Swagger UI 路径**：默认 **`/swagger/index.html`**（gin-swagger 惯例）。
9. **UI 门控**：生产默认关或需鉴权；测试 `httptest` 注入为开启。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| Hop-by-hop / `Content-Length` 与 chunked 冲突 | `ReverseProxy` 默认行为；不手动复制 `Transfer-Encoding` |
| Gin `Writer` 与 Flush | `http.Flusher` / `gin.ResponseWriter.Flush` |
| swag v2 不稳定 | 锁定版本或采用方案 B 固定转换器版本 |
| Chat schema 过大拖慢 `swag init` | 独立 DTO 包、精简示例结构体 |
| 大请求体内存 | 首版依赖 streaming；后续 `MaxBytesReader` |

## Migration Plan

1. 部署网关：配置上游与鉴权；文档 UI 按环境门控。
2. 开发者执行一次文档生成 target，提交 OAS3 与 `docs.go`。
3. 灰度客户端 base URL 至网关；回滚指回上游。
4. 运维：内网访问 Swagger UI（若启用）。

## Open Questions

- 上游 **mTLS** 首版是否支持（默认否）。
- 是否在响应中注入 **`X-Request-Id`**（与全局中间件对齐后决定）。
- 是否额外暴露 **`GET /openapi.json`**（推荐 **是**）。
