# Chat 上游出站不转发 Accept-Encoding

## 摘要

对 **`POST /v1/chat/completions`**，网关在构造发往上游的出站 HTTP 请求时 MUST 移除 **`Accept-Encoding`**，使上游不因客户端对**响应**压缩的声明而改变编码协商；其它业务请求头在既有 hop-by-hop 剔除规则之外仍按透传要求保留。

## 行为说明

- 与 hop-by-hop 头剔除同属出站请求净化；对 Chat 专用上游 HTTP 客户端关闭「自动请求压缩响应」语义，使出站请求不因标准库默认行为再附带压缩算法声明。
- 入站 **`Content-Encoding: gzip`** 的请求体解压（引擎级中间件）语义不变。
