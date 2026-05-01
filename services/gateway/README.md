# Gateway（Go）

NexusRouter 网关服务，模块路径：`github.com/NexusRouter/nexusrouter/services/gateway`。

## 环境要求

- **Go 1.25.x**（与 `go.mod` 中 `go` 指令一致）
- （可选）**Wire**：`go install github.com/google/wire/cmd/wire@v0.7.0`
- （可选）**Swag**：由 `make docs` 自动 `go run`，无需全局安装
- **Node.js 22+**：执行 `make docs` 时用 `npx swagger2openapi` 将 Swagger 2 转为 **OpenAPI 3**（CI 已配置）
- （可选）**Air 热重载**：`go install github.com/air-verse/air@latest`，在项目根执行 `air`（需自行添加 `.air.toml`）

## 常用命令

```bash
cd services/gateway
go run ./cmd/api
```

默认监听 **:8080**。


| 路径                          | 说明                                                               |
| --------------------------- | ---------------------------------------------------------------- |
| `GET /health`               | 健康检查                                                             |
| `GET /openapi.yaml`         | **OpenAPI 3.0** 规范（YAML）                                         |
| `GET /openapi.json`         | 同上（由嵌入 YAML 转 JSON）                                              |
| `GET /swagger/index.html`   | **Swagger UI**（当 `NEXUSROUTER_ENABLE_SWAGGER_UI` 不为 `false` 时启用） |
| `POST /v1/chat/completions` | OpenAI 兼容 Chat Completions **反向代理**（需网关鉴权与上游配置）                  |


OpenAI 官方 REST 概览与认证约定见：[https://developers.openai.com/api/reference/overview](https://developers.openai.com/api/reference/overview)。

### Chat 代理环境变量


| 变量                                         | 说明                                                               |
| ------------------------------------------ | ---------------------------------------------------------------- |
| `NEXUSROUTER_UPSTREAM_BASE_URL`            | 上游基址（须含 scheme+host，可带 path 前缀）；未设置时 POST 返回 **503**             |
| `NEXUSROUTER_GATEWAY_API_KEYS`             | 逗号分隔；请求须 `Authorization: Bearer <key>` 或 `X-API-Key: <key>` 之一匹配 |
| `NEXUSROUTER_UPSTREAM_API_KEY`             | 发往上游的 Bearer（在未开启透传客户端 Authorization 时注入）                        |
| `NEXUSROUTER_FORWARD_CLIENT_AUTHORIZATION` | `true` 时将客户端 `Authorization` 原样转发上游                              |
| `NEXUSROUTER_UPSTREAM_TIMEOUT`             | 等待上游响应头超时，如 `120s`（默认 120s）                                      |
| `NEXUSROUTER_ENABLE_SWAGGER_UI`            | 设为 `false` 关闭 Swagger UI（OpenAPI 路由仍可用）                          |


### 文档单一事实来源（代码优先）

- **只维护 Go 源码中的 swag 注释**（`cmd/api/main.go`、`internal/handler/swagger_*.go` 等）。
- `make docs` 依次：`swag init` → 生成 `docs/`（Swagger 2）→ `swagger2openapi` 生成 `**internal/openapi/openapi.yaml`（OAS3）**。
- 运行时将 **OAS3 文件嵌入二进制**（`go:embed`），`/openapi.yaml` 与 `/openapi.json` 均来自该生成物；**请勿手改** `internal/openapi/openapi.yaml`。

```bash
cd services/gateway
make docs
```

CI 会校验 `**docs/**` 与 `**internal/openapi/openapi.yaml**` 与仓库一致（`git diff --exit-code`）。

### 重新生成 Wire

```bash
cd services/gateway
make wire
```

或：

```bash
cd services/gateway/cmd/api
go run github.com/google/wire/cmd/wire@v0.7.0
```

