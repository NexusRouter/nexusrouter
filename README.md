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

**环境**：Go **1.24.x**、Node **≥ 22（LTS）**、[pnpm](https://pnpm.io/) **9.x**（建议 `corepack enable`）。

**网关（Go）**

```bash
cd services/gateway
go run ./cmd/api
```

默认 **http://127.0.0.1:8080**，健康检查：`GET /health`。

**控制台（前端）**

```bash
cd web/dashboard
pnpm install
pnpm dev
```

更多说明见 `openspec/project.md` 与 `services/gateway/README.md`。

## 参与贡献

向本仓库提交贡献前，请阅读并签署 **贡献者许可协议（CLA）**。合并到 `main` 分支的 PR 会通过 [CLA Assistant](https://github.com/cla-assistant/cla-assistant) 校验签名；签名文件存放于 `signatures/version1/cla.json`，协议正文见 [NexusRouter CLA 文档](https://github.com/NexusRouter/cla/blob/main/CLA.md)。

## 许可证

本项目采用 [GNU Affero General Public License v3.0](https://www.gnu.org/licenses/agpl-3.0.html)（AGPL-3.0）。使用、修改与分发时请遵守许可证全文。
