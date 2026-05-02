## 1. 词表与 i18n 基线

- [x] 1.1 在 `web/dashboard/src/locales/zh.ts` 与 `en.ts` 增加 `consoleTerms.*`（或 `glossary.*`）前缀，覆盖侧栏域标题、菜单项及跨页核心术语（与 `admin-console-global-ux`、`admin-i18n` delta 一致）
- [x] 1.2 将 `AdminLayout` 侧栏域分组与菜单项文案改为引用上述 i18n 键，移除终态硬编码字符串

## 2. 壳层：侧栏域分组与导航

- [x] 2.1 在 `AdminLayout.tsx` 实现域分组 UI（桌面侧栏 + 移动抽屉共用数据源），分组顺序与 `design.md` 决策一致
- [x] 2.2 校验所有路由的选中高亮逻辑（含 `/gateway/*`、`/model-library` 子路径），补必要的 `startsWith` 规则
- [x] 2.3 手测：登录 → 各一级入口可达，书签 `/gateway/policy#section-rate-limits` 仍可用

## 3. 共享组件：危险操作与空错载

- [x] 3.1 新增 `confirmDestructive`（或等价）封装：展示资源名/ID、不可恢复说明、`okType="danger"`、异步 `onOk` 与 loading
- [x] 3.2 新增可复用 `PageEmpty`、`PageError`（或扩展现有模式）：空态含下一步引导链接；错误含重试
- [x] 3.3 在 Story/文档注释或 `README` 片段中说明三级强度（L1/L2）使用场景（可选，不强制 Storybook）

## 4. 各页接入删除确认与空错载

- [x] 4.1 `ApiKeys.tsx`：删除/吊销类操作接入统一确认；空列表与错误态走基线组件
- [x] 4.2 `Upstreams.tsx`：删除上游接入统一确认；空错载基线
- [x] 4.3 `GatewayPolicyPage`（及页内子模块，若仍拆分文件）：删除规则等接入统一确认；与 `gateway-policy-page-ux` 已有布局不冲突
- [x] 4.4 `ModelLibrary.tsx`：各实体删除由链式 danger 改为带确认对话框，展示实体标识
- [x] 4.5 `AccessLogs.tsx`、`Dashboard.tsx`：加载/错误/空数据与全局基线一致（日志无数据时说明筛选条件）
- [x] 4.6 `SystemSettings.tsx`：Operator 禁用态补充 Tooltip/Alert（与 `admin-rbac` 一致），不新增 API

## 5. 跨页引导与权限文案

- [x] 5.1 在关键空态增加指向「上游」「API Key」「模型库」等的 `Link`（仅当规格场景适用）
- [x] 5.2 统一只读/无权限时的短说明文案键，避免「点了没反应」

## 6. 验证与收尾

- [x] 6.1 `pnpm exec tsc --noEmit`、`pnpm lint`（`web/dashboard`）通过
- [x] 6.2 中英切换全路径烟雾测试；确认无新增硬编码用户文案
- [x] 6.3 对照 `openspec/changes/console-global-ui-ux-refactor/specs/**/*.md` 做一次场景自检清单（可在 PR 描述勾选）
