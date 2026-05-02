## Why

运维与集成方需要在网关侧维护**可检索、可绑定上游、可对客户端声明**的模型集合；仅靠手写 `gateway.yaml` 与凭记忆无法规模化。应在 NexusRouter 内提供与 OpenAI 生态兼容的**模型发现与目录**能力，并与现有上游、密钥、管理控制台体系一致。本变更在**不依赖、不引用任何特定第三方开源项目**的前提下，采纳业界常见的「目录 + 绑定 + 可选别名 + 可选上游发现同步」模式。

## What Changes

- 新增**模型库**持久化模型：模型目录项、与配置中上游 id 的**可用性绑定**、**alias**（**首版必做**：`POST /v1/chat/completions` 在命中上游后，按绑定将请求 JSON 的 **`model`** 改写为 **`actual_model`** 再转发）。
- 新增管理端 REST：**列表 / 创建 / 更新 / 删除 / 批量导入**，以及**从指定上游拉取 `/v1/models` 同步**（可选使用上游 API Key，不落库完整密钥）。
- 新增或扩展网关公开路由：**GET `/v1/models`**（及可选 **GET `/v1/models/:id`**），响应形状与 [OpenAI List models](https://platform.openai.com/docs/api-reference/models/list) **子集对齐**，且仅暴露**已启用且绑定有效上游**的模型；与现有 **POST `/v1/chat/completions`** 使用同一套网关入口鉴权（Bearer / 可选 X-API-Key）。
- 控制台新增**模型库**菜单与页面：表格、筛选、绑定上游、别名编辑、同步操作、与 i18n 一致。
- OpenAPI / 嵌入文档随新路径更新（与仓库「可选生成」策略一致）。
- **非目标（首版不实现）**：按模型维度的计费配额、多租户分组路由矩阵、**按请求 model 自动选择不同上游**（与当前「先 Picker 选路、再按绑定改写 model」不同；可后续变更）。

## Capabilities

### New Capabilities

- `model-library`: 模型元数据目录、上游可用性绑定、别名、管理 API、可选上游发现同步、公开 `GET /v1/models` 子集、控制台页面与验收场景。

### Modified Capabilities

- `gateway-backend`: 增加模型库相关管理路由、持久化与迁移、错误体与鉴权与现有 `/api/admin/v1` 一致。
- `dashboard-frontend`: 增加模型库导航项与独立页面，符合现有 Ant Design + React Query 模式。
- `openai-chat-completions-proxy`: 增加与 OpenAI 兼容的 **GET `/v1/models`**（及可选 **Retrieve model**）的规范要求与 OpenAPI 描述。

## Impact

- **代码**：`services/gateway`（`internal/handler`、`internal/repository`、迁移、`router`）、`web/dashboard`（新页面、路由、`api` 封装）。
- **数据**：新表；需 GORM 迁移或与现有 SQLite/Postgres 策略一致。
- **运维**：可选环境变量（如同步超时、是否启用公开 list models）。
- **文档**：用户文档仅描述 NexusRouter 能力与配置键；**不出现**外部项目名或「与某某项目对齐」类表述。
