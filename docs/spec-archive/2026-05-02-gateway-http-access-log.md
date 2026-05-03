# 网关：引擎级 HTTP 访问日志

## 功能说明

引擎在 **`AcceptLanguage`** 与 **`GzipRequestDecode`** 之间注册 **`ZapHTTPAccessLog`**：每个请求在整链处理结束后向主 Zap 写入一条 **`http_access`** 结构化日志，包含 **`request_id`**、**`status`**、**`method`**、**`path`**、**`client_ip`**、**`latency_ms`**，便于运维关联与排查，而不记录 query、body 或鉴权头。

## 实现要点

- 中间件注册顺序见 **`ProvideEngine`**（**`RequestID`** → **`AcceptLanguage`** → **`ZapHTTPAccessLog`** → **`GzipRequestDecode`** → …）。
- **`path`**：优先 **`c.FullPath()`**，否则回退 **`c.Request.URL.Path`**。
- 与 **`internal/accesslog`** 中可选的 Chat 代理 JSON 访问日志独立；二者可并存。

## 兼容性

- 不改变响应体或响应头（除既有中间件外）；仅增加日志侧车。
