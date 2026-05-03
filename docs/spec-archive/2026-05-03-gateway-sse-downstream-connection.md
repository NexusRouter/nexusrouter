# Chat SSE 下游 Connection 响应头

## 功能说明

当上游对 Chat Completions 返回 **`Content-Type`** 含 **`text/event-stream`** 时，在按 hop-by-hop 规则剔除上游响应中的 **`Connection`** 等头之后，若回写至客户端的响应中 **`Connection`** 仍缺省或仅空白，网关在回写前补充 **`Connection: keep-alive`**，以明确 HTTP/1.x 长连接语义。

## 实现要点

- 与既有 **`X-Accel-Buffering`**、**`Cache-Control`** 补充逻辑同属 **`ensureSSEProxyResponseHeaders`**，仅在判定为 SSE 响应时生效。

## 兼容性

- 非 SSE 响应不添加上述补充。
