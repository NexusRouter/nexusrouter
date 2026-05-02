## 1. 路由与布局骨架

- [x] 1.1 在 `App.tsx` 中增加网关策略统一路由（如 `/gateway/policy` 或调整 `/gateway` 子布局），保留旧路径重定向至带 section/hash 的新入口
- [x] 1.2 重构 `GatewayPolicyLayout`：移除或与单页场景布局合并，提供三场景 `Card` 容器及可选锚点/Sticky 子导航
- [x] 1.3 确认侧栏 `menuGatewayPolicy` 高亮与当前路由、深链一致

## 2. 组件抽取与区块复用

- [x] 2.1 将 `IpAccess.tsx`、`RateLimitRules.tsx`、`CorsSettings.tsx` 中表单/表格主体抽取为可复用组件（或保持页面内联但统一外层标题、间距、`max-w` 与底部操作条）
- [x] 2.2 统一三区块底部操作区：主按钮「保存」、`persist`/写回 YAML 开关与说明文案位置一致
- [x] 2.3 对齐限流/CORS/IP 各区块的 `persist` 行为（若 API 支持则三处均可控；若某处曾写死，改为显式开关且默认与线上一致）

## 3. 文案与 i18n

- [x] 3.1 在 `zh.ts` / `en.ts` 增加场景化标题、副标题、术语对照与 Tooltip 文案键（覆盖 allowlist/denylist、维度、PATCH 等）
- [x] 3.2 替换界面中仅以技术枚举为标签的 `Select.Option`，改为「业务标签 + 值」或 helper 文本

## 4. CORS 分层与表单校验

- [x] 4.1 实现 CORS 基础/高级折叠区：默认折叠方法与请求头等高级项
- [x] 4.2 为 IP/CIDR、路径前缀、RPS/burst、origins 等添加 antd Form rules 或 Zod 前置校验，与 API 错误展示格式统一
- [x] 4.3 实现限流规则重复/冲突的前端警告（与后端校验策略一致）

## 5. 风险披露与只读

- [x] 5.1 为写回 `gateway.yaml`、IP 模式切换等增加 Alert/Modal/次要说明文案
- [x] 5.2 验证 Operator 只读在三区块均生效且保存与写回均禁用

## 6. 验证

- [x] 6.1 手测：中英文切换、三区块保存成功/失败、旧 URL 重定向、无控制台错误
- [x] 6.2 运行 `pnpm exec tsc --noEmit` 与项目 lint（`web/dashboard`）
