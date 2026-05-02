# 归档：网关 OpenAI 兼容语音合成 POST /v1/audio/speech 反向代理

## 功能摘要

网关注册 **`POST /v1/audio/speech`**，在鉴权与按 Key 限流后与图像生成等代理共享上游选择及模型改写语义；请求体须为 JSON 对象且含非空字符串 **`input`** 与 **`voice`**；若未提供非空 **`model`**，转发前合并默认 **`tts-1`**。非 POST（除 OPTIONS 返回 204）返回 405。

## 验收要点

- YAML 快照上游与模型库实例两种路径均可转发。
- 出站剔除 hop-by-hop 与 **`Accept-Encoding`**，与其它 OpenAI 兼容代理一致。
