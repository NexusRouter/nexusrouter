## Context

网关已完成：多上游（轮询）、密钥文件与热加载、统一错误与 `X-Request-ID`、基础 `/health`、OpenAPI/Swagger。配置仍以 **环境变量 + 零散文件**（如 keys JSON）为主。进阶需求要求：**可运营的路由策略**、**可检索的代理访问轨迹**、**防刷限流**、**浏览器 CORS**、以及 **单一配置源的热更新**。

## Goals / Non-Goals

**Goals:**

- **上游**：在现有多地址列表上增加 **默认上游 id**、**权重（整数或相对权重）**、**当前生效上游** 概念；提供 **受保护的管理 API**（Bearer 管理令牌，与业务 API Key 分离）或 **`SIGHUP` + 配置片段原子替换** 实现「手动切换」；策略 **round-robin / weighted-random / pinned-default** 中至少实现 **weighted-random + default fallback**，其余可在实现中标记 experimental。
- **访问日志**：对 **`POST /v1/chat/completions`**（MAY 含 **`OPTIONS`** 预检若启用 CORS）记录一行结构化日志：**timestamp、request_id、method、path、client_ip、resolved_api_key_id（哈希或匿名占位）、upstream_id/host、status、duration_ms**；请求头为 **允许名单 + 截断** 或 **仅记录头名列表**；**info** 记录成功与 4xx/5xx 上游透传状态，**error** 记录网关自身错误与上游连接失败；支持 **文件路径 + 滚动**（按大小或按日）与 **stdout**。
- **限流**：令牌桶或漏桶，**每 API Key** 与 **每 IP** 独立计数（二者同时配置时取 **更严** 或 **先命中先拒绝**，在 `design.md` 固定一种）；超限 **429**，`Retry-After` MAY 设置；**进程内** 实现为默认，**Redis** 为可选扩展（若启用则文档化部署）。
- **CORS**：`gin` 中间件或等价；配置 **AllowOrigins**（精确列表或通配规则子集）、**AllowMethods**、**AllowHeaders**、**MaxAge**；默认关闭。
- **持久化配置**：单一 **`gateway.yaml`**（或 JSON，在实现中固定一种）包含上游、路由、日志、限流、CORS；**加载顺序**：文件 → 环境变量覆盖（仅当显式 `env_override: true` 或等价）或 **文件优先、env 仅补默认**——在 README 二选一文案化；热更新 **`SIGHUP`** + **可选 `POST /internal/reload-config`**（与现有 reload-keys 可合并为统一 admin router）。

**Non-Goals：**

- Dashboard UI 上的可视化配置编辑器、多租户控制台、分布式追踪全链路（OpenTelemetry）、WAF 规则引擎。
- 将上游 LLM 计费与配额持久化到网关数据库（本阶段不引入强制 DB）。
- 修改 OpenAI JSON body 的语义或 `/v1/chat/completions` 路径。

## Decisions

1. **配置文件格式**  
   - **决策**：**YAML**（`gateway.yaml`），便于人类编辑与注释。  
   - **备选**：JSON5/TOML——工具链成熟度在 Go 侧略逊或需新依赖。

2. **限流存储**  
   - **决策**：默认 **内存 `x/time/rate` 或分段锁 map**；可选 **`NEXUSROUTER_RATE_LIMIT_BACKEND=redis`** 使用现有 **go-redis**。  
   - **备选**：纯 Redis——增加运维耦合，作为进阶开关保留。

3. **管理 API 认证**  
   - **决策**：复用 **`NEXUSROUTER_ADMIN_RELOAD_TOKEN`** 模式，扩展为 **`NEXUSROUTER_ADMIN_TOKEN`**（或保留旧名兼容），所有 **`/internal/*`** 管理路由共用 Bearer。  
   - **备选**：mTLS——超出本迭代。

4. **访问日志与 Zap 关系**  
   - **决策**：访问日志 **独立 `zapcore.Core` 写文件**（或 `lumberjack` 滚动），与现有 stderr Zap **分离**，避免污染应用错误日志；字段 schema 在 `proxy-access-logging` 规格中固定。  
   - **备选**：单 logger 多 encoder——配置复杂度高。

5. **CORS、限流与鉴权顺序**  
   - **决策**：**CORS（含 OPTIONS）→ RequestID →（可选）per-IP 限流 → GatewayAuth →（可选）per-Key 限流 → ChatProxy**；**OPTIONS** 的计数策略见 **`http-rate-limiting`** 规格。  
   - **备选**：单一限流中间件全在鉴权后——无法在无 key 时防刷 IP，故不采用。

6. **上游手动切换**  
   - **决策**：配置文件字段 **`active_upstream_id`** 或 **`routing.override_target_id`**，热加载后下一轮请求生效；**`PUT /internal/upstream/active`** MAY 作为快捷方式直接写内存并异步刷盘（若实现，须防并发写坏文件，采用临时文件 rename）。  
   - **备选**：仅文件编辑——满足「手动」但 UX 差，故 API 为可选增强。

## Risks / Trade-offs

- **[Risk] 配置文件写坏导致服务不可用** → Mitigation：启动时 schema 校验失败则拒绝加载并保留旧快照；文档提供 `gateway.yaml.example`。  
- **[Risk] 访问日志磁盘打满** → Mitigation：滚动与最大保留份数；**error** 级别时减少字段。  
- **[Risk] 内存限流多 key 膨胀** → Mitigation：TTL 清理 inactive key；文档建议 Redis 大流量场景。  
- **[Risk] CORS 过宽导致 CSRF 面** → Mitigation：禁止 `*` + credentials 组合；默认关闭 CORS。

## Migration Plan

1. 发布 **`gateway.yaml.example`**，将现有 env 等价项迁移说明表列在 README。  
2. 首版允许 **仅 env** 运行（文件缺失时行为与当前一致），**渐进启用** 文件配置。  
3. 灰度：先开访问日志与 CORS，再开限流与权重路由。  
4. 回滚：关闭文件配置开关或移除 `gateway.yaml` 挂载，进程回退 env-only。

## Open Questions

- 管理 API 是否合并 **`reload-keys`** 与 **`reload-config`** 为 **`POST /internal/reload`** 单端点（携带 `?scope=`）——实现阶段再定。  
- **加权** 是否与 **健康检查**（主动 ping 上游）捆绑发布——本迭代不强制，可在后续 change 引入。
