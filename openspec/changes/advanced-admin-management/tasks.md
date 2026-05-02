## 1. 运行时模型与配置 Schema

- [x] 1.1 扩展 `gateway.yaml` / `runtime.Snapshot`：新增 `rate_limit_rules`、`ip_access` 等字段；`validateSnapshot` 与合并逻辑向后兼容
- [x] 1.2 文档化 YAML 示例与字段说明（README / `gateway.yaml.example`）

## 2. IP 名单与中间件

- [x] 2.1 实现 CIDR 匹配工具（IPv4 必保；IPv6 视依赖可选）与单元测试
- [x] 2.2 在 Gin 链插入 IP 名单检查（顺序见 `design.md`）；403/错误码与 JSON 体符合 `gateway-backend` 约定
- [x] 2.3 管理 API：`GET/PUT /api/admin/v1/security/ip-access`（命名以实现为准）与批量 `PATCH`

## 3. 限流规则引擎

- [x] 3.1 将现 `IPRateLimit`/`KeyRateLimit` 重构为可组合：全局回退 + 规则表匹配（priority + path_prefix + dimension）
- [x] 3.2 管理 API：规则 CRUD 与校验；写盘 + `Reload()` 与乐观锁（可选 `etag`）

## 4. CORS 管理 API

- [x] 4.1 `GET/PUT` CORS 段；支持 bulk origins 解析；与 `DynamicCORS` 读快照一致
- [x] 4.2 集成测试：保存后 OPTIONS 预检头变化可观测

## 5. 日志查询与导出

- [x] 5.1 实现日志文件发现（当前与 lumberjack 轮转文件）与逆序扫描过滤器
- [x] 5.2 `GET /api/admin/v1/logs/query` 分页 JSON；`GET /api/admin/v1/logs/export.csv` 流式 CSV
- [x] 5.3 硬上限与超时保护；Zap/指标记录慢查询

## 6. 仪表盘前端

- [x] 6.1 新增路由与侧栏：日志、限流规则、CORS、IP 名单
- [x] 6.2 日志页：时间范围、路径、状态、指纹、IP 筛选 + 表格 + 导出链接
- [x] 6.3 限流规则页：可编辑表格 + 保存
- [x] 6.4 CORS 页：methods/headers + bulk origins
- [x] 6.5 IP 名单页：mode 切换 + bulk CIDR + 列表删除

## 7. 质量与文档

- [x] 7.1 `go test ./...`、`pnpm lint`、`pnpm exec tsc --noEmit`、`pnpm test`
- [x] 7.2 更新 `services/gateway/README.md` 进阶管理章节与安全注意事项
