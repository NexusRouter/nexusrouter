# 网关：请求 ID 双响应头与备用入站头

## 功能说明

**`RequestID`** 中间件在确定最终请求 ID 后，除 **`X-Request-ID`** 外，将**同一值**写入响应头 **`X-Oneapi-Request-Id`**，便于仅读取该头的客户端与日志系统对齐。入站时 **`X-Request-ID`** 优先（经首尾空白去除后非空则用之）；若其为空而 **`X-Oneapi-Request-Id`** 非空，则采用后者；二者皆空时生成新 ID。本地开发场景的 **CORS** 预检允许列表包含上述两请求头名，避免浏览器拦截自定义头。

## 实现要点

- **`internal/router/middleware.go`**：读取顺序、双头回写、**`strings.TrimSpace`**。
- **`internal/provider/cors.go`**：本机开发 Origin 下的 **`AllowHeaders`** 增补 **`X-Oneapi-Request-Id`**。

## 兼容性

- 已使用 **`X-Request-ID`** 的客户端行为不变；**`request_id`** 错误字段仍与 **`X-Request-ID`** 一致。
