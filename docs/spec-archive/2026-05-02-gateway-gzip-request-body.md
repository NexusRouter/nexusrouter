# 网关：Gzip 压缩请求体透明解码

## 功能说明

当 HTTP 请求携带 **`Content-Encoding: gzip`** 时，引擎在业务处理前将请求体解压为明文流，并移除该编码声明、将 **`Content-Length`** 更新为解压后长度。后续路由（含 Chat 反向代理）按与客户端直接发送未压缩 body 相同的语义读取与转发。

解压失败时返回 **400**，错误体符合网关统一 JSON 约定。

## 实现要点

- 中间件 **`GzipRequestDecode`** 注册于 **`ZapHTTPAccessLog`** 之后、**`ZapRecovery`** 之前（见 `ProvideEngine`）。
- 使用标准库 **`compress/gzip`** 全量读入解压结果后替换 **`Request.Body`**，避免与原始 **`Content-Length`**（压缩长度）不一致。

## 兼容性

- 未带 **`Content-Encoding: gzip`** 的请求行为不变。
- 仅识别值为 **`gzip`** 的编码声明（大小写不敏感）；其它编码不在本轮处理范围。
