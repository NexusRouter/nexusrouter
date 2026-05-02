## ADDED Requirements

### Requirement: 首次启动门闸中间件

`services/gateway` MUST 在未初始化全局状态下，对除**明确白名单**外的所有 HTTP 请求返回 **403** 或 **503**（实现选定其一并在 OpenAPI 中固定），且 body MUST 符合既有 **`gateway-backend`** JSON 错误约定（含 **`code`**、**`message`**、**`request_id`**）。白名单 MUST 至少包含：**初始化状态查询**、**完成初始化提交**、**健康检查**（若项目已存在标准健康路径）、**静态资源**（若由网关直出）。**已初始化**状态下，白名单逻辑 MUST 不再拦截正常业务与管理 API。

#### Scenario: 未初始化访问非白名单 API

- **WHEN** 系统 **`initialized=false`** 且客户端请求任意非白名单已注册路由（如管理业务 API）
- **THEN** 响应为 JSON 错误且 **`code`** 为稳定枚举（如 **`bootstrap_required`**），且 **`request_id`** 与头一致

#### Scenario: 已初始化不触发门闸拒绝

- **WHEN** 系统 **`initialized=true`** 且客户端请求普通管理或代理 API
- **THEN** MUST NOT 因本门闸单独返回 **`bootstrap_required`**

### Requirement: 初始化与重置路由注册

初始化状态查询、完成初始化提交、超级管理员重置 MUST 注册为独立 handler，路径前缀 MUST 在 `design.md` 或 OpenAPI 中固定（建议 **`/api/bootstrap`** 或等价）。完成初始化与状态查询在未初始化时 MUST 不要求 Bearer 令牌；重置 MUST 要求认证与超管角色。

#### Scenario: 重置未携带令牌

- **WHEN** 客户端不带认证调用重置接口
- **THEN** 响应 **401** 且 JSON 错误格式符合规范
