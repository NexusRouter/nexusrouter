## 1. 配置与依赖

- [x] 1.1 在 `services/gateway/go.mod` 增加 **`gorm.io/driver/sqlite`**（及传递依赖），运行 `go mod tidy`，确认 **`go build ./...`** 通过
- [x] 1.2 在 `internal/config` 增加 DSN 与 SQLite 文件路径相关字段（如 **`NEXUSROUTER_DATABASE_URL`** 空则 SQLite；**`NEXUSROUTER_SQLITE_PATH`** 或等价键覆盖默认 **`gateway.db`**），并在 README 中说明

## 2. 数据库引导与模型

- [x] 2.1 实现 **`internal/repository`**（或子包）中的 GORM 打开逻辑：按 DSN 选择 **postgres** / **sqlite** Dialector，配置合理 **`gorm.Config`** 与日志
- [x] 2.2 定义网关快照、API Key、管理员（若迁入表）等 **GORM 模型**；在启动路径调用 **`AutoMigrate`**
- [x] 2.3 将 DB 构造纳入 **Wire**（`internal/provider`），衔接现有依赖注入并清理仅用于钉版的空白导入（在确有业务引用后）

## 3. 领域迁移与真源切换

- [x] 3.1 实现启动时 **「DB 为空则从 `gateway.yaml` / keys JSON 导入」** 逻辑，并写结构化 Zap 日志；明确空库判定与冲突策略（DB 已有数据则跳过导入）
- [x] 3.2 将 **`runtime.Store`** 持久化从仅 YAML **`PersistSnapshot`** 改为写 DB；读路径以 DB 为真源
- [x] 3.3 将 **`keystore.Store`** 从 JSON 文件改为 DB 读写；实现 **`SIGHUP`** / **`POST /internal/reload-keys`** 从 DB 刷新内存视图

## 4. 管理员与安全

- [x] 4.1 定稿并实现 **管理员凭据** 策略（纯 DB 表 vs DB+env 回退），与 **`adminauth`** 集成；禁止在日志中输出密码或完整 secret

## 5. 验证与文档

- [x] 5.1 为仓储与导入路径补充 **单元测试**（可使用内存 SQLite 或临时文件）
- [x] 5.2 更新 **`services/gateway/README.md`** 与 **`openspec/project.md`**（若提及持久化）：默认 **`gateway.db`**、Postgres DSN 切换、升级自 YAML/JSON 的步骤与回滚注意
- [x] 5.3 全量 **`go test ./...`** 与手动冒烟：无 DSN 启动、导入、管理 API 改配、Key 鉴权、切 Postgres（可选 CI 矩阵或文档化手工步骤）
