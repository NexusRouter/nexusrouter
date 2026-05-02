## ADDED Requirements

### Requirement: 管理面板内嵌 API 调试

管理控制台 MUST 集成 **Swagger UI**（或功能等价的 OpenAPI 调试界面），使用户可在**同一应用内**浏览并尝试调用网关已发布的 HTTP 接口（至少覆盖 **`/health`**、文档化之 **`/v1/chat/completions`** 与当前 OpenAPI 中声明的管理/内部路径之可见子集）。用户 MUST **无需**手动在浏览器打开独立 Swagger 基址即可完成基本调试路径。

#### Scenario: 已登录管理员可打开内嵌调试页

- **WHEN** 已认证管理员从导航进入「接口调试」页面
- **THEN** 页面内呈现可交互的 API 文档与 Try it out 能力（或等价）

#### Scenario: 未登录时不泄露需保护之调试能力

- **WHEN** 用户未通过控制台认证且策略要求保护内嵌调试
- **THEN** MUST 重定向登录或拒绝加载调试器，且 MUST NOT 将管理权限与匿名调试混用（与 `design.md` 安全策略一致）

### Requirement: 与 OpenAPI 来源一致

内嵌调试所使用的 API 定义 MUST 与网关 **`/openapi.json`** 或 **`/openapi.yaml`** 在内容上同源，避免面板与真实服务漂移。

#### Scenario: 版本或路径变更可发现

- **WHEN** 网关升级导致 OpenAPI 变更
- **THEN** 内嵌页在刷新后 MUST 反映新定义（无长期缓存导致陈旧契约）
