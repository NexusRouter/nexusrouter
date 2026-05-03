# 归档：网关 `/v1/models` 的 OPTIONS

## 摘要

网关注册 **`OPTIONS /v1/models`** 与 **`OPTIONS /v1/models/:model`**：不要求入口鉴权，响应 **`204 No Content`**、无 JSON body，与既有 **`OPTIONS /v1/chat/completions`** 行为一致；与引擎级 CORS 中间件并存，用于在无 **`Origin`** 等不触发 CORS 分支时仍避免模型路径落入未匹配路由。
