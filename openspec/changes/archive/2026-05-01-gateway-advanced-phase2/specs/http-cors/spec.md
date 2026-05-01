## ADDED Requirements

### Requirement: 可配置 CORS 中间件

当配置启用 **`cors`** 段时，网关 MUST 注册 **CORS** 中间件，对 **`AllowOrigins`** 列表中的 Origin 返回匹配的 **`Access-Control-Allow-Origin`**（**MUST NOT** 在携带凭证时使用 **`*`** 作为回包 Origin，除非 `design.md` 明确禁止凭证场景）。**`AllowMethods`**、**`AllowHeaders`**、**`ExposeHeaders`**、**`Max-Age`** MUST 可从配置文件读取；未列出的请求头在预检中 MAY 拒绝。

#### Scenario: 预检成功

- **WHEN** 浏览器对 **`OPTIONS /v1/chat/completions`** 发起预检且 **`Origin`** 在允许列表
- **THEN** 响应 **204** 或 **200**（实现固定），且包含 **`Access-Control-Allow-Methods`** 覆盖 **`POST`**

#### Scenario: 未允许的 Origin

- **WHEN** **`Origin`** 不在允许列表
- **THEN** 响应 MUST **不**带可误导浏览器成功跨域的 **`Access-Control-Allow-Origin`**（或返回错误由 `design.md` 固定）

### Requirement: 默认关闭

若未配置 **`cors`** 或 **`enabled: false`**，网关 MUST **不**添加全局 CORS 头（与当前行为兼容）。

#### Scenario: 关闭时不影响 curl

- **WHEN** 客户端无 **`Origin`** 头请求 **`POST /v1/chat/completions`**
- **THEN** 行为与未引入 CORS 功能前一致（除其他中间件外无额外 CORS 头）
