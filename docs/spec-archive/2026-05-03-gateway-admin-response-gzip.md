# 管理端 JSON 响应可选 Gzip

## 行为

在 **`/api/admin/v1`** 前缀的路由组上注册响应压缩中间件（与引擎级「入站 **`Content-Encoding: gzip`** 解压」分离）。当请求头 **`Accept-Encoding`** 表明客户端接受 **`gzip`** 时，可对 JSON 等可压缩响应使用 **`Content-Encoding: gzip`**；未声明接受时不对响应强制 gzip。Chat 代理与 SSE 路径不在该组内，不受本中间件影响。
