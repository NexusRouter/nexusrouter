## 1. 配置与聚合加载

- [x] 1.1 定义 **`gateway.yaml`**（或设计最终格式）顶层 schema：upstream、cors、rate_limit、proxy_access_log、与 env 优先级字段；提供 **`gateway.yaml.example`**。
- [x] 1.2 实现 **`NEXUSROUTER_GATEWAY_CONFIG_FILE`** 加载、校验失败策略（保留旧快照或启动失败）及与现有 `config.Load` 的合并层。
- [x] 1.3 实现 **`SIGHUP` / `POST /internal/reload-config`** 热加载（与 `design.md` 一致），并写 Zap 成功/失败日志。

## 2. 上游目标管理

- [x] 2.1 将上游列表重构为带 **`id`/`weight`/`base_url`** 的结构体切片；实现 **`default_upstream_id`** 与 **加权随机** 选择器（可注入 `rand` 便于测试）。
- [x] 2.2 实现 **`active_upstream_id`** pin 模式与解除；单元测试覆盖切换后固定命中。
- [x] 2.3（可选）实现 **`PUT /internal/upstream/active`**（管理 Bearer），与配置文件双向一致策略在 README 说明。

## 3. 代理访问日志

- [x] 3.1 实现独立 **access log** `zapcore`（或等价）写文件 + lumberjack/按日滚动；字段集与 **`proxy-access-logging`** 对齐。
- [x] 3.2 实现 **`info`/`error`** 级别过滤；脱敏 **`Authorization`** / **`X-API-Key`** / **`Cookie`**。
- [x] 3.3 集成到 **ChatProxy** 完成路径（含 SSE **`duration_ms`** 定义的单测）。

## 4. 限流

- [x] 4.1 实现 **per-IP** 令牌桶（鉴权前）与 **per-key** 令牌桶（鉴权后）；超限 **429** + 统一 JSON + **`RATE_LIMITED`**（或规格枚举）。
- [x] 4.2 **`OPTIONS`** 预检计数策略按规格实现；Zap Warn 含 **`request_id`**。
- [x] 4.3（可选）**Redis** 后端：当 `NEXUSROUTER_RATE_LIMIT_BACKEND=redis` 时使用 **go-redis** 共享计数；文档化连接参数。（**本迭代已放弃**，保留为后续变更。）

## 5. CORS

- [x] 5.1 实现可配置 **CORS** 中间件（AllowOrigins/Methods/Headers/MaxAge）；默认关闭。
- [x] 5.2 集成测试：**OPTIONS** 预检与 **POST** 带 `Origin` 行为；未允许 Origin 不返回误导头。

## 6. 中间件链与路由

- [x] 6.1 在 `ProvideEngine` 按 **`design.md`** 注册：**CORS → RequestID → IP 限流 → … → Auth → Key 限流 → Register**；代码中文注释说明顺序。
- [x] 6.2 更新 **`NoRoute`**/**错误**路径不与 CORS/限流冲突；确保 **`/health`** 仍免鉴权。

## 7. 文档与契约

- [x] 7.1 更新 **`services/gateway/README.md`**：配置表、中间件顺序图、限流与 CORS 示例、管理 API。
- [x] 7.2 为 **`/internal/*`**（若新增）补充 **swag** 注释并 **`make docs`**；OpenAPI 契约测试扩展（若有新路径）。
- [x] 7.3 **`go test ./...`**、**`gofmt`**、**`openspec validate --changes`** 全绿。

## 8. 回归与发布说明

- [x] 8.1 无配置文件时行为与当前 **env-only** 部署一致性的集成测试。
- [x] 8.2 在 PR 描述或变更目录中写明 **迁移** 与 **回滚** 检查清单。
