## MODIFIED Requirements

### Requirement: 错误响应与日志约定

网关在处理 **自身生成的错误响应**（不包含上游透传的原始错误体）时 MUST 使用 **JSON**，且 body MUST 为对象并 MUST 包含：**`code`**（字符串，机器可读、稳定枚举）、**`message`**（字符串，人类可读）、**`request_id`**（字符串，与响应头 **`X-Request-ID`** 一致；若客户端请求已携带 **`X-Request-ID`**，MUST 使用该值）。错误路径 MUST 记录到 **Zap**，且日志 MUST 包含 **`request_id`** 与 **`code`**（或等价结构化字段），且 MUST NOT 记录完整 API Key 或 **`Authorization`** 头原文。

#### Scenario: 未知路由携带请求 ID

- **WHEN** 客户端请求未注册路由且请求头包含 **`X-Request-ID: probe-1`**
- **THEN** 响应为 JSON，其 **`request_id`** 为 **`probe-1`**，响应头 **`X-Request-ID`** 为 **`probe-1`**，且 Zap 记录该 **404**（或等价）事件并带相同 **`request_id`**

#### Scenario: 未带请求 ID 时服务端生成

- **WHEN** 客户端请求触发网关错误且未携带 **`X-Request-ID`**
- **THEN** 响应头 **`X-Request-ID`** 非空，JSON body 中 **`request_id`** 与该头一致，且 Zap 含相同 **`request_id`**

#### Scenario: Panic 恢复响应含请求 ID

- **WHEN** 处理链触发 panic 并由恢复中间件转换为 **500** JSON
- **THEN** 该 JSON MUST 包含 **`request_id`** 且与 **`X-Request-ID`** 响应头一致

## ADDED Requirements

### Requirement: 请求 ID 中间件全局启用

HTTP 引擎 MUST 在业务路由之前注册 **`RequestID`**（或等价命名）中间件：若请求已带 **`X-Request-ID`**，MUST 透传该值；否则 MUST 生成并写入响应头 **`X-Request-ID`**；且 MUST 将当前请求 ID 存入 Gin 上下文供后续处理器与错误封装读取。

#### Scenario: 下游处理器可读 request_id

- **WHEN** 任意已注册业务处理器执行
- **THEN** 其可通过上下文读取与 **`X-Request-ID`** 一致的请求 ID 字符串
