**简体中文** | [English](./README_EN.md)

# NexusRouter

**开源 LLM API 网关 | AGPLv3 | 多厂商智能路由**

NexusRouter 是一款高性能的开源 LLM API 网关，通过 **OpenAI 兼容** 接口统一接入主流大模型服务（OpenAI、Anthropic、通义千问、Gemini 等），并提供用量统计、计费、限流与故障自动切换等能力。

## 核心特性

- 与 **OpenAI API** 高度兼容，便于现有客户端与 SDK 迁移
- **多厂商聚合** 与负载均衡
- **API Key** 管理与配额控制
- **Token 用量** 追踪与实时计费
- **限流** 与请求重试
- 轻量、高性能：**Go 1.24** + **Gin**（`services/gateway`），控制台 **React + Vite**（`web/dashboard`）
- **AGPLv3** 开源许可（闭源商业使用需遵守许可或另行取得授权）

## 快速开始

**环境**：Go **1.24.x**、Node **≥ 22（LTS）**、[pnpm](https://pnpm.io/) **9.x**（建议 `corepack enable && corepack prepare pnpm@9.15.9 --activate`）。

**仓库根目录（Husky 提交钩子）**

```bash
pnpm install
```

**网关（Go）**

```bash
cd services/gateway
go run ./cmd/api
```

默认 **[http://127.0.0.1:8080](http://127.0.0.1:8080)**，健康检查：`GET /health`。

**控制台（前端）**

```bash
cd web/dashboard
pnpm install
pnpm dev
```

更多说明见 `openspec/project.md` 与 `services/gateway/README.md`。

## 参与贡献

向本仓库提交贡献前，请阅读并签署 **贡献者许可协议（CLA）**。合并到 `main` 分支的 PR 会通过 [CLA Assistant](https://github.com/cla-assistant/cla-assistant) 校验签名；签名文件存放于 `signatures/version1/cla.json`，协议正文见 [NexusRouter CLA 文档](https://github.com/NexusRouter/cla/blob/main/CLA.md)。

关系型数据库表结构的设计与评审约定见 `[openspec/specs/database-design-standards/spec.md](openspec/specs/database-design-standards/spec.md)`。**引入或变更关系型 schema 的 PR** 请在描述中对照该规范中的需求（SHALL/MUST）与场景自检并简要说明结论。

## 开发：提交前检查（Husky）

克隆后请在**仓库根目录**执行一次 `**pnpm install`**（与 `web/dashboard` 相同，使用 **pnpm 9.x**，建议先 `corepack enable`），以安装 **Husky** 并注册 Git `pre-commit`（`core.hooksPath` 指向 `.husky/_`）。

`git commit` 前会根据**暂存区路径**自动运行与 CI 对齐的检查（未改动的子项目会跳过）：

- 改了 `services/gateway/`：gofmt、golangci-lint、go vet、go test、go build  
- 改了 `web/dashboard/`：`pnpm install --frozen-lockfile`、lint、带 coverage 的 test、build  
- 改了 `openspec/` 或 CI/脚本：另跑 `openspec validate`

**未包含**：网关侧可选维护步骤（见 `services/gateway/README.md`）。临时跳过钩子（仅限应急）：`HUSKY=0 git commit ...`。

## 许可证

本项目采用 [GNU Affero General Public License v3.0](https://www.gnu.org/licenses/agpl-3.0.html)（AGPL-3.0）。使用、修改与分发时请遵守许可证全文。