## Context

NexusRouter 网关（`services/gateway`）已具备：`GET /health`、OpenAPI/Swagger（`/openapi.yaml`、`/swagger/*`）、`POST /v1/chat/completions` 鉴权与代理、基于 **`gateway.yaml`** 的运行时快照（`internal/runtime`）、`PUT /internal/upstream/active` 与 `POST /internal/reload-config` 等管理端点（受 **`NEXUSROUTER_ADMIN_RELOAD_TOKEN`** 保护）、API Key 自 JSON 文件加载（`api-key-management` 规范）。仪表盘前端（`web/dashboard`）为 Vite + React + Ant Design，已有目录与依赖基线（`dashboard-frontend` 规范）。

本变更在**不破坏**现有 OpenAI 兼容路径与错误体约定的前提下，增加**管理控制台**能力：认证、指标、上游与 Key 的可视化运维、以及内嵌 API 调试。

## Goals / Non-Goals

**Goals:**

- 提供管理员**账号密码登录**，登录后持有**权限令牌**（会话）；未登录访问受保护管理功能时**跳转登录页**。
- 支持**记住密码**（安全可用的客户端持久化策略）与**忘记密码**（可落地的重置路径，见决策）。
- **仪表盘**展示网关在线状态、请求量、成功率、平均耗时；**今日/昨日**请求对比；**错误类型**统计。
- **上游配置**可在面板中增删改、设默认、查看**当前生效**上游；变更**实时**反映到运行时（与磁盘配置一致性策略见决策）。
- **API Key** 列表（脱敏）、状态、有效期、创建时间；新增、禁用、删除、**批量**操作；与网关鉴权使用的数据源一致。
- 管理面板内**嵌入 Swagger UI**（或等价），在已登录或策略允许的前提下调试网关 API，无需单独打开独立 Swagger URL。

**Non-Goals:**

- 替换或弱化现有 **Bearer 管理令牌**运维路径（可并存，逐步收敛到统一控制台鉴权）。
- 完整多租户 RBAC、审计日志持久化到外部 SIEM（可作为后续增强；本设计仅预留钩子）。
- 上游/Key 存储从文件**完全迁移**到数据库（除非实现阶段明确选型；默认仍以文件为源、管理 API 写回文件）。

## Decisions

1. **管理端认证模型**  
   - **选定**：独立「控制台会话」与 **`NEXUSROUTER_ADMIN_*` 环境配置**结合：首版可采用**用户名/密码哈希**存于配置或单独 admin 文件，登录成功后签发 **JWT**（或 signed session cookie），前端 **`Authorization: Bearer <admin_jwt>`** 或 **HttpOnly Cookie** 二选一，在 `tasks.md` 实现时固定一种。  
   - **理由**：与现有 Gin 栈、已依赖的 **jwt/v5** 一致；与 OpenAI 客户端 Bearer 区分命名空间（路径前缀 `/admin` 或独立 middleware）。  
   - **备选**：仅复用 `AdminReloadToken` 作为唯一凭证——拒绝，无法满足「账号密码登录」与登出/过期语义。

2. **记住密码**  
   - **选定**：仅将**用户名**与「是否自动登录」偏好存 **localStorage**；**密码**不以明文落盘；若勾选「记住我」则签发**较长有效期刷新令牌**或延长 JWT 过期（具体倍数在实现中配置）。  
   - **备选**：浏览器密码管理器依赖——可作为文档建议，不作为唯一实现。

3. **忘记密码**  
   - **选定**：**可配置**路径——若设置 **`NEXUSROUTER_ADMIN_RESET_EMAIL_*`**（或等价）则发送一次性重置链接；**未配置**时提供「联系运维」文案 + **CLI/环境变量重置**文档（`design.md`/`README`）。  
   - **理由**：自托管场景常无 SMTP；避免阻塞首版交付。

4. **仪表盘指标来源**  
   - **选定**：进程内 **原子计数器 + 滑动/固定窗口**（成功率、QPS、延迟直方图或均值）；**今日/昨日**对比基于 UTC 日界或配置时区；**错误类型**从网关统一 **`code`** 字段聚合（与现有 JSON 错误一致）。  
   - **备选**：Prometheus 拉取——可作为后续；首版不强制外部依赖。

5. **上游配置持久化与实时生效**  
   - **选定**：以 **`GatewayConfigFile` 指向的 YAML** 为权威时：管理 API 在校验通过后 **原子写文件**并调用现有 **`Runtime.Reload()`**（或等价）使内存快照更新；**`SetActiveUpstream`** 继续仅改内存时，面板需明确展示「未写盘」与「提供持久化」两个操作（或合并为单一「应用并保存」）。  
   - **理由**：与当前 `runtime.Store` 模型一致，避免双源。

6. **API Key 管理写路径**  
   - **选定**：管理 API **读写与 `NEXUSROUTER_GATEWAY_KEYS_FILE` 同一路径 JSON**；写后调用现有 **`KeyStore` 重载**（与 `reload-keys` 相同语义）；列表响应 **`secret` 脱敏**（仅首尾片段）。**`created_at`** 若文件无该字段，可由实现补充为可选元数据或首次导入时间。  
   - **理由**：保持与 `api-key-management` 单源一致。

7. **内嵌 Swagger UI**  
   - **选定**：管理页以 **`iframe`** 嵌入同源 **`/swagger/index.html`**（或 Vite 代理到网关 origin），并依赖**控制台登录**：路由层对 `/swagger` 可增加「仅本机或已登录」策略（实现二选一：**关闭公网 Swagger**、仅控制台内嵌可见；或 Swagger 仍公开、仅强调面板便捷入口——以安全评审为准，首版倾向**与 admin 同域 + 可选关闭独立入口**）。  
   - **备选**：npm 包内嵌 swagger-ui-react 拉取 `/openapi.json`——可作为增强，减少 iframe 限制。

## Risks / Trade-offs

- **[Risk] 文件并发写入**（多人或脚本与管理台同时改 YAML/JSON）→ **Mitigation**：写前读校验、原子替换、可选文件锁或 **ETag/版本** 字段冲突检测。  
- **[Risk] JWT 泄露** → **Mitigation**：短 TTL、HTTPS 强制说明、HttpOnly Cookie 选项、管理接口限流。  
- **[Risk] 指标内存增长** → **Mitigation**：有界环形缓冲、按日重置、可配置保留时长。  
- **[Trade-off] 忘记密码无邮件时依赖运维** → 文档明确 CLI/配置重置步骤。

## Migration Plan

1. 引入新环境变量（admin 用户、JWT 密钥、可选 SMTP）与 **`/admin` 或 `/api/admin`** 路由组；默认**关闭**控制台直至显式启用（避免意外暴露）。  
2. 前端增加路由与布局；开发环境通过 Vite proxy 连网关。  
3. 部署：先发布后端再前端，或同版本；回滚时关闭控制台 feature flag，保留文件编辑运维路径。

## Open Questions

- 控制台是否**默认启用**或必须 **`NEXUSROUTER_ENABLE_ADMIN_CONSOLE=true`**（推荐后者，待任务阶段确认）。  
- **CORS**：管理 API 与仪表盘同域部署时是否仍需额外 CORS（通常不需要）。  
- **国际化**：首版仅中文 UI 是否可接受（提案已用中文描述，默认可接受）。
