# Chat SSE 下游响应头

## 行为

当上游对 Chat Completions 返回 **`Content-Type`** 含 **`text/event-stream`** 时，网关在将响应回写至客户端前，在响应头中设置 **`X-Accel-Buffering: no`**，以降低中间反向代理对响应体的缓冲，使增量数据更易及时到达客户端。

若上游未提供非空 **`Cache-Control`**，网关补充 **`Cache-Control: no-cache`**，以降低中间层误缓存流式正文的风险；若上游已给出 **`Cache-Control`**，网关不覆盖。

非 SSE 响应不添加上述补充。上游其余响应头仍按既有 hop-by-hop 剔除与透传规则处理。
