# dashboard-frontend Specification

## Purpose
TBD - created by archiving change project-initialization. Update Purpose after archive.
## Requirements
### Requirement: 运行时与包管理器版本

仪表盘工程 MUST 使用 **Node.js ≥ 22.x（LTS）** 与 **pnpm 9.x** 作为唯一包管理器；MUST NOT 在 `web/dashboard` 使用 npm/yarn 作为默认安装方式。

#### Scenario: 本地安装依赖

- **WHEN** 开发者在 `web/dashboard` 执行 `pnpm install`
- **THEN** 安装成功且锁文件（`pnpm-lock.yaml`）被提交并可复现构建

### Requirement: 核心依赖版本矩阵

`web/dashboard` MUST 在 `package.json` 中声明并满足以下依赖的兼容版本范围：**TypeScript ^5.7.0**、**react / react-dom ^19.1.0**、**vite ^6.3.0**、**antd ^6.0.0**、**tailwindcss ^4.1.0**、**@tailwindcss/vite ^4.1.0**、**zustand ^5.0.0**、**@tanstack/react-query ^5.74.0**、**axios ^1.8.0**、**react-router ^7.5.0**、**react-hook-form ^7.55.0**、**zod ^3.24.0**、**lucide-react ^0.487.0**、**dayjs ^1.11.0**；开发依赖 MUST 包含 **eslint ^9.x**、**eslint-plugin-react**、**eslint-plugin-react-hooks**、**prettier ^3.5.0**、**vitest ^3.1.0**、**@testing-library/react ^16.3.0**。

#### Scenario: 类型检查通过

- **WHEN** 开发者执行 `pnpm exec tsc --noEmit`（或项目脚本等价物）
- **THEN** 在无业务代码错误的前提下以零错误退出

### Requirement: Tailwind CSS v4 与 Vite 集成

前端 MUST 使用 **Tailwind CSS v4**，并通过 `**@tailwindcss/vite`** 注册到 Vite；入口样式文件 MUST 包含 `**@import "tailwindcss";`**。

#### Scenario: 开发服务器加载样式

- **WHEN** 开发者执行 `pnpm dev` 并打开应用根页面
- **THEN** Tailwind 实用类在页面中生效（例如测试用 `className` 可见样式）

### Requirement: 源码目录约定

`web/dashboard/src/` MUST 存在以下目录（可为占位）：`**components/`**、`**pages/`**、`**stores/**`、`**services/**`、`**utils/**`。

#### Scenario: 目录存在性检查

- **WHEN** 审查者列出 `src/` 子目录
- **THEN** 上述五个目录均存在

### Requirement: ESLint 与 Prettier 基线

工程 MUST 提供可运行的 **ESLint** 与 **Prettier** 配置文件，并与 TypeScript + React 源码兼容。

#### Scenario: 静态检查命令

- **WHEN** 开发者执行项目定义的 lint 与 format 检查命令（如 `pnpm lint` / `pnpm format`）
- **THEN** 在干净基线下以成功状态退出（允许后续规则收紧）

### Requirement: 文档与提交语言

仪表盘源码中的 **注释**与 **Git commit message** MUST 使用 **中文**；变量与函数名 MUST 使用 **英文**。**用户可见界面文案** MUST 通过国际化资源提供 **中文（默认）** 与 **英文** 两种语言，且 MUST 提供语言切换入口；键名与资源文件组织 MUST 符合 `admin-i18n` 能力规范。

#### Scenario: 代码审查抽样

- **WHEN** 审查者打开主要入口与配置文件的注释
- **THEN** 注释为中文且标识符为英文

#### Scenario: 默认语言为中文

- **WHEN** 用户首次打开仪表盘且无持久化语言偏好
- **THEN** 用户可见文案为中文

#### Scenario: 切换为英文

- **WHEN** 用户选择 English
- **THEN** 用户可见文案切换为英文

### Requirement: 模型库页面与导航

`web/dashboard` SHALL 提供 **`/model-library`**（或实现选定路径）路由页面，并在桌面侧栏导航中增加一项（图标 + 文案 i18n）；页面 MUST 使用 **Ant Design** 与 **TanStack Query** 与现有管理页一致。

#### Scenario: 导航可达

- **WHEN** 已登录管理员打开侧栏并点击模型库
- **THEN** 路由切换至模型库页面且无控制台报错

### Requirement: 国际化

模型库页面所有用户可见文案 MUST 具备 **中文与英文** 键（`pages.modelLibrary.*` 或等价前缀），并符合 **`dashboard-frontend`** 注释语言约定。

#### Scenario: 切换语言

- **WHEN** 用户在顶栏切换语言
- **THEN** 模型库标题与表格列头随之切换

### Requirement: 可访问性

主要操作按钮（查询、保存、同步）MUST 具备可聚焦与 **`aria-label` 或可见文本**；表格 MUST 提供合理 **`rowKey`**。

#### Scenario: 键盘可达

- **WHEN** 使用 Tab 聚焦主要按钮
- **THEN** 可触发且焦点环可见（浏览器默认或主题）

### Requirement: 初始化向导路由

`web/dashboard` MUST 提供独立路由（如 **`/setup`**）承载**首次初始化向导**页面，页面文案与校验 MUST 使用**中文**为主（与现有规范一致）。该路由在未初始化阶段 MUST 可匿名访问。

#### Scenario: 未初始化可打开向导

- **WHEN** 全局未初始化且用户直接访问向导路径
- **THEN** 页面渲染成功且无无限重定向

### Requirement: 全局路由守卫与初始化后禁止访问向导

应用根布局或路由守卫 MUST 在启动或导航时拉取后端初始化状态；当 **`initialized=false`** 时，除向导页与状态轮询所需资源外，访问其他业务路由 MUST **重定向**至向导页。当 **`initialized=true`** 时，访问向导路径 MUST **重定向**至登录页（或已存在之认证入口路径）。

#### Scenario: 未初始化强制进入向导

- **WHEN** **`initialized=false`** 且用户访问非向导内部路径
- **THEN** 浏览器地址栏最终落在向导路径

#### Scenario: 已初始化禁止向导

- **WHEN** **`initialized=true`** 且用户访问向导路径
- **THEN** 浏览器被重定向至登录页

### Requirement: 管理控制台路由与导航

`web/dashboard` MUST 提供管理控制台相关页面路由，至少包括：**登录**、**仪表盘**、**上游/代理配置**、**API Key 管理**、**模型库**、**网关策略**（限流 / CORS / IP）、**访问日志**、**系统设置**。MUST NOT 提供内嵌 Swagger / OpenAPI「Try it out」类接口调试页；MUST 提供清晰导航入口；受保护页面 MUST 与 **`admin-auth`** 规范一致进行访问控制。

#### Scenario: 从根路径可发现登录或首页

- **WHEN** 用户打开仪表盘应用根路径
- **THEN** 可导航至登录页或（已登录）进入默认管理首页，且无死链

### Requirement: 管理功能与依赖矩阵兼容

新增页面 MUST 继续使用 **`dashboard-frontend`** 已锁定的技术栈（React、Ant Design、TanStack Query、react-router 等），且 MUST 通过 `pnpm exec tsc --noEmit` 在干净基线下可类型检查（无新增业务类型错误）。

#### Scenario: Lint 基线不回归

- **WHEN** 开发者执行项目既有 `pnpm lint`（或等价脚本）
- **THEN** 在无既有债务的前提下以成功状态退出，或新增之豁免 MUST 在变更内说明范围

### Requirement: 进阶管理页面与导航

`web/dashboard` MUST 在已登录管理布局中提供入口与路由页面：**日志查询**（筛选表单、表格、导出按钮）、**限流规则**（表格与编辑抽屉/弹窗）、**跨域配置**（含批量域名输入）、**IP 黑白名单**（类型切换、批量粘贴、列表展示）。文案与注释 MUST 使用**中文**（与现有仪表盘规范一致）。

#### Scenario: 各页在无后端时优雅降级

- **WHEN** 管理 API 返回错误或网络失败
- **THEN** 页面 MUST 展示 Ant Design `Result`/`Alert` 或等价错误提示，且 MUST NOT 白屏崩溃

### Requirement: 导出与批量交互

日志页 MUST 提供 **CSV 导出**触发控件；CORS 与 IP 名单页 MUST 提供**批量输入**控件（多行文本）。操作成功或失败后 MUST 有明确 `message` 反馈。

#### Scenario: 批量域名为空提交被拦截

- **WHEN** 用户未输入任何有效域名即提交批量添加
- **THEN** 前端 MUST 提示校验错误且不发送空污染请求（或后端拒绝二者择一，须在实现中一致）

### Requirement: 系统设置页

仪表盘 MUST 提供 **系统设置** 路由与表单（或分组只读卡片），展示 `admin-system-settings` 规范中的字段及 **mutability** 提示；管理员保存后 MUST 根据响应展示「已热更新」或「需重启进程」说明。

#### Scenario: 操作员不可见保存控件

- **WHEN** 当前用户角色为 `operator`
- **THEN** 系统设置页不展示写提交或等价保存入口

### Requirement: 运行告警展示

仪表盘 MUST 在管理布局内消费告警状态 API，并按 `admin-runtime-alerts` 规范展示告警条。

#### Scenario: 告警与 i18n

- **WHEN** 告警条展示摘要编码
- **THEN** 文案经 i18n 解析为中英之一，与当前语言一致

