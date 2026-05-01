# Gateway（Go）

NexusRouter 网关服务，模块路径：`github.com/NexusRouter/nexusrouter/services/gateway`。

## 环境要求

- **Go 1.24.x**
- （可选）**Wire**：`go install github.com/google/wire/cmd/wire@v0.6.0`
- （可选）**Air 热重载**：`go install github.com/air-verse/air@latest`，在项目根执行 `air`（需自行添加 `.air.toml`）

## 常用命令

```bash
cd services/gateway
go run ./cmd/api
```

默认监听 **:8080**，健康检查：`GET /health`。

### 重新生成 Wire

```bash
cd services/gateway
go run github.com/google/wire/cmd/wire@v0.6.0 ./cmd/api
```
