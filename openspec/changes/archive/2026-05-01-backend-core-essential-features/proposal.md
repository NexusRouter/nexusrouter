## Why

网关需要在一轮「基础迭代」内把运维可观测性、上游弹性、租户级访问控制与可调试契约对齐到产品预期：外部监控依赖明确的健康语义；多模型/多供应商要求可配置多上游；静态单密钥无法支撑密钥轮换与失效窗口；生产排障需要请求维度的统一错误载体；团队交付依赖可机读且可交互的 OpenAPI 3.0 文档。

## What Changes

- 扩展 **`GET /health`**：在 JSON 中返回服务运行状态、应用/构建版本标识、服务端当前时间（RFC3339），便于探针与合成监控解析。
- **OpenAI 兼容代理**：保持 **`POST /v1/chat/completions`**；支持配置 **多个上游基址**，按约定策略（如轮询、随机或权重，具体在 `design.md` 固定）选择目标；在剔除 hop-by-hop 及安全例外后 **尽可能透传客户端请求头与请求体**，上游 HTTP 状态码与响应体 **原样** 回传客户端。
- **API Key 鉴权**：中间件校验 **`Authorization: Bearer <API_KEY>`**；支持密钥 **新增、禁用、可选有效期**；缺失、无效、已禁用或已过期一律 **401**，且 **不调用上游**；持久化与配置热路径在 `design.md` 明确（文件/DB 等）。
- **统一错误响应**：所有网关自身产生的错误（鉴权、路由、代理失败、超时、panic 捕获等）使用 **同一 JSON 结构**，包含 **HTTP 状态码语义**、**机器可读 `code`**、**人类可读 `message`**、**请求 ID**（与中间件注入的 request id 对齐）；上游成功/失败路径仍按现有代理规范 **透传上游状态码与 body**（不包一层网关错误壳）。
- **OpenAPI 3.0 + Swagger UI**：延续 **swaggo/swag** 从注释生成；为 **`/health`**、**`/v1/chat/completions`** 及（若引入）**管理用 API** 补充注释与 OAS3 片段，使 **Swagger UI** 可在线查看参数并发起调试请求。

## Capabilities

### New Capabilities

- `gateway-health`：`GET /health` 的字段契约、版本与时间语义、与监控系统的使用约定。
- `api-key-management`：Bearer API Key 校验、密钥元数据（启用/禁用、过期时间）、401 场景及与代理链路的先后顺序。

### Modified Capabilities

- `openai-chat-completions-proxy`：从单上游扩展为 **多上游可配置与选择策略**；澄清 **请求头透传** 范围与 hop-by-hop 剔除列表；OpenAPI/Swagger 覆盖新增与变更的 HTTP 面。
- `gateway-backend`：**统一错误 JSON** 必须包含 **`request_id`**（及与请求头关联规则），并与 Zap 日志字段对齐；必要时补充 **请求 ID 中间件** 的全局行为要求。

## Impact

- **代码**：`services/gateway` 下 `internal/router`、`internal/handler`、`internal/config`、鉴权与可能的 `internal/service` / `internal/repository`；`cmd/api` 与 `docs` 生成物；Wire 注入图。
- **配置**：Viper/环境变量新增多上游列表、API Key 存储路径或数据源等。
- **依赖**：若密钥持久化选用数据库，牵动 **GORM/迁移**；否则可能为启动时加载的 **本地配置文件**（设计阶段二选一或组合）。
- **规范**：`openspec/specs/gateway-backend`、`openspec/specs/openai-chat-completions-proxy` 将随本变更增补；新增 `gateway-health`、`api-key-management` 规格目录。
