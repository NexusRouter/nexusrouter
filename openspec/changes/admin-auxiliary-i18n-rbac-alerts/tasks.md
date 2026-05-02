## 1. 国际化（前端 + 资源）

- [x] 1.1 引入 `i18next`、`react-i18next`、`i18next-browser-languagedetector`；封装初始化与 `fallbackLng: zh`
- [x] 1.2 新增 `src/locales/zh.json`、`en.json`（或 TS 模块）与命名空间；接入 Ant Design `ConfigProvider` + `dayjs` 语言同步
- [x] 1.3 顶栏或设置内语言切换器；`localStorage` 持久化语言键
- [x] 1.4 迁移现有页面：布局、登录、仪表盘、上游、密钥、调试、进阶页等文案与 `message` 调用至 `t()`；补充 vitest 或 smoke 断言关键键存在

## 2. RBAC（后端 JWT + 中间件 + 前端）

- [x] 2.1 扩展 `adminauth`：签发 JWT 含 `role`（`admin` | `operator`）；登录响应或 `/auth/me` 暴露角色供前端使用
- [x] 2.2 配置模型：支持至少一个操作员账号（环境变量或独立 bcrypt 配置，见 `design.md` 定稿）
- [x] 2.3 Gin 中间件：`RequireAdminRole()` 挂到所有写路由；维护只读 GET 白名单并加测试
- [x] 2.4 前端：`authStore` 存 `role`；`AdminLayout` 与页面按钮按角色隐藏/禁用； axios 403 统一提示（i18n）
- [x] 2.5 `go test` 覆盖操作员禁止写、允许读样例

## 3. 系统设置（API + 页面 + 文档）

- [x] 3.1 定义设置 DTO 与 `mutability` 元数据；`GET` 聚合 env + `gateway.yaml` 可读字段
- [x] 3.2 `PUT`/`PATCH`：热更新类映射到现有 `PersistSnapshot` / `reload-config` 路径；端口类仅写模板或返回 `restart_required`
- [x] 3.3 仪表盘「系统设置」页（表单 + 说明文案 i18n）；操作员只读
- [x] 3.4 更新 `services/gateway/README.md`：设置项、重启与热更新边界

## 4. 运行态告警（配置 + 评估 + UI）

- [x] 4.1 `gateway.yaml`（或快照）增加 `admin_alerts` 段与校验；默认关闭
- [x] 4.2 后台评估器：定时读取 metrics 窗口错误率，防抖后更新原子状态
- [x] 4.3 `GET /api/admin/v1/alerts/status` 实现与 Zap 日志（可选指标计数）
- [x] 4.4 管理布局告警条组件 + i18n；与 `critical` / `warning` 样式区分
- [x] 4.5 单元/集成测试：超阈值时状态切换

## 5. 质量与交付

- [x] 5.1 `go test ./...`（`services/gateway`）与 `pnpm lint`、`pnpm exec tsc --noEmit`、`pnpm test`（`web/dashboard`）通过
- [x] 5.2 自检：英文语言下主要路径无漏键（控制台无 missingKey）
