## Context

网关当前将 **运行时快照**（上游、路由、CORS、限流等）持久化到 **`gateway.yaml`**，**API Key** 到 **JSON 文件**，**管理员**到 **环境变量 bcrypt**。`go.mod` 已引入 GORM 与 Postgres 驱动，但 `internal/repository` 仅占位，`internal/deps` 以空白导入钉版本。目标是在不引入第二套 ORM 的前提下，用 **GORM + 默认可写 SQLite 文件** 实现零配置持久化，并通过 **DSN** 切换 Postgres，与远期规划一致。

## Goals / Non-Goals

**Goals:**

- 未配置显式数据库 DSN 时，使用 **`gateway.db`**（SQLite，路径可配置但默认值固定为工作目录或数据目录下的该文件名，见实现任务）作为唯一结构化存储；启动时 **`AutoMigrate`** 创建/演进表结构。
- 配置 **Postgres DSN**（单一环境变量，如 `NEXUSROUTER_DATABASE_URL`，具体键名在实现中固定并文档化）时，使用同一套 GORM 模型与仓储接口，**仅更换 Dialector**。
- 将现有「需持久化的」域数据（网关配置快照、API Key 行、管理员账号行等，与 `tasks.md` 拆解一致）纳入 DB；对外 HTTP 行为保持兼容（管理 API、鉴权语义不变）。
- 提供从 **YAML/JSON** 到 DB 的 **一次性或启动时导入** 路径，避免静默丢配置。

**Non-Goals:**

- 手写 SQL 迁移文件或强制引入 `golang-migrate` 驱动的版本化迁移流水线（本变更以 **AutoMigrate** 为准；若后续需要可另起变更）。
- 多主复制、读写分离、连接池高级调优（Postgres 模式下采用 GORM 默认合理配置即可）。
- 改变上游 OpenAI 兼容协议或 Dashboard 的 REST 契约（除非为配合持久化字段所必需的最小调整）。

## Decisions

1. **默认 SQLite、可选 Postgres**  
   - **决策**：通过 **是否设置非空 DSN** 选择驱动：`sqlite` + 文件路径 vs `postgres` + URL。  
   - **理由**：与用户需求「无配置 → SQLite」「有配置 → Postgres、业务代码不变」一致。  
   - **备选**：始终要求外部 Postgres——违背零配置；双配置文件——增加运维面。

2. **单一 GORM `*gorm.DB` 注入**  
   - **决策**：Wire 注入一个由配置工厂打开的 DB；Repository 层只依赖接口或 `*gorm.DB`。  
   - **理由**：切换 DSN 时不动 handler 逻辑。  
   - **备选**：运行时双实现（文件 + DB）长期并存——复杂度高，仅在迁移窗口短期保留。

3. **AutoMigrate 而非 migrate CLI**  
   - **决策**：启动时 `AutoMigrate` 全部领域模型。  
   - **理由**：用户明确要求不写迁移文件；表结构随模型演进。  
   - **风险**：生产大表改列需后续引入显式迁移策略（记入 Non-Goals / Open Questions）。

4. **旧文件与新 DB 的优先级**  
   - **决策**：首次启动若 DB 为空且检测到已存在的 `gateway.yaml` / keys JSON，则 **导入** 到 DB 并记录日志；导入成功后 MAY 将「仅 DB」作为真源（是否重命名/备份旧文件在实现中二选一，须在 README 写明）。若 DB 已有数据，则以 DB 为准。  
   - **理由**：避免升级丢数据。  
   - **备选**：永远双写——复杂且易不一致。

5. **SQLite 驱动**  
   - **决策**：使用官方 **`gorm.io/driver/sqlite`**（底层 `modernc.org/sqlite` 或 `mattn/go-sqlite3` 以模块传递为准），与现有 GORM 版本兼容。  
   - **理由**：与 Postgres 驱动同属 GORM 生态，技术栈单一。

6. **`SIGHUP` / reload-keys**  
   - **决策**：在 DB 模式下，热加载语义为 **从 DB 重新加载密钥（及若适用之其他可热载配置）到进程内缓存**，而非读 JSON 文件（除非保留显式「仅文件」降级模式；若保留须在 spec 与 README 标清）。默认实现以 **DB 为真源**。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| AutoMigrate 在生产 Postgres 上锁表或慢 | 首次部署在小数据量网关场景可接受；文档说明低峰变更；后续可加 migrate 变更。 |
| SQLite 并发写 | 网关写多为管理路径；GORM 单连接串行化写；高并发写场景文档推荐 Postgres。 |
| 与现有「仅文件」部署习惯冲突 | README 给出迁移步骤与环境变量对照表；可选保留只读导入期。 |
| `gateway.db` 路径与容器只读根 | 支持环境变量覆盖 DB 文件路径；文档示例挂载 volume。 |

## Migration Plan

1. 发版说明：新增默认 `gateway.db`、可选 `DATABASE_URL`；列出从 YAML/JSON 导入行为。  
2. 升级：备份原 `gateway.yaml` 与 keys JSON；启动新版本一次完成导入；验证管理台与鉴权。  
3. 回滚：保留旧版本镜像与文件备份；若已写 DB，回滚前导出或接受以文件旧快照恢复（在 README 说明限制）。  
4. 切换 Postgres：设置 DSN、空库或预先 `AutoMigrate`、数据迁移工具（若需）由运维执行；应用仅换配置。

## Open Questions

- 管理员凭据是否完全迁入表（含 bcrypt）并弃用纯 env，还是 DB 与 env 组合策略——需在实现前在 `tasks.md` 中定稿并与安全评审对齐。  
- `gateway.yaml` 是否在 DB 模式后仍支持「导出/导入」作为运维工具格式（推荐保留导出 API 或 CLI 以降低锁定感）。
