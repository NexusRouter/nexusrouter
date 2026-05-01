## Why

仓库已确定 **Go 网关**（`services/gateway`）与 **React 仪表盘**（`web/dashboard`）的全栈形态，但缺少与团队约定一致的依赖版本、目录脚手架与工程化基线；需要一次性的「项目初始化」规范与落地任务，避免后续各开发者自行引入不一致的栈与版本。

## What Changes

- 在 **`web/dashboard`** 按约定版本初始化 **pnpm + Vite 6 + React 19 + TypeScript 5.7 + Tailwind v4 + antd + React Query + Zustand** 等依赖与基础目录、ESLint/Prettier/Vitest 配置。
- 在 **`services/gateway`** 将模块与目录对齐为 **cmd/internal** 分层，按约定版本引入 **Gin、GORM、Redis、Wire、Viper、Zap、JWT、validator、migrate、swag、testify、air** 等，并提供带 Zap 与 Gin 的 **8080** 入口骨架。
- **BREAKING**：后端 Go 工具链目标从当前仓库的 **Go 1.25 / Gin 1.12** 收敛为规范中的 **Go 1.24.x / Gin v1.10.0**（实施时需降级或显式批准保留更高版本并修订本变更）。
- 文档与提交约定：**注释与文档、Git commit 使用中文**；标识符使用英文。

## Capabilities

### New Capabilities

- `dashboard-frontend`：仪表盘前端的技术栈版本、目录约定（`src/components|pages|stores|services|utils`）、Tailwind v4 + `@tailwindcss/vite`、ESLint/Prettier/Vitest 基线。
- `gateway-backend`：网关后端的模块路径、分层目录（`cmd/api`、`internal/config|router|handler|service|repository`）、依赖版本与日志/HTTP 启动基线。

### Modified Capabilities

- （无）当前 `openspec/specs/` 下尚无已发布能力规范。

## Impact

- **路径**：`web/dashboard/`、`services/gateway/`。
- **依赖**：前端 `pnpm` 与大量 npm 包；后端 `go.mod`/`go.sum` 全面调整。
- **CI**：需在合并本变更后对齐 **Node 22**、**Go 1.24**（若采用规范版本）及可选的前端 `pnpm` 检查任务。
- **工具链**：开发者需安装 **Node ≥22 LTS**、**pnpm 9.x**、**Go 1.24.x**；可选全局 **air**、**wire**。
