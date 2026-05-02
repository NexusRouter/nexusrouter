# 归档：OpenAI 兼容 `POST /v1/completions` 显式 501

## 功能

对已鉴权、按 key 限流后的 **`POST /v1/completions`**（旧版文本补全，非 Chat Completions），网关返回 **501 Not Implemented**，响应体为统一 JSON，**`code`** 为 **`NOT_IMPLEMENTED`**。

## 动机

与已实现的 **`POST /v1/chat/completions`** 区分；避免客户端将该路径误判为未注册路由（如 **404**），与「已识别但未实现」的 OpenAI 兼容子路径族行为一致。
