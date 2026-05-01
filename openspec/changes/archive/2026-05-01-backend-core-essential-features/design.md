## Context

NexusRouter 网关（`services/gateway`，Gin）已具备：`GET /health` 极简 JSON、`POST /v1/chat/completions` 反向代理、基于环境变量 `NEXUSROUTER_GATEWAY_API_KEYS` 的静态密钥列表、`X-Request-ID` 注入、部分统一 JSON 错误、swag 生成 OpenAPI 与可开关的 Swagger UI。现有 `openai-chat-completions-proxy` 与 `gateway-backend` 规格已描述单上游、有限请求头转发与静态鉴权。

本变更将上述能力收敛到「可运维、可多供应商、可密钥治理、可排障、可文档化调试」的基础迭代目标。

## Goals / Non-Goals

**Goals:**

- `/health` 返回 **status**、**version**（构建或发布标识）、**time**（服务端 UTC RFC3339），供合成监控解析。
- **多上游** 可配置（至少两个基址 URL），以 **轮询（round-robin）** 作为默认选择策略；策略在配置中可扩展占位（如未来 `random`/`weighted`）。
- **请求头**：在反向代理中复制客户端传入的 header（`http.Header.Clone` 语义），并 **剔除 hop-by-hop 与由 Go `ReverseProxy` 重算的头**；名单固定为：**`Connection`**, **`Keep-Alive`**, **`Proxy-Authenticate`**, **`Proxy-Authorization`**, **`Te`**, **`Trailer`**, **`Transfer-Encoding`**, **`Upgrade`**（及空名头）。**`Host`**：显式设置为目标上游 host（避免把客户端 Host 传给错误上游）。**`Content-Length`**：由框架/上游传输链重算时可不转发客户端值，以实际 body 为准。
- **请求体**：仍以 **未改动的字节流** 转发（不解析 JSON）。
- **响应**：保持现有契约——上游状态码与 body **原样** 回传；网关自身错误才使用统一 JSON 外壳。
- **API Key**：以 **`Authorization: Bearer <API_KEY>`** 为唯一网关凭证来源（与用户需求一致）；支持 **启用/禁用**、**可选过期时间**；存储首版采用 **JSON 配置文件**（路径由 `NEXUSROUTER_GATEWAY_KEYS_FILE` 指定），进程启动时加载，并支持 **`SIGHUP` 触发热加载**（Unix）及文档化的 **运维侧替换文件** 流程；可选 **`NEXUSROUTER_ADMIN_RELOAD_TOKEN`** 保护下的 **`POST /internal/reload-keys`**，仅用于无法发信号的环境。
- **统一错误 JSON**：网关生成的错误体 MUST 包含 **`code`**（机器可读）、**`message`**（人类可读）、**`request_id`**（与 **`X-Request-ID`** 响应头一致，若客户端已传入则沿用其值）。Zap 日志 MUST 带同名字段。
- **OpenAPI/Swagger**：为 `/health`、鉴权约定、（若实现）`/internal/reload-keys` 补充 swag 注释并重新 `swag init`，保证 Swagger UI 可调试。

**Non-Goals:**

- Dashboard 前端上的密钥管理 UI、多租户 RBAC、按模型路由（model-based routing）、配额与限流、审计日志持久化。
- 将上游错误 JSON 再包装为网关统一格式（会破坏 OpenAI 兼容客户端解析）。
- 跨地域主动健康探测上游（仅客户端请求驱动）。

## Decisions

1. **多上游列表配置**  
   - **决策**：新增 `NEXUSROUTER_UPSTREAM_BASE_URLS`（逗号分隔 URL 列表）；若同时存在旧键 `NEXUSROUTER_UPSTREAM_BASE_URL`，**列表优先**；列表为空则回退单键行为以保持兼容。  
   - **备选**：仅保留单键并支持 DNS 轮询——放弃，无法满足「配置多个明确目标」的验收表述。

2. **上游选择**  
   - **决策**：进程内 **`sync/atomic` 轮询索引** 选择下一个基址；无健康检查加权。  
   - **备选**：随机——可测性较差；加权——超出基础迭代。

3. **API Key 持久化**  
   - **决策**：JSON 数组，元素字段：`id`（UUID 字符串，日志与排障用）、`secret`（明文比对，**文件权限 MUST 0600** 且文档警告）、`disabled`（布尔）、`expires_at`（RFC3339 或 null）。  
   - **备选**：仅环境变量——无法实现「禁用/有效期」而不重启；Postgres——引入迁移与连接配置，本迭代延后。

4. **鉴权与上游 Authorization 冲突**  
   - **决策**：沿用现有 `NEXUSROUTER_FORWARD_CLIENT_AUTHORIZATION`：为 false 时，网关用 `NEXUSROUTER_UPSTREAM_API_KEY` 重写发往上游的 `Authorization`；为 true 时原样转发客户端 `Authorization`（此时客户端 Bearer 同时承担网关与上游身份，文档 MUST 警示）。  
   - **备选**：拆分 `Authorization-Internal` 自定义头——破坏「与 OpenAI 一致」的客户端体验。

5. **405 vs 404**  
   - **决策**：对非法方法固定 **405**，与现有 `router` 中 `METHOD_NOT_ALLOWED` 对齐并写入规格。

6. **版本号来源**  
   - **决策**：`main` 通过 **`-ldflags "-X ...Version=..."`** 注入；未注入时 **`version` 字段为 `dev`**。时间字段 **`server_time`** 使用 `time.Now().UTC().Format(time.RFC3339Nano)`。

## Risks / Trade-offs

- **[Risk] 明文密钥文件泄露** → Mitigation：文件权限、不入库 Git、生产推荐后续接入密管或 DB 哈希存储。  
- **[Risk] 热加载与进行中的请求** → Mitigation：原子替换指针到 key 切片；比对在锁或 RCU 语义下只读。  
- **[Risk] 全量透传头导致意外 Cookie/内部头泄漏到上游** → Mitigation：在 `design` 列出 hop-by-hop；可选后续「剥离前缀 `X-Internal-*`」作为增强项记入 Open Questions。  
- **[Risk] Windows 无 SIGHUP** → Mitigation：提供 `POST /internal/reload-keys` 备选路径。

## Migration Plan

1. 发布前：文档说明新环境变量与 **keys JSON** 格式；提供从 `NEXUSROUTER_GATEWAY_API_KEYS` 迁移到文件的示例脚本（运维一次性）。  
2. 部署：先写 keys 文件并挂载卷，再滚动发布；灰度期间可同时保留旧 env 列表，实现按 `design` 优先级读取直至移除。  
3. 回滚：恢复旧版本二进制仍可读单 `UPSTREAM` 与逗号 keys env（若实现兼容层）。  
4. 监控：将探针从「仅 HTTP 200」升级为解析 `/health` JSON 中 `status == "ok"`（可选）。

## Open Questions

- 是否在下一迭代将 **`X-API-Key`** 从网关入口 **彻底移除**（本提案以 Bearer 为主，可在实现阶段保留只读兼容一个版本）。  
- 多上游是否需要 **粘性会话**（同一 `X-Request-ID` 或 API key hash 路由到固定上游）——当前非目标，若 LLM 有状态需求再议。
