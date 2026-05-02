# 网关：`/health` 支持 HEAD 探活

## 功能说明

除 **`GET /health`** 外，网关注册 **`HEAD /health`**：成功时返回 **200**，**`Content-Type`** 为 **`application/json; charset=utf-8`**，**`Content-Length`** 与同字段 JSON 的 **GET** 响应体长度一致；**HEAD** 不返回响应体，便于负载均衡或编排组件以 **HEAD** 做探活且不传输正文。

## 实现要点

- **`internal/router/router.go`** 对 **`GET`** 与 **`HEAD`** 复用同一 **`handler.Health`** 闭包。
- **`internal/handler/health.go`** 在 **`HEAD`** 分支内 **`json.Marshal`** 与 **GET** 相同的负载以计算长度，再 **`WriteHeader(200)`** 而不写入 body。

## 兼容性

- **`GET /health`** 的 JSON 形状与字段含义不变。
