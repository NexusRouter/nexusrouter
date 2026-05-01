## 1. 配置与数据结构

- [x] 1.1 在 `internal/config` 增加 **`NEXUSROUTER_UPSTREAM_BASE_URLS`**（逗号分隔）解析，并与现有 **`NEXUSROUTER_UPSTREAM_BASE_URL`** 按 `design.md` 约定合并为 **`[]string` 上游列表**（列表为空时回退单键）。
- [x] 1.2 增加 **`NEXUSROUTER_GATEWAY_KEYS_FILE`**、可选 **`NEXUSROUTER_ADMIN_RELOAD_TOKEN`** 配置项，并在 `services/gateway/README.md` 用表格文档化全部相关环境变量。
- [x] 1.3 定义密钥记录结构体（`id`、`secret`、`disabled`、`expires_at`）与加载器接口，单元测试覆盖 JSON 样例解析与字段默认值。

## 2. 健康检查

- [x] 2.1 实现专用 **`GET /health`** 处理器：返回 **`status`**、**`version`**（`-ldflags` 注入变量，缺省为 **`dev`**）、**`server_time`**（UTC RFC3339Nano）。
- [x] 2.2 在 `router.Register` 中确保 **`/health`** 不挂载 API Key 鉴权；为 `internal/handler` 或等价位置补充 **中文** 包/函数注释。
- [x] 2.3 添加 `httptest` 测试：断言 **200**、JSON 三字段存在且 **`server_time`** 可 `time.Parse`。

## 3. API Key 存储与热加载

- [x] 3.1 实现启动时从 JSON 文件加载密钥列表；文件权限与错误路径按 `design.md` 固定策略（启动失败或拒绝受保护路由）并写 Zap。
- [x] 3.2 实现 **`SIGHUP`** 监听与原子替换内存中的密钥切片（或等价无锁快照）；在 macOS/Linux 上用手动测试步骤或集成测试验证（若 CI 限制则文档化本地步骤）。
- [x] 3.3（可选）若实现 **`POST /internal/reload-keys`**：校验 **`NEXUSROUTER_ADMIN_RELOAD_TOKEN`**（如 **`Authorization: Bearer <token>`**），失败 **401**，成功则重载并 **204/200** JSON。

## 4. 鉴权中间件

- [x] 4.1 用新存储替换/封装现有 **`GatewayAuth`**：仅 **`Authorization: Bearer`** 与密钥 **`secret`** 比对；处理禁用、过期、缺失、格式错误，一律 **401** 且 body 走统一错误封装（含 **`request_id`**）。
- [x] 4.2 决定是否保留 **`X-API-Key`** 兼容层；若保留须在 README 标注 **deprecated** 并加回归测试。
- [x] 4.3 将 **`POST /v1/chat/completions`** 中间件顺序写死为：**RequestID →（Recovery）→ 鉴权 → ChatProxy**（与现有 `ZapRecovery`/`ErrorJSON` 协调），并在代码注释说明。

## 5. 多上游反向代理

- [x] 5.1 在 Chat 代理构造路径注入 **round-robin** 选择器（`atomic` 索引），将请求转发到列表中下一上游基址。
- [x] 5.2 按 `design.md` 实现请求头复制与 hop-by-hop 剔除；**`Host`** 设为目标上游；验证 **`Content-Type`/`Accept`/`User-Agent`/自定义 `X-`** 等到达测试上游（可用 `httptest.Server` 双实例）。
- [x] 5.3 保持上游 **状态码与 body 原样** 回传；为连接失败/超时保留 **502/504** 与统一 JSON（含 **`request_id`**）。

## 6. 统一错误与 Panic 响应

- [x] 6.1 抽取 **`WriteGatewayError(c, status, code, message)`**（或等价），统一写入 **`code`/`message`/`request_id`**，且保证 **`X-Request-ID`** 头已设置。
- [x] 6.2 更新 **`ZapRecovery`**、**`ErrorJSON`**、**`NoRoute`**、鉴权失败路径、**405** 分支等所有网关自产错误调用点迁移到统一封装。
- [x] 6.3 单元或集成测试：无 **`X-Request-ID`** 时错误 JSON 与响应头同时出现且一致；带客户端 **`X-Request-ID`** 时保持一致。

## 7. OpenAPI 与 Swagger

- [x] 7.1 为 **`/health`**、**`/v1/chat/completions`**（及若存在的 **`/internal/reload-keys`**）补充 **swag** 注释（`main.go` 或 handler 旁）。
- [x] 7.2 运行 **`swag init`**（或项目 Makefile 目标）更新 `docs` 产物；断言 **`openapi`** 字段以 **`3.0.`** 开头。
- [x] 7.3 扩展现有 OpenAPI 测试包：检索 **`paths./health.get`** 与 Bearer **`security`** 仍与 Chat 操作一致。

## 8. 收尾与验证

- [x] 8.1 **`go test ./...`** 与 **`go build ./...`** 在 `services/gateway` 模块零错误通过。
- [x] 8.2 更新 **`services/gateway/README.md`**：示例 **`keys.json`**、探针建议、多上游示例、**`SIGHUP`** 重载说明。
- [x] 8.3 本变更目录下 **`openspec validate`**（若 CLI 提供）或按项目约定执行规格校验；准备归档前自检 **tasks** 勾选与实现一致。
