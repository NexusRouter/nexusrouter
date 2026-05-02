## ADDED Requirements

### Requirement: 管理控制台路由与导航

`web/dashboard` MUST 提供管理控制台相关页面路由，至少包括：**登录**、**仪表盘**、**上游/代理配置**、**API Key 管理**、**接口调试（内嵌 Swagger）**。MUST 提供清晰导航入口；受保护页面 MUST 与 **`admin-auth`** 规范一致进行访问控制。

#### Scenario: 从根路径可发现登录或首页

- **WHEN** 用户打开仪表盘应用根路径
- **THEN** 可导航至登录页或（已登录）进入默认管理首页，且无死链

### Requirement: 管理功能与依赖矩阵兼容

新增页面 MUST 继续使用 **`dashboard-frontend`** 已锁定的技术栈（React、Ant Design、TanStack Query、react-router 等），且 MUST 通过 `pnpm exec tsc --noEmit` 在干净基线下可类型检查（无新增业务类型错误）。

#### Scenario: Lint 基线不回归

- **WHEN** 开发者执行项目既有 `pnpm lint`（或等价脚本）
- **THEN** 在无既有债务的前提下以成功状态退出，或新增之豁免 MUST 在变更内说明范围
