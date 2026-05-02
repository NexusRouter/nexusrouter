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

仪表盘源码中的 **注释与用户可见文案**（若占位）MUST 使用 **中文**；变量与函数名 MUST 使用 **英文**；Git commit message MUST 使用 **中文**（与本变更组织约定一致）。

#### Scenario: 代码审查抽样

- **WHEN** 审查者打开主要入口与配置文件的注释
- **THEN** 注释为中文且标识符为英文

