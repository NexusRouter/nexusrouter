# 网关对未注册 OpenAI 兼容 v1 路径返回 invalid_request_error 形 404

## 行为

- 当 **`NoRoute`** 触发且请求路径为 **`/v1`** 或以 **`/v1/`** 开头时，响应 **HTTP 404**，body 为顶层含 **`error`** 的 JSON；**`error.type`** 为 **`invalid_request_error`**，**`error.message`** 含方法与路径说明；**`error.param`** 与 **`error.code`** 为空字符串。
- 响应头 **`X-Request-ID`** 与 **`RequestID`** 中间件已确定的请求 ID 一致；若此前无 ID 则生成并回写。
- 其它未匹配路径仍返回网关统一 **`code`/`message`/`request_id`** 的 **404**。

## 动机

与常见 OpenAI 兼容客户端对「未知子路径」错误体的解析习惯对齐，区别于网关内部 **`NOT_FOUND`** 与已列路径的 **501**。
