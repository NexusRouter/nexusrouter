# NexusRouter 项目说明

## 概述

NexusRouter 是开源 LLM API 网关：以 **OpenAI 兼容** API 统一接入多厂商模型，并提供路由、配额、观测等能力。许可证：**AGPL-3.0**。

## 技术栈与目录

| 区域 | 技术 | 路径 |
|------|------|------|
| 网关（后端） | Go 1.25+，[Gin](https://github.com/gin-gonic/gin) | `services/gateway/` |
| 前端（预留） | 待定（将置于 `web/`） | `web/` |
| 规范驱动 | [OpenSpec](https://github.com/fission-ai/openspec) | `openspec/`（`specs/` 能力规范，`changes/` 变更提案） |
| CI | GitHub Actions | `.github/workflows/` |

## 仓库布局（约定）

```
openspec/specs/     # 能力规范（真理源）
openspec/changes/   # 进行中的变更提案
services/gateway/ # Go 网关服务（go.mod 在此模块根目录）
web/                # 前端工程（后续初始化）
```

## 本地开发（网关）

```bash
cd services/gateway
go run .
```

默认监听 `:8080`，健康检查：`GET /health`。

---

## English (short)

NexusRouter: open-source LLM API gateway (OpenAI-compatible), AGPL-3.0. Backend: **Go + Gin** in `services/gateway/`. Frontend TBD under `web/`. Specs and proposals live under `openspec/`.
