# 网关请求体可复用读取（2026-05-03）

## 摘要

提供 **`internal/bodyreuse`**：**`GetRequestBody`** 首次读取并缓存请求体；**`ResetRequestBody`** 在改写后写回 **`c.Request.Body`** 与缓存；**`UnmarshalBodyReusable`** 解析 JSON（或其它类型经 **`ShouldBind`**) 后还原 body，供后续处理器或反向代理再次消费。

## 行为说明

- 与 **`GzipRequestDecode`** 之后得到的明文 body 语义一致；大对象仍完全读入内存，与一次性 **`ReadAll`** 相当。
- Chat 转发路径在模型改写后调用 **`ResetRequestBody`**，保证上游收到的 body 与内部解析一致。
