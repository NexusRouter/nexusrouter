# 归档：网关 OpenAI 兼容 Moderations 反向代理

## 功能摘要

网关注册 **`POST /v1/moderations`**，在鉴权与按 Key 限流后与 Embeddings 代理共享上游选择及模型改写语义；请求体须为含合法 **`input`** 的 JSON 对象；若省略或空 **`model`**，转发前合并为 **`text-moderation-latest`**。非 POST（除 OPTIONS 返回 204）返回 405。

## 验收要点

- YAML 快照上游与模型库实例两种路径均可转发。
- 出站剔除 hop-by-hop 与 **`Accept-Encoding`**，与 Chat 出站头处理一致。
