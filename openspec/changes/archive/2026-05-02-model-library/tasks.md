## 1. 数据层与迁移

- [x] 1.1 定义 `model_catalog_entry`、`model_upstream_binding` GORM 模型与仓库方法（CRUD、按启用状态列表）
- [x] 1.2 接入现有迁移/AutoMigrate 路径，并在 SQLite 与 Postgres 下验证
- [x] 1.3 为仓库层编写表驱动单元测试（含唯一约束与软删除策略若采用）

## 2. 管理端 API

- [x] 2.1 在 `RegisterAdminConsole` 中注册 `/api/admin/v1/model-library/*` 路由（列表、条目、绑定、同步）
- [x] 2.2 实现同步：HTTP 客户端 GET 上游 `/v1/models`，超时与 body 上限，错误映射为统一 JSON
- [x] 2.3 管理端集成测试：JWT 鉴权、401/403、成功路径

## 3. 公开 GET /v1/models

- [x] 3.1 在 `router` 注册 GET `/v1/models`（及可选 GET `/v1/models/:model`），挂载 `GatewayAuth` 链
- [x] 3.2 实现聚合查询：启用绑定 ∩ 快照中存在 upstream_id → OpenAI 形状 JSON
- [x] 3.3 `go test` 覆盖空列表、单模型、未知模型 404

## 4. OpenAPI 与文档

- [x] 4.1 更新 swag 注释与 `internal/openapi` 嵌入 YAML（含 GET `/v1/models` 与安全方案）
- [x] 4.2 `services/gateway/README.md` 补充模型库与环境变量（不出现外部项目名）

## 5. 控制台

- [x] 5.1 新增 `ModelLibrary` 页面：表格、筛选、编辑抽屉、绑定上游选择（数据源来自现有 snapshot API）
- [x] 5.2 `App.tsx` 路由与 `AdminLayout` 导航、`zh`/`en` 文案
- [x] 5.3 `api.ts` 封装与 React Query 查询/变更；`pnpm test` / `tsc` 通过

## 6. 验收与合并

- [x] 6.1 按 `specs/**` 场景编写或更新 E2E/集成测试清单并执行
- [x] 6.2 代码审查：确认用户文档与 spec 中无第三方开源项目名或「对标某某」表述

## 7. Chat 请求体 `model` 改写（首版必做）

- [x] 7.1 更新 `design.md` / `proposal.md` / `specs/**`：将改写从「后续」提升为必做
- [x] 7.2 实现 `repository.RewriteChatCompletionsModelBody`，`ChatProxy` 注入 `*gorm.DB` 并在转发前改写
- [x] 7.3 单测：仓库 + `TestChatProxy_RewritesModelFromModelLibraryBinding`；README 与控制台文案补充
