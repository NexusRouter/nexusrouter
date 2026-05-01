## Context

- 仓库已存在 `services/gateway`（当前为 **Go 1.25 + Gin 1.12**、根目录 `main.go`）与占位 `web/dashboard/`。
- 本变更在用户输入的《项目初始化规范》中，要求 **Go 1.24.x + Gin v1.10.0** 及完整前端依赖矩阵；与现状存在版本与入口结构差异。
- OpenSpec 变更 `project-initialization` 将能力拆为 `dashboard-frontend` 与 `gateway-backend` 两条规范。

## Goals / Non-Goals

**Goals:**

- 后端采用 **`cmd/api` + `internal/*` 分层**，集成 **Zap + Gin**，监听 **8080**；依赖版本与规范表一致（在实施阶段执行 `go get` 与 `go mod tidy`）。
- 前端在 **`web/dashboard`** 使用 **pnpm 9 + Vite 6 + React 19 + TS 5.7 + Tailwind v4（`@tailwindcss/vite`）+ antd + React Query + Zustand** 等，并建立约定子目录与 ESLint/Prettier/Vitest。
- **Wire** 生成注入代码；**Air** 用于本地热重载（文档化安装方式，CI 可选不装）。
- 注释、文档与 **Git commit 使用中文**；代码标识符使用英文。

**Non-Goals:**

- 不实现具体业务 API、鉴权流程、数据库迁移脚本内容（仅骨架与依赖）。
- 不规定生产部署拓扑（K8s/Docker 等另案）。

## Decisions

1. **模块路径**：保持现有 Go 模块名 **`github.com/NexusRouter/nexusrouter/services/gateway`**，不采用规范中的 `go mod init gateway` 短名，以免与仓库多模块布局冲突。
2. **Go / Gin 版本**：以用户规范为准实施 **Go 1.24.x、Gin v1.10.0**；若团队决议保留 1.25/1.12，则须修订 proposal 并更新本 design 与 delta specs。
3. **前端创建方式**：在 **`web/dashboard` 目录内** 执行 `pnpm create vite@latest . --template react-ts`（或等价），避免生成后再搬运；与规范中「根目录 create 再移动」等价且路径更清晰。
4. **Tailwind v4**：使用 **`@tailwindcss/vite`** 插件 + 入口 CSS **`@import "tailwindcss";`**，与 Vite 6 官方推荐一致。
5. **依赖白名单**：除规范表列出的运行时/开发依赖外，**不新增**未声明的第三方库；对等版本以 **pnpm 与 go.mod 锁文件** 为准。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 降级 Go/Gin 与现有提交假设冲突 | 在实施 PR 中单开分支；合并前跑全量 CI；必要时在 proposal 中撤销 BREAKING 并固定新版本。 |
| pnpm/Vite 与 antd 版本解析冲突 | 使用 `pnpm.overrides` 或锁定 `package.json` 中显式版本；任务阶段记录解析命令。 |
| Wire/Air 全局安装污染本机 | 文档推荐 `go install` 固定版本；可选 `tools.go` + `go run` 模式后续迭代。 |

## Migration Plan

1. 合并实施 PR 前：备份或分支当前 `services/gateway`。
2. 实施：按 `tasks.md` 顺序执行；完成后删除旧根 `main.go`（若已迁至 `cmd/api`）。
3. 合并后：更新 `openspec/project.md` 与 CI（Node 22、Go 1.24）与 README 本地开发说明。
4. 回滚：还原 `go.mod` / `web/dashboard` 目录与 CI 提交。

## Open Questions

- 组织是否 **强制** Go 1.24（与现有 1.25 工具链）— 需负责人确认后锁定 `go` directive。
- 前端是否在首版即接入 **antd 全量** 或按需子包 — 默认按规范全量 `antd`；若包体过大可后续拆为子路径导入（需改 spec）。
