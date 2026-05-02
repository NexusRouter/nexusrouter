# 网关：公开 GET /api/status

## 功能说明

网关注册 **`GET /api/status`**，无需鉴权，响应为 JSON：**`success`**、**`message`**、**`data`**；**`data`** 含 **`version`**（与构建信息一致）与 **`start_time`**（进程启动时刻，RFC3339Nano）。在首次初始化未完成、引导门闸拦截其它业务路径时，该路径仍允许访问，便于外部脚本或控制台在引导阶段读取运行元信息。

## 实现要点

- **`internal/handler/apistatus.go`** 提供 **`APIStatus`** 处理器。
- **`internal/router/router.go`** 在 **`/health`** 之后注册 **`GET /api/status`**。
- **`internal/router/bootstrap_gate.go`** 将 **`GET /api/status`** 列入引导期白名单。

## 兼容性

- 不改变 **`GET /health`** / **`HEAD /health`** 的语义与负载形状；本路径采用独立的 **`success`/`data`** 封装，与探活接口并存。
