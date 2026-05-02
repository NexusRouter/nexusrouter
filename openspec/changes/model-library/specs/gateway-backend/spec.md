# gateway-backend Specification（变更 `model-library` — 增量）

## ADDED Requirements

### Requirement: 模型库管理路由注册

`services/gateway` SHALL 在管理控制台已启用且鉴权链完整时，注册 **`/api/admin/v1/model-library/**`**（或等价前缀）下的处理器，用于模型目录与绑定的 CRUD 及同步触发；路由 MUST 复用现有 **`adminJWTMiddleware`** 与写保护策略。

#### Scenario: 路由可发现

- **WHEN** 审查者检索 `RegisterAdminConsole` 或路由注册表
- **THEN** 可见模型库相关路径与 HTTP 方法映射

### Requirement: 模型库数据库迁移

模型库相关表 MUST 通过项目约定的迁移机制（GORM AutoMigrate 或 SQL 迁移）创建；迁移 MUST 在 **SQLite 与 Postgres** 目标下均可应用（与现有 `repository` 策略一致）。

#### Scenario: 全新数据库启动

- **WHEN** 进程首次连接空库并执行迁移
- **THEN** 模型库表存在且 `go test` 或冒烟脚本通过

### Requirement: 错误响应一致性

模型库处理器在失败时 MUST 返回与 **`gateway-backend`** 既有要求一致的 JSON 错误体（**`code`/`message`/`request_id`**），且 MUST 记录 Zap（含 **`request_id`**）。

#### Scenario: 客户端错误

- **WHEN** 提交非法 JSON 或违反校验
- **THEN** 响应 **400** 且 body 含稳定 **`code`**
