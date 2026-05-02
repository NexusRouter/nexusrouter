# 归档：网关 OpenAI 兼容 Embeddings 反向代理

## 功能摘要

网关注册 **`POST /v1/embeddings`** 与 **`POST /v1/engines/:model/embeddings`**，在鉴权与按 Key 限流后与 Chat 代理共享上游选择及模型改写语义；请求体须为含非空 **`input`** 的 JSON 对象；若路径为 engines 形式且 body 未带 **`model`**，则用路径段补全 **`model`**。非 POST（除 OPTIONS 返回 204）返回 405。

## 验收要点

- YAML 快照上游与模型库实例两种路径均可转发。
- 出站剔除 hop-by-hop 与 **`Accept-Encoding`**，与 Chat 出站头处理一致。
