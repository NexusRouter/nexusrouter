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
make docs        # 首次克隆或修改 swag 注释后需要：生成 internal/openapi/openapi.yaml 供 go:embed
go run ./cmd/api
```

默认监听 **:8080**（可用环境变量 **`NEXUSROUTER_HTTP_LISTEN_ADDR`** 覆盖，例如 `:9090`）。

### 持久化（SQLite / Postgres）

- 进程启动时会打开 **GORM** 数据库：未设置 **`NEXUSROUTER_DATABASE_URL`** 时使用本地 **SQLite** 文件（默认文件名 **`gateway.db`**，可用 **`NEXUSROUTER_SQLITE_PATH`** 覆盖路径）。
- 设置 **`NEXUSROUTER_DATABASE_URL`**（Postgres 连接串）时改用 **Postgres**，与 SQLite 共用同一套模型与业务逻辑。
- 启动时执行 **`AutoMigrate`** 建表；若库中尚无网关 YAML / API Key / 管理员数据，且存在已配置的 **`NEXUSROUTER_GATEWAY_CONFIG_FILE`**、**`NEXUSROUTER_GATEWAY_KEYS_FILE`** 或环境变量中的管理员 bcrypt，则**一次性导入**到数据库。
- **真源**为数据库：管理 API `persist: true` 写入数据库；**`SIGHUP`** / **`POST /internal/reload-keys`** / **`reload-config`** 在数据库模式下从库重载。

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
| `POST /internal/reload-keys`    | 热加载 API Key（从数据库重新读取；仅当设置了 `NEXUSROUTER_ADMIN_RELOAD_TOKEN` 时注册；需 Bearer 该令牌）                                    |
| `POST /internal/reload-config`  | 从数据库重新加载网关快照（同上 Bearer）                       |
| `PUT /internal/upstream/active` | 仅更新内存中的 `**active_upstream_id**`（JSON：`{"active_upstream_id":"..."}`，空字符串解除 pin；**不写回持久化**，重启后以库/文件为准） |


OpenAI 官方 REST 概览与认证约定见：[https://developers.openai.com/api/reference/overview](https://developers.openai.com/api/reference/overview)。

### 运行时配置（`gateway.yaml` 与数据库）

- 默认以 **数据库** 中保存的网关 YAML 片段与环境变量中的上游列表合并为运行时快照；若曾设置 `**NEXUSROUTER_GATEWAY_CONFIG_FILE**`，该文件可在**首次启动空库**时导入数据库（详见 `**gateway.yaml.example**`）。
- 解析或校验失败时：**启动阶段**拒绝启动；**热加载**（`SIGHUP` 或 `POST /internal/reload-config`）失败时**保留旧快照**并打错误日志。
- **Linux / macOS**：已启用数据库或配置文件路径非空时，`**SIGHUP`** 会触发与 `reload-config` 相同的重载逻辑。

### 中间件与限流顺序

引擎级：**CORS**（可选，来自 `gateway.yaml` 的 `cors`）→ `**X-Request-ID`** → **Recovery** → **统一 JSON 错误** → **按 IP 限流**（全局 `rate_limit.rps_per_ip` 与 `rate_limit_rules` 中 `dimension: ip` 的规则，鉴权前；`/health`、`/openapi*`、`/swagger*`、`/internal*`、`/api/admin*` 与 **OPTIONS** 跳过）→ **IP 名单**（`ip_access`；跳过路径同上）→ 业务路由。

`POST /v1/chat/completions` 链：**GatewayAuth** → **按 Key 限流**（全局 `rps_per_key` 与规则维度 `api_key_fp`；**OPTIONS** 跳过）→ **ChatProxy**。

两维限流同时启用时：**任一超限即 429**（先执行 IP，再执行 Key）。超限 Zap 为 **Warn**，含 `**request_id`** 与 `**reason**`：`RATE_LIMIT_IP` / `RATE_LIMIT_KEY`。名单拒绝为 **403**，错误码 **`IP_BLOCKED`**。

### 进阶管理（限流规则 / CORS / IP 名单 / 日志）

在启用管理控制台时，可通过 **`/api/admin/v1`** 读写 **`rate_limit_rules`**、**`ip_access`**、**`cors`**，并查询 **`proxy_access_log`** 指向的 JSON 行日志（有扫描字节与行数上限）。误配 **IP 白名单** 时仍可通过本机 **`POST /internal/reload-config`** 或已登录管理端将 `ip_access.mode` 切回 **`off`**。CSV 导出**不包含**明文 API Key。

### 代理访问日志

在 `gateway.yaml` 中配置 `**proxy_access_log**`（`enabled`、`path` 滚动、`level` 为 `**info**` 或 `**error**`）后，对完成的 Chat 代理请求写独立 JSON 行日志（与 stderr 应用日志分离）。字段含 `**request_id**`、`**method**`、`**path**`、`**client_ip**`、`**upstream_id**`、`**upstream_host**`、`**status**`、`**duration_ms**`，以及匿名 `**api_key_fp**`（鉴权成功后）。**不**记录 `Authorization` / `X-API-Key` 明文。

### 环境变量一览


| 变量                                         | 说明                                                                                                                                                     |
| ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `NEXUSROUTER_UPSTREAM_BASE_URLS`           | **多上游**：逗号分隔的基址列表；**非空时优先**于 `NEXUSROUTER_UPSTREAM_BASE_URL`，请求按 **轮询** 选择上游。                                                                          |
| `NEXUSROUTER_UPSTREAM_BASE_URL`            | **单上游**（遗留）：当 `BASE_URLS` 为空时使用；未配置任何上游时 `POST /v1/chat/completions` 返回 **503**。                                                                       |
| `NEXUSROUTER_DATABASE_URL`               | **Postgres** 连接串；**非空**时启用 Postgres 持久化（与 GORM 兼容的 DSN）。为空则使用下方 SQLite 文件。                                                                 |
| `NEXUSROUTER_SQLITE_PATH`                  | **SQLite** 文件路径；在 `DATABASE_URL` 为空时生效，默认 **`gateway.db`**（相对进程工作目录）。                                                                 |
| `NEXUSROUTER_GATEWAY_KEYS_FILE`            | **API 密钥 JSON 文件路径**（可选）：可作为**首次启动空库**时的导入源；日常真源为数据库。未配置文件且无库内密钥、无 `GATEWAY_API_KEYS` 时，受保护路由将无可用密钥。                                                                           |
| `NEXUSROUTER_GATEWAY_API_KEYS`             | **遗留**：逗号分隔明文密钥；可作为空库导入源或测试。                                                                                              |
| `NEXUSROUTER_ADMIN_RELOAD_TOKEN`           | 非空时注册 `**POST /internal/reload-keys`**、`**POST /internal/reload-config**`、`**PUT /internal/upstream/active**`；请求须 `Authorization: Bearer <与本变量相同的令牌>`。 |
| `NEXUSROUTER_GATEWAY_CONFIG_FILE`          | 可选 `**gateway.yaml**` 路径：可作为**首次启动空库**导入源；运行态与 env 上游等在数据库中合并，支持 **SIGHUP** / `**reload-config`** 热加载。                                                                          |
| `NEXUSROUTER_UPSTREAM_API_KEY`             | 发往上游的 Bearer（在未开启透传客户端 `Authorization` 时注入）。                                                                                                           |
| `NEXUSROUTER_FORWARD_CLIENT_AUTHORIZATION` | `true` 时将客户端 `Authorization` 原样转发上游。                                                                                                                   |
| `NEXUSROUTER_UPSTREAM_TIMEOUT`             | 等待上游响应头超时，如 `120s`（默认 120s）。                                                                                                                           |
| `NEXUSROUTER_ENABLE_SWAGGER_UI`            | 设为 `false` 关闭 Swagger UI（OpenAPI 路由仍可用）。                                                                                                               |
| `NEXUSROUTER_ENABLE_ADMIN_CONSOLE`         | `true` 时启用 **`/api/admin/v1/*`** 管理 HTTP API（仍需下方 JWT 与账号配置完整）。                                                                                         |
| `NEXUSROUTER_ADMIN_JWT_SECRET`             | 管理端 JWT **HMAC** 密钥（建议随机长字符串，勿提交到 Git）。                                                                                                         |
| `NEXUSROUTER_ADMIN_JWT_EXPIRE`             | 访问令牌有效期，如 `24h`（默认 24h）。                                                                                                                               |
| `NEXUSROUTER_ADMIN_REFRESH_EXPIRE`         | 「记住我」时的最长会话，如 `168h`（默认 7d）；须大于等于 `ADMIN_JWT_EXPIRE` 方生效。                                                                                          |
| `NEXUSROUTER_ADMIN_USERNAME`               | 管理员登录用户名。                                                                                                                                                    |
| `NEXUSROUTER_ADMIN_PASSWORD_BCRYPT`        | 管理员密码 **bcrypt** 哈希（`$2a$...`）；可用 `go run golang.org/x/crypto/bcrypt` 或运维工具生成。**禁止**在环境变量中存放明文生产密码。                                      |
| `NEXUSROUTER_ADMIN_OPERATOR_USERNAME`      | （可选）**操作员**登录用户名；与管理员账号独立。需同时设置 `NEXUSROUTER_ADMIN_OPERATOR_PASSWORD_BCRYPT` 才生效。操作员 JWT `role` 为 `operator`，管理 API **写**路由返回 **403**（`POST /auth/logout` 除外）。 |
| `NEXUSROUTER_ADMIN_OPERATOR_PASSWORD_BCRYPT` | （可选）操作员密码 **bcrypt** 哈希。 |
| `NEXUSROUTER_HTTP_LISTEN_ADDR`             | （可选）网关 HTTP 监听地址，默认 `:8080`。 |
| `NEXUSROUTER_ADMIN_PASSWORD_RESET_SMTP`    | （预留）邮件重置相关配置；未配置时「忘记密码」仅返回运维指引。                                                                                                                    |

### 首次初始化（`/api/bootstrap/v1`）

数据库表 **`system_bootstrap`** 保存全局 **`initialized`** 标志。未完成首次向导时，网关对除白名单外的 HTTP 请求返回 **403**（`code: BOOTSTRAP_REQUIRED`）；白名单含 **`GET /health`**、**`GET /api/bootstrap/v1/status`**、**`POST /api/bootstrap/v1/complete`**、**`POST /api/admin/v1/auth/login`**、**`GET /api/admin/v1/auth/password-reset-info`** 及 OpenAPI/Swagger 静态路径等。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/bootstrap/v1/status` | 匿名；返回 `initialized`（bool）、`phase`（`ready` \| `initializing` \| `completed`）。 |
| `POST` | `/api/bootstrap/v1/complete` | 匿名；body：`admin_username`、`admin_password`（≥8）、可选 `site_display_name`；**须**已配置 **`NEXUSROUTER_ADMIN_JWT_SECRET`** 且启用管理控制台。成功至多一次；重复返回 **409**（`BOOTSTRAP_ALREADY_COMPLETED`）。 |
| `POST` | `/api/bootstrap/v1/reset` | 需 **`Authorization: Bearer`** 且 JWT `role` 为 **`admin`**（非 operator）；清空 **`admin_users`** 并将 **`initialized`** 置回 **false**，用于运维重新走向导。 |

在 **`initialized=false`** 时，**不会**从环境变量 **`NEXUSROUTER_ADMIN_USERNAME` / `NEXUSROUTER_ADMIN_PASSWORD_BCRYPT`** 静默创建管理员（避免绕过向导）；向导完成后或 **`initialized=true`** 且库中仍无用户时，仍可按既有逻辑从 env 导入。

### 管理控制台 API（`/api/admin/v1`）

当 `**NEXUSROUTER_ENABLE_ADMIN_CONSOLE=true**` 且 JWT 密钥已配置，且（**未完成首次向导** 或 **数据库中存在 admin 用户** 或环境变量中已配置 **`NEXUSROUTER_ADMIN_USERNAME`** + **`NEXUSROUTER_ADMIN_PASSWORD_BCRYPT`**）时注册：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/admin/v1/auth/login` | 登录，body：`username`、`password`、`remember`；返回 `access_token`（Bearer）、`role`（`admin` \| `operator`）等。 |
| `GET` | `/api/admin/v1/auth/me` | 当前 JWT 对应用户名与 `role`。 |
| `GET` | `/api/admin/v1/auth/password-reset-info` | 忘记密码说明（无需鉴权）。 |
| `POST` | `/api/admin/v1/auth/logout` | 登出（客户端丢弃令牌即可，服务端无状态）。 |
| `GET` | `/api/admin/v1/system/settings` | 聚合可读系统设置项（含 `mutability` 元数据）。 |
| `PUT` | `/api/admin/v1/system/settings` | 热更新类字段（当前实现含 `proxy_access_log`）；监听类项需改环境变量并重启。 |
| `GET` | `/api/admin/v1/alerts/status` | 运行态告警状态（依赖 `gateway.yaml` 中 `admin_alerts` 与进程内指标）。 |
| `GET` | `/api/admin/v1/metrics/summary` | 进程内指标：请求量、成功率、平均耗时、今日/昨日、按 `code` 分桶错误等。 |
| `GET` | `/api/admin/v1/gateway/snapshot` | 当前快照：`upstreams`、`routing`、`cors`、`rate_limit`、`rate_limit_rules`、`ip_access`、`proxy_access_log` 等。 |
| `PUT` | `/api/admin/v1/gateway/config` | 替换 `upstreams` + `routing`；`persist: true` 时写入**数据库**并刷新内存快照。 |
| `PUT` | `/api/admin/v1/gateway/active-upstream` | 设置 `active_upstream_id`；可选 `persist: true` 写回持久化。 |
| `GET` / `PUT` | `/api/admin/v1/gateway/cors` | 读写 CORS 段；`PUT` 可带 `allow_origins_bulk`（换行/逗号分隔）与 `persist`。 |
| `GET` / `PUT` | `/api/admin/v1/gateway/rate-limit-rules` | 读写 `rate_limit_rules` 数组（整体替换）；`PUT` body：`{ "rules": [...], "persist": true }`。 |
| `GET` / `PUT` | `/api/admin/v1/security/ip-access` | 读写 `ip_access`；`PATCH` 支持 `add` / `remove` CIDR 列表与可选 `mode`。 |
| `GET` | `/api/admin/v1/logs/query` | 日志筛选与分页（`from`、`to`、`path_prefix`、`status_min`、`status_max`、`api_key_fp`、`client_ip`、`limit`、`cursor`）。 |
| `GET` | `/api/admin/v1/logs/export.csv` | 同筛选条件的 CSV 流式下载。 |
| `GET` | `/api/admin/v1/keys` | 列出 API Key（脱敏），依赖数据库中已有密钥记录。 |
| `POST` | `/api/admin/v1/keys` | 新建密钥（响应体一次性返回 `secret`）。 |
| `PATCH` | `/api/admin/v1/keys/:id` | 更新 `disabled` / `expires_at`。 |
| `DELETE` | `/api/admin/v1/keys/:id` | 删除。 |
| `POST` | `/api/admin/v1/keys/batch-disable` | body：`{ "ids": ["..."] }`。 |
| `POST` | `/api/admin/v1/keys/batch-delete` | body：`{ "ids": ["..."] }`。 |

除 `login` 与 `password-reset-info` 外，均须在请求头携带 **`Authorization: Bearer <access_token>`**。

**安全建议**：管理 API 与控制台应仅在内网或通过 **HTTPS** 暴露；JWT 与 API Key 勿写入前端仓库；生产环境将 `NEXUSROUTER_ENABLE_SWAGGER_UI` 设为 `false` 时，仪表盘「接口调试」页依赖同源代理访问 `/swagger`（仍须单独评估是否对管理网段开放）。

### Web 仪表盘（`web/dashboard`）

- 本地开发：`cd web/dashboard && pnpm dev`，Vite 已将 **`/api`**、**`/health`**、**`/openapi.json`**、**`/swagger`** 代理到 **`http://127.0.0.1:8080`**（请先启动网关）。
- 生产部署：将构建产物与网关同域，或设置 **`VITE_API_BASE_URL`** 指向网关根 URL。

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
- `**created_at**`：可选，RFC3339；管理端新建密钥时会自动写入。
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

- **只维护 Go 源码中的 swag 注释**（`cmd/api/main.go`、`internal/handler/swagger_*.go` 等）；`docs/docs_test.go` 为手写保留，**`swag init` 不会删除**，用于 `go test -cover` 与生成物冒烟。
- `make docs` 依次：`swag init` → 生成 `docs/`（Swagger 2）→ `swagger2openapi` 生成 `**internal/openapi/openapi.yaml`（OAS3）**。
- 运行时将 **OAS3 文件嵌入二进制**（`go:embed`），`/openapi.yaml` 与 `/openapi.json` 均来自该生成物；**请勿手改** `internal/openapi/openapi.yaml`（该文件由 `make docs` 产出，**已加入 .gitignore**，不入库）。

```bash
cd services/gateway
make docs
```

克隆或改注释后：**先** `make docs` **再** `go test` / `go build`（或使用 **`make test`**、**`make build-api`**，二者均依赖 `docs`）。CI 在测试与编译前会执行 `make docs`，且仅对 **`docs/`** 做 `git diff --exit-code` 防漂移（`openapi.yaml` 不纳入版本对比）。

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

