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

版本号可通过构建注入，例如：

```bash
go run -ldflags "-X github.com/NexusRouter/nexusrouter/services/gateway/internal/buildinfo.Version=1.2.3" ./cmd/api
```

未注入时 `**GET /health**` 的 `**version**` 字段为 `**dev**`。


| 路径                              | 说明                                                                                                  |
| ------------------------------- | --------------------------------------------------------------------------------------------------- |
| `GET /health`                   | 健康检查（**无需鉴权**），返回 `status`、`version`、`server_time`（UTC RFC3339Nano）                                 |
| `GET /openapi.yaml`             | **OpenAPI 3.0** 规范（YAML）                                                                            |
| `GET /openapi.json`             | 同上（由嵌入 YAML 转 JSON）                                                                                 |
| `GET /swagger/index.html`       | **Swagger UI**（当 `NEXUSROUTER_ENABLE_SWAGGER_UI` 不为 `false` 时启用）                                    |
| `POST /v1/chat/completions`     | OpenAI 兼容 Chat Completions **反向代理**（需网关鉴权与上游配置）                                                     |
| `POST /internal/reload-keys`    | 热加载密钥文件（仅当设置了 `NEXUSROUTER_ADMIN_RELOAD_TOKEN` 时注册；需 Bearer 该令牌）                                    |
| `POST /internal/reload-config`  | 重新读取 `**gateway.yaml`**（同上 Bearer；未设置 `NEXUSROUTER_GATEWAY_CONFIG_FILE` 时无操作）                       |
| `PUT /internal/upstream/active` | 仅更新内存中的 `**active_upstream_id**`（JSON：`{"active_upstream_id":"..."}`，空字符串解除 pin；**不写回磁盘**，重启后以文件为准） |


OpenAI 官方 REST 概览与认证约定见：[https://developers.openai.com/api/reference/overview](https://developers.openai.com/api/reference/overview)。

### 运行时配置（`gateway.yaml`）

- 设置 `**NEXUSROUTER_GATEWAY_CONFIG_FILE`** 指向 YAML 文件后，进程启动时与 **环境变量中的上游列表** 合并为单一运行时快照（详见仓库内 `**gateway.yaml.example`**）。
- 解析或校验失败时：**启动阶段**拒绝启动；**热加载**（`SIGHUP` 或 `POST /internal/reload-config`）失败时**保留旧快照**并打错误日志。
- **Linux / macOS**：配置文件路径非空时，`**SIGHUP`** 会触发与 `reload-config` 相同的重载逻辑。

### 中间件与限流顺序

引擎级：**CORS**（可选，来自 `gateway.yaml` 的 `cors`）→ `**X-Request-ID`** → **Recovery** → **统一 JSON 错误** → **按 IP 限流**（`rate_limit.rps_per_ip`，鉴权前；`/health`、`/openapi*`、`/swagger*`、`/internal*` 与 **OPTIONS** 跳过）→ 业务路由。

`POST /v1/chat/completions` 链：**GatewayAuth** → **按 Key 限流**（`rate_limit.rps_per_key`，鉴权成功后；**OPTIONS** 跳过）→ **ChatProxy**。

两维限流同时启用时：**任一超限即 429**（先执行 IP，再执行 Key）。超限 Zap 为 **Warn**，含 `**request_id`** 与 `**reason**`：`RATE_LIMIT_IP` / `RATE_LIMIT_KEY`。

### 代理访问日志

在 `gateway.yaml` 中配置 `**proxy_access_log**`（`enabled`、`path` 滚动、`level` 为 `**info**` 或 `**error**`）后，对完成的 Chat 代理请求写独立 JSON 行日志（与 stderr 应用日志分离）。字段含 `**request_id**`、`**method**`、`**path**`、`**client_ip**`、`**upstream_id**`、`**upstream_host**`、`**status**`、`**duration_ms**`，以及匿名 `**api_key_fp**`（鉴权成功后）。**不**记录 `Authorization` / `X-API-Key` 明文。

### 环境变量一览


| 变量                                         | 说明                                                                                                                                                     |
| ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `NEXUSROUTER_UPSTREAM_BASE_URLS`           | **多上游**：逗号分隔的基址列表；**非空时优先**于 `NEXUSROUTER_UPSTREAM_BASE_URL`，请求按 **轮询** 选择上游。                                                                          |
| `NEXUSROUTER_UPSTREAM_BASE_URL`            | **单上游**（遗留）：当 `BASE_URLS` 为空时使用；未配置任何上游时 `POST /v1/chat/completions` 返回 **503**。                                                                       |
| `NEXUSROUTER_GATEWAY_KEYS_FILE`            | **API 密钥 JSON 文件路径**（非空则优先于下方 `GATEWAY_API_KEYS`）。文件不可读或 JSON 无效时**进程启动失败**。                                                                           |
| `NEXUSROUTER_GATEWAY_API_KEYS`             | **遗留**：逗号分隔明文密钥；仅当未设置 `GATEWAY_KEYS_FILE` 时使用。每条视为启用、无过期。                                                                                              |
| `NEXUSROUTER_ADMIN_RELOAD_TOKEN`           | 非空时注册 `**POST /internal/reload-keys`**、`**POST /internal/reload-config**`、`**PUT /internal/upstream/active**`；请求须 `Authorization: Bearer <与本变量相同的令牌>`。 |
| `NEXUSROUTER_GATEWAY_CONFIG_FILE`          | 可选 `**gateway.yaml**` 路径；与 env 上游等合并，支持 **SIGHUP** / `**reload-config`** 热加载。                                                                          |
| `NEXUSROUTER_UPSTREAM_API_KEY`             | 发往上游的 Bearer（在未开启透传客户端 `Authorization` 时注入）。                                                                                                           |
| `NEXUSROUTER_FORWARD_CLIENT_AUTHORIZATION` | `true` 时将客户端 `Authorization` 原样转发上游。                                                                                                                   |
| `NEXUSROUTER_UPSTREAM_TIMEOUT`             | 等待上游响应头超时，如 `120s`（默认 120s）。                                                                                                                           |
| `NEXUSROUTER_ENABLE_SWAGGER_UI`            | 设为 `false` 关闭 Swagger UI（OpenAPI 路由仍可用）。                                                                                                               |


### 密钥文件格式（`NEXUSROUTER_GATEWAY_KEYS_FILE`）

```json
[
  {
    "id": "key-1",
    "secret": "sk-your-gateway-secret",
    "disabled": false,
    "expires_at": null
  },
  {
    "id": "key-2",
    "secret": "sk-rotated",
    "disabled": false,
    "expires_at": "2027-12-31T23:59:59Z"
  }
]
```

- `**expires_at**`：可省略或 `null` 表示永不过期；否则为 **RFC3339 / RFC3339Nano**；到达该时刻起密钥视为过期（返回 **401**）。
- `**disabled`**：`true` 时不接受该密钥。
- 生产环境请将文件权限限制为 `**0600**`，并避免将真实密钥提交到 Git。

### 探活与监控

- 负载均衡可轮询 `**GET /health**`，建议校验 HTTP **200** 且 JSON 中 `**status == "ok"`**。
- 可解析 `**server_time**` 做时钟漂移告警（策略由监控系统自定）。

### 热加载密钥与网关配置（`SIGHUP`）

- 当使用 `**NEXUSROUTER_GATEWAY_KEYS_FILE**` 时，在 **Linux / macOS** 上可向进程发送 `**SIGHUP`** 以重新读取密钥文件（无需重启）。
- 当 `**NEXUSROUTER_GATEWAY_CONFIG_FILE**` 非空时，同一 `**SIGHUP**` 也会尝试重新加载 `**gateway.yaml**`（失败则保留旧快照）。
- **Windows** 无 `SIGHUP`，请使用对应 `**POST /internal/*`** 管理接口（需配置 `NEXUSROUTER_ADMIN_RELOAD_TOKEN`）。

### 客户端鉴权

- **推荐**：`Authorization: Bearer <API_KEY>`（与 OpenAI 客户端习惯一致）。
- `**X-API-Key`**：**deprecated**，仍支持以兼容旧客户端；新集成请迁移到 Bearer。

### 文档单一事实来源（代码优先）

- **只维护 Go 源码中的 swag 注释**（`cmd/api/main.go`、`internal/handler/swagger_*.go` 等）。
- `make docs` 依次：`swag init` → 生成 `docs/`（Swagger 2）→ `swagger2openapi` 生成 `**internal/openapi/openapi.yaml`（OAS3）**。
- 运行时将 **OAS3 文件嵌入二进制**（`go:embed`），`/openapi.yaml` 与 `/openapi.json` 均来自该生成物；**请勿手改** `internal/openapi/openapi.yaml`。

```bash
cd services/gateway
make docs
```

CI 会校验 `**docs/****` 与 `**internal/openapi/openapi.yaml**` 与仓库一致（`git diff --exit-code`）。

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

