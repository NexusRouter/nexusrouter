[简体中文](./README.md) | **English**

# NexusRouter

**Open-Source LLM API Gateway | AGPLv3 | Multi-Provider Smart Routing**

NexusRouter is a high-performance, open-source LLM API gateway that exposes **OpenAI-compatible** endpoints to unify access to major LLM providers (OpenAI, Anthropic, Qwen, Gemini, and more), with usage tracking, billing, rate limiting, and automatic failover.

## Core Features

- Strong **OpenAI API** compatibility for existing clients and SDKs
- **Multi-provider aggregation** and load balancing
- **API key** management and quota control
- **Token usage** tracking and real-time billing
- **Rate limiting** and request retries
- Lightweight stack: **Go 1.24** + **Gin** (`services/gateway`), dashboard **React + Vite** (`web/dashboard`)
- **AGPLv3** open-source license (closed-source commercial use requires compliance or separate permission)

## Quick Start

**Prerequisites:** Go **1.24.x**, Node **≥ 22 (LTS)**, [pnpm](https://pnpm.io/) **9.x** (recommended: `corepack enable`).

**Gateway (Go)**

```bash
cd services/gateway
go run ./cmd/api
```

Listens on **http://127.0.0.1:8080** by default. Health: `GET /health`.

**Dashboard (frontend)**

```bash
cd web/dashboard
pnpm install
pnpm dev
```

See `openspec/project.md` and `services/gateway/README.md` for more detail.

## Contributing

Before contributing, please read and sign the **Contributor License Agreement (CLA)**. Pull requests merged into `main` are checked with [CLA Assistant](https://github.com/cla-assistant/cla-assistant); signatures are stored in `signatures/version1/cla.json`, and the agreement text is published in the [NexusRouter CLA repository](https://github.com/NexusRouter/cla/blob/main/CLA.md).

## License

This project is licensed under the [GNU Affero General Public License v3.0](https://www.gnu.org/licenses/agpl-3.0.html) (AGPL-3.0). Use, modification, and distribution are subject to the full license text.
