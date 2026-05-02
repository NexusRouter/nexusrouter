# 归档：网关 OpenAI 兼容图像生成 POST /v1/images/generations 反向代理

## 功能摘要

网关注册 **`POST /v1/images/generations`**，在鉴权与按 Key 限流后与 Embeddings、Moderations 代理共享上游选择及模型改写语义；请求体须为含非空字符串 **`prompt`** 的 JSON 对象；若未提供非空 **`model`**，转发前合并默认 **`dall-e-2`**。非 POST（除 OPTIONS 返回 204）返回 405。

## 验收要点

- YAML 快照上游与模型库实例两种路径均可转发。
- 出站剔除 hop-by-hop 与 **`Accept-Encoding`**，与其它 OpenAI 兼容代理一致。
