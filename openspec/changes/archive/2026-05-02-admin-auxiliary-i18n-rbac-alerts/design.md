## Context

- 网关：`services/gateway`，配置来自环境变量与可选 `gateway.yaml`；已有管理 JWT（`adminauth`）、指标收集器、热加载 `reload-config`。
- 仪表盘：`web/dashboard`，Ant Design + React Router；文案当前以中文为主（见基线 spec）。
- 用户诉求：双语、系统设置、多角色、面板内告警。

## Goals / Non-Goals

**Goals:**

- 默认 **中文**；用户可切换到 **英文**，刷新或同会话内保持（以 localStorage + i18n 语言 detector 实现为准）。
- **系统设置**：在管理端可读写的参数集合与「是否需重启」在 UI 与 API 中明确标注；能热更新的项走现有 `PersistSnapshot` / `reload-config` 路径对齐。
- **RBAC**：`admin` 全量；`operator` 只读子集，后端强制 **403**，前端隐藏或禁用入口避免误点。
- **告警**：阈值可配置（建议 `gateway.yaml` 下 `admin_alerts` 段或等价快照）；指标来源为进程内已有 metrics（错误率、上游探测若已有则复用，否则先做错误率 + 可选上游健康占位）。

**Non-Goals:**

- 完整多租户、审批流、外部告警渠道（钉钉/PagerDuty）——可作为二期。
- 无缓存层时「缓存时长」不作为假配置写入运行路径；仅文档或 UI 标注「未启用」。

## Decisions

1. **i18n 技术栈**  
   - **选定**：`i18next` + `react-i18next` + `i18next-browser-languagedetector`；Ant Design `ConfigProvider` 的 `locale` 与 `dayjs` 语言随 `i18n.language` 同步。  
   - **备选**：仅 Ant Design 内置中英文 — 拒绝，无法覆盖业务文案与 `message`。

2. **语言资源组织**  
   - **选定**：`src/locales/zh.json`、`src/locales/en.json`（或 `.ts` 导出对象），按页面或域分命名空间（`common`、`layout`、`pages.upstreams`）。  
   - **默认**：`fallbackLng: 'zh'`。

3. **角色来源**  
   - **选定**：登录时在 JWT claims 写入 **`role`**（字符串枚举：`admin` | `operator`）；`operator` 账号通过环境变量或配置文件静态声明（如 `NEXUSROUTER_ADMIN_OPERATORS` 为 bcrypt 列表或独立用户名表），首版避免引入 DB。  
   - **备选**：独立用户表 — 二期。

4. **系统设置与端口**  
   - **选定**：设置 API 返回「当前生效值」与「配置文件中的值」；**监听端口** 仅展示与校验，写入 `.env.example` / 文档化模板；实际改端口需 **重启进程**（UI 明确提示）。**请求超时** 映射现有 `NEXUSROUTER_UPSTREAM_TIMEOUT` 等；**日志路径** 映射 `proxy_access_log.path` 或应用日志路径（以现有 config 字段为准）。  
   - **缓存时长**：若代码中无独立 TTL，设置页显示只读「未实现」或关联未来 Redis 配置键 — 在实现阶段二选一并在 spec 中固定。

5. **告警计算**  
   - **选定**：后台定时（如每 30s）从 `metrics.Collector` 拉窗口内错误率，与 `admin_alerts.error_rate_threshold`（百分比）比较；上游不可用若有健康检查则并入，否则 **ADDED** 占位接口返回 `degraded` 供前端条展示。  
   - **Mitigation**：误报 — 支持最小持续时间（如连续 2 个周期超阈值才告警）。

## Risks / Trade-offs

- **[Risk] 修改端口导致锁死** → **Mitigation**：写端口前二次确认；文档强调 systemd/k8s 滚动重启。  
- **[Risk] RBAC 仅前端隐藏可被绕过** → **Mitigation**：所有写接口 Gin 中间件校验 `role`。  
- **[Risk] i18n 漏键** → **Mitigation**：开发脚本或 ESLint 规则扫描 `t('` 键；E2E 抽样英文化页面。

## Migration Plan

1. 发布后端（新 JWT claims、中间件、可选 YAML 段）→ 发布前端。  
2. 现有单管理员：默认 `role=admin`，行为不变。  
3. 新增操作员：运维配置第二个用户名/密码或 env 列表后登录验证只读。  
4. 回滚：移除 RBAC 中间件与 claims（feature flag 可选）。

## Open Questions

- 操作员账号是否允许「自助改密码」——首版可 **否**，仅管理员重置。  
- 告警是否需 **邮件** — 本变更 **否**，仅面板内视觉告警。
