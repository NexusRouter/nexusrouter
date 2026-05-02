# admin-swagger-embed Specification

## Purpose

历史变更曾规划在管理控制台或网关侧提供 **Swagger / OpenAPI** 调试与机读契约。**当前产品决策**：**不**在 `web/dashboard` 提供任何 OpenAPI 调试界面；**不**在网关暴露 **`/openapi.json`** 等 OpenAPI HTTP 端点；**不**要求交付或维护 **OpenAPI 3.0** 规范文档（含仓库内生成物或以其他形式发布的 OAS3）。

## Requirements

### Requirement: 无控制台调试器

`web/dashboard` MUST NOT 提供 **Swagger UI**、**OpenAPI Try it out** 或等价内嵌 API 调试页。

#### Scenario: 侧栏无调试入口

- **WHEN** 审查者检查已登录管理布局的导航项与路由表
- **THEN** MUST NOT 存在「接口调试」「Swagger」「OpenAPI Playground」等入口

### Requirement: 无网关 OpenAPI 端点

`services/gateway` MUST NOT 注册 **`GET /openapi.json`**、**`GET /openapi.yaml`** 或 **Swagger UI** 静态路由。

#### Scenario: 规范端点不存在

- **WHEN** 客户端请求 **`GET /openapi.json`** 或 **`GET /openapi.yaml`**
- **THEN** 响应为 **404**（或路由未匹配之等价 404 行为），且**不**返回 OpenAPI 文档体
