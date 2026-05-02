## Why

网关策略相关能力（限流、CORS、IP 访问控制）已在 `web/dashboard` 分路由实现，但当前以**后端配置维度**（限流 / CORS / IP 名单）拆成顶栏标签，与管理员「防刷、跨域、IP 管控」的心智模型不一致；界面充斥技术术语且各模块保存与校验不一致，高级选项无分层，写回 `gateway.yaml` 的风险对用户不透明。需在**不改变网关运行时与配置 schema** 的前提下，统一做一次前端信息架构与交互重构，降低误操作与理解成本。

## What Changes

- 将网关策略区从「三标签页」调整为**单页或同页滚动**下的**三大场景卡片**：IP 访问控制、限流规则、跨域配置（顺序与命名对齐业务语言），子路由可保留为锚点/深链，但首屏结构场景化。
- **术语与文案**：技术词（鉴权前/后、denylist、PATCH 等）改为业务说明 + 极简 helper；中英文 i18n 同步更新。
- **交互统一**：各模块在「保存 / 新增 / 删除 / 写回 YAML」上采用一致模式（按钮位置、成功/失败反馈、禁用态、加载态）；客户端与服务端校验错误展示一致。
- **CORS 等模块**：默认「基础模式」（开关、常用域名批量导入等）；「高级模式」折叠或分步展示方法与请求头等细节。
- **风险与校验**：IP/CIDR、域名、路径前缀、重复规则等做前端格式与冲突提示；对「写回 gateway.yaml」及重载后果提供简短说明与错误时的配置保留说明。
- **非目标**：不修改 Go 网关对 `gateway.yaml` 的解析与限流/CORS/IP 逻辑；不新增或删减管理 API 语义（仅消费现有 API）。

## Capabilities

### New Capabilities

- `gateway-policy-page-ux`：定义网关策略控制台（IP 访问控制 / 限流规则 / 跨域配置）的场景化页面流程、交互规则、术语与分层展示、校验与风险提示；与现有 `admin-cors-management`、`admin-rate-limit-policy`、`admin-ip-access-control` 的后端行为一致，仅约束前端呈现与操作体验。

### Modified Capabilities

- （无）后端配置语义与既有 admin-* 规格中的运行时要求不变；本变更通过新增 `gateway-policy-page-ux` 规格承载前端体验要求，避免与后端规格混淆。

## Impact

- **前端**：`web/dashboard/src/layouts/GatewayPolicyLayout.tsx`、`web/dashboard/src/pages/` 下与 `/gateway/*` 相关的页面与组件（如 `RateLimitRules.tsx`、`CorsSettings.tsx`、IP 访问页）、`web/dashboard/src/locales/zh.ts` 与 `en.ts`、`App.tsx` 路由结构。
- **后端**：无代码变更预期；仍依赖现有管理 API 与热更新/写盘流程。
- **配置**：继续读写既有 `gateway.yaml` 字段；兼容现有部署与升级路径。
