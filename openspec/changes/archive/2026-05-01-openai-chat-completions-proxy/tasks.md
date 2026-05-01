## 1. 测试先行 — OpenAPI 与 Swagger UI（红 → 绿）

- [x] 1.1 新增测试：`httptest` 请求 **GET `/openapi.yaml`**（或设计选定的 JSON 路径），断言 **200**、正文含 **`openapi: 3.0`**（YAML）或根 JSON **`"openapi":"3.0`** 前缀；禁止 `t.Skip` 掩盖未实现（应失败直至实现）
- [x] 1.2 解析 YAML/JSON：断言 **`paths./v1/chat/completions.post`** 存在且含 **`requestBody`**
- [x] 1.3 断言 **`components.securitySchemes`** 含 bearer 约定且等价 HTTP Bearer；断言 **`/v1/chat/completions` POST** 在根或操作级 **`security`** 中引用
- [x] 1.4 断言文档含 **`https://developers.openai.com/api/reference/overview`**
- [x] 1.5 断言 **GET `/swagger/index.html`**（或等价）在 UI 开启配置下 **200**，且 HTML 引用 OAS3 的 **`url:`** 或 **`configUrl`**（与实现一致）

## 2. swag 注解与 OpenAPI 生成

- [x] 2.1 在 `cmd/api` 或集中入口补充 **通用 API** 注释（`@title`、`@version`、`@description` 含 overview 链接、`@host`、`@BasePath`）
- [x] 2.2 为 **POST `/v1/chat/completions`** 添加 swag 操作注释；定义精简 **request/response** DTO（与网关支持子集一致）
- [x] 2.3 配置 **`swag init`**（`-g`、`-d`、`--parseInternal`）及 **OAS3** 产出（升级 swag v2 或 2→3 转换）；**`openapi.yaml`（3.0）** 与 `docs.go` 纳入提交策略

## 3. Gin — 文档路由与 Swagger UI

- [x] 3.1 注册 **Swagger UI**（`ginSwagger` + `files`），加载本服务 **OpenAPI 3** URL
- [x] 3.2 注册 **GET `/openapi.yaml`**（及可选 **GET `/openapi.json`**），`Content-Type` 正确
- [x] 3.3 配置项控制文档 UI 开关（测试环境默认开）

## 4. CI 与 README（文档面）

- [x] 4.1 CI：`swag init` 或 `make docs` + 可选 **`git diff --exit-code`** 防生成物漂移
- [x] 4.2 更新 **`services/gateway/README.md`**：生成命令、OpenAPI/UI URL、[OpenAI API 参考](https://developers.openai.com/api/reference/overview)

## 5. 配置与依赖注入（代理）

- [x] 5.1 在 `internal/config` 增加上游基址、超时、鉴权密钥（及可选 `forward_client_authorization`）等，接入 Viper/环境变量
- [x] 5.2 Wire 装配注入代理 handler 依赖（配置、日志、Transport）

## 6. 反向代理核心

- [x] 6.1 新增代理模块：基于 **`httputil.ReverseProxy`**，`Director` 目标为配置基址 + `/v1/chat/completions`
- [x] 6.2 hop-by-hop 与 **`ModifyResponse`**（如需），保证 JSON/SSE
- [x] 6.3 流式 **`Flusher`** 路径验证（单测或集成测试）

## 7. 路由与中间件（代理）

- [x] 7.1 **`POST /v1/chat/completions`**：鉴权中间件 + 代理 handler
- [x] 7.2 非法方法：**405/404** 策略固定 + 统一 JSON
- [x] 7.3 代理链 **`recover`** → 500 + JSON + Zap

## 8. 鉴权（代理）

- [x] 8.1 Bearer / **`X-API-Key`** 校验，失败 **401** 不转发
- [x] 8.2 发往上游的 **`Authorization`** 按设计默认剥离或注入、可配置透传

## 9. 错误与可观测（代理）

- [x] 9.1 网关侧 **502/504/500/401** 统一 JSON，与现有中间件对齐
- [x] 9.2 网关拦截路径写 Zap（脱敏）；上游错误体 **原样透传**、不包网关包装

## 10. 代理测试与文档

- [x] 10.1 **`httptest`** 双端：鉴权失败、上游 200 JSON、上游 4xx 原样、连接失败→502/504
- [x] 10.2 可选：短 SSE 或 **`-short`** 跳过说明（当前以 README 说明为主，SSE 专项用例可后续补充）
- [x] 10.3 README / `openspec/project.md` 链接：**代理**环境变量说明
