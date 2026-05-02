# NexusRouter 项目说明

## 概述

NexusRouter 是开源 LLM API 网关：以 **OpenAI 兼容** API 统一接入多厂商模型，并提供路由、配额、观测等能力。许可证：**AGPL-3.0**。

## 技术栈与目录

| 区域 | 技术 | 路径 |
|------|------|------|
| 网关（后端） | Go **1.25.x**，Gin，Wire，Zap，GORM；默认 **SQLite**（`gateway.db`）、可选 **Postgres**（`NEXUSROUTER_DATABASE_URL`）；Redis 等见 `go.mod` | `services/gateway/` |
| 控制台（前端） | Node **≥22**，**pnpm 9**，Vite 6，React 19，TypeScript 5.7，Tailwind v4，antd 6 | `web/dashboard/` |
| 仓库根工具链 | **pnpm 9**（仅 Husky 等极少量 devDependency，与控制台包管理器一致） | `package.json`、`pnpm-lock.yaml`（根目录） |
| 规范驱动 | [OpenSpec](https://github.com/fission-ai/openspec) | `openspec/`（`specs/` 能力规范，`changes/` 变更提案） |
| CI | GitHub Actions | `.github/workflows/` |

## 仓库布局（约定）

```
openspec/specs/       # 能力规范（真理源）
openspec/changes/     # 进行中的变更提案
services/gateway/     # Go 网关（入口：cmd/api，go.mod 在模块根）
web/dashboard/        # React 控制台（pnpm）
```

## 本地开发

**仓库根（Git 提交前钩子）**

```bash
pnpm install   # 安装 Husky；与 web/dashboard 使用同一 pnpm 9.x（建议 corepack）
```

**网关**

```bash
cd services/gateway
go run ./cmd/api
```

默认 **:8080**，`GET /health`；OpenAPI 3 与 Swagger UI、Chat 代理环境变量见 `services/gateway/README.md`。Wire 重生成见同 README。

**控制台**

```bash
cd web/dashboard
pnpm install
pnpm dev
```

---

## English (short)

NexusRouter: open-source LLM API gateway (OpenAI-compatible), AGPL-3.0. Backend: **Go 1.24 + Gin** in `services/gateway/` (entry `cmd/api`). Dashboard: **React + Vite + pnpm** in `web/dashboard/`. Specs under `openspec/`.
