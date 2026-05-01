## Why

核心迭代已提供多上游轮询、密钥文件、统一错误与基础可观测性；在多团队、多模型供应商与前端联调并存的场景下，仍缺少：**可控的上游路由策略**（默认、权重、应急切换）、**可审计的代理访问日志**、**滥用防护（限流）**、**浏览器跨域调试体验**，以及将分散环境变量收敛为**可版本化、可热更新的单一配置源**。本变更在不破坏 OpenAI 兼容对外契约的前提下补齐上述进阶能力。

## What Changes

- **上游目标管理**：在现有多地址基础上，支持 **默认上游**、**按权重分配**、**运行时手动切换**当前生效上游（管理 API 或等价机制，见 `design.md`）；策略与轮询行为的关系 MUST 文档化。
- **请求 / 响应代理日志**：对代理路径记录 **方法、URL、关键请求头摘要、上游目标、响应状态码、耗时**；支持 **info / error** 等分级与 **日志输出路径**（文件或 stdout，二选或组合见设计）；敏感头 MUST 脱敏。
- **限流**：中间件支持按 **API Key**、**客户端 IP** 维度计数，**可配置每秒请求数（RPS）** 阈值；触发时 **429** + 网关统一错误体（含 `request_id`），且不消耗上游配额（MUST 不调用上游）。
- **跨域（CORS）**：可开关的中间件，支持配置 **允许的 Origin 列表**、**方法**、**请求头**、**预检缓存**等，便于 `web/dashboard` 本地调试对接网关。
- **配置持久化**：将上游列表、路由策略、限流、CORS、日志级别与路径等 **写入本地配置文件**（格式见 `design.md`）；支持 **热更新**（`SIGHUP` 与/或管理接口与/或定时轮询），**无需重启进程**即可生效；与现有 `NEXUSROUTER_*` 环境变量的 **优先级** MUST 在 README 固定。

## Capabilities

### New Capabilities

- `upstream-target-management`：多上游之上的 **默认项、权重、手动切换当前目标** 及与 `POST /v1/chat/completions` 的衔接。
- `proxy-access-logging`：代理链路的 **结构化访问日志**、分级、输出路径与脱敏规则。
- `http-rate-limiting`：按 **API Key / IP** 的 **RPS** 限流与 **429** 行为。
- `http-cors`：可配置的 **CORS** 中间件与预检行为。
- `gateway-config-persistence`：网关 **运行时配置文件的格式、加载、校验、热更新** 及与 env 回退策略。

### Modified Capabilities

- `openai-chat-completions-proxy`：上游选择从「仅轮询」扩展为 **由 `upstream-target-management` 定义的策略集合**（轮询 MAY 保留为默认之一）；明确 **限流与 CORS 与代理链** 的相对顺序引用其他新规范。
- `gateway-backend`：补充 **访问日志与限流失败** 与现有 **Zap / 错误 JSON** 的协同要求（字段不重复定义处引用 `proxy-access-logging` / `http-rate-limiting`）。

## Impact

- **代码**：`services/gateway` 下 `internal/config`、`internal/router`、`internal/handler`、新增 `internal/limit`、`internal/cors` 或等价包、配置加载与 **watch/signal** 路径。
- **依赖**：限流可能引入 **Redis** 或进程内 **token bucket**（设计二选一）；若选 Redis，牵动 `go.mod` 与部署文档。
- **运维**：新配置文件路径、权限、与 **Kubernetes ConfigMap reload** 等场景的说明。
- **安全**：日志与持久化文件中的 **密钥与 Authorization** 脱敏与文件权限。
