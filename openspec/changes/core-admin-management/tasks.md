## 1. 配置与后端骨架

- [x] 1.1 在 `internal/config` 增加管理控制台相关环境变量（启用开关、JWT 密钥/过期、管理员账号或初始哈希、可选 SMTP/重置）并在 README 文档化
- [x] 1.2 新增 `internal` 包：管理员认证（密码校验、JWT 签发/校验中间件）、与现有 Gin `router` 注册方式对齐
- [x] 1.3 定义管理 API 路由组前缀（如 `/api/admin`），与 OpenAI 与现有 `/internal/*` 不冲突；Wire 注入新依赖

## 2. 指标采集与查询 API

- [x] 2.1 在代理与错误路径挂接计数：总请求、成功/失败、延迟聚合、按 **`code`** 分桶；进程内线程安全、有界存储
- [x] 2.2 实现按日（UTC 或可配置时区）滚动计数，支持今日/昨日对比查询
- [x] 2.3 暴露 `GET` 管理端指标聚合接口，响应满足 `admin-dashboard-metrics` 与 `gateway-backend` 增量规范

## 3. 上游配置管理 API

- [x] 3.1 实现上游列表/详情读接口（来自 `runtime.Store.Snapshot()`），包含当前 **`active_upstream_id`** 与解析结果说明
- [x] 3.2 实现上游 CRUD、`default_upstream_id` 更新、`PUT` 固定 active（与现有 handler 统一鉴权模型）；校验 **`id`** 唯一与 URL 合法
- [x] 3.3 按 `design.md` 实现 YAML 原子写盘 + `Reload()`；冲突/校验失败时保留旧快照并返回明确错误

## 4. API Key 管理 API

- [x] 4.1 扩展或封装 `keystore`：支持读全量元数据、脱敏展示、单条与批量更新、写回 JSON 文件并触发重载
- [x] 4.2 实现管理端 REST：`GET` 列表（脱敏）、`POST` 新增、`PATCH` 禁用/过期、`DELETE` 删除、批量接口；符合 `admin-api-key-console` 与 `api-key-management` 增量规范
- [x] 4.3 可选：为 JSON 记录增加 `created_at` 写入策略（新建时写入 UTC 时间）

## 5. 认证与账户流程 API

- [x] 5.1 实现 `POST /login`、`POST /logout`（或等价）、令牌刷新（若采用双令牌）；错误体符合统一 JSON 约定
- [x] 5.2 实现忘记密码：无邮件时的占位说明；有邮件时一次性 token 流程（若首版裁剪则文档说明「后续迭代」）
- [x] 5.3 集成测试：错误密码、过期令牌、无令牌访问管理 API

## 6. 仪表盘前端

- [x] 6.1 增加路由：登录、仪表盘、上游配置、API Key、接口调试；未登录重定向与 Query 缓存（React Query）
- [x] 6.2 仪表盘页：卡片展示在线状态、QPS、成功率、耗时；图表或表格展示今日/昨日与错误类型分布（对接指标 API）
- [x] 6.3 上游页：表格 + 表单对话框；展示当前生效上游；保存后刷新快照展示
- [x] 6.4 API Key 页：脱敏列表、批量选择、禁用/删除确认；新增密钥仅创建时完整展示一次（若规范要求）

## 7. 内嵌 Swagger 与联调

- [x] 7.1 「接口调试」页：iframe 或 swagger-ui-react 加载同源 `/openapi.json`；处理 base URL 与鉴权头说明（文档或预设）
- [x] 7.2 与 `NEXUSROUTER_ENABLE_SWAGGER_UI` 策略对齐：开发/生产文档说明如何仅通过控制台访问
- [x] 7.3 端到端验证：登录 → 各页数据加载 → 修改上游/Key 后健康检查与样例 `chat` 请求行为符合预期

## 8. 文档与质量

- [x] 8.1 更新 `services/gateway/README.md` 与管理控制台相关环境变量、安全建议（HTTPS、令牌保管）
- [x] 8.2 后端 `go test ./...` 与前端 `pnpm lint`、`pnpm exec tsc --noEmit` 通过；必要时补充 Vitest/处理器单测
