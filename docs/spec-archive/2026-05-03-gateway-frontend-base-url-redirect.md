# 网关：可选外置前端基址与未匹配路由 301 重定向

## 功能说明

当需要把管理控制台或静态资源部署在与 API 不同的源上时，可将环境变量 **`NEXUSROUTER_FRONTEND_BASE_URL`** 设为合法 **`http`** 或 **`https`** 基址（可带路径前缀）。网关对**未命中任何已注册路由**的请求返回 **`301 Moved Permanently`**，**`Location`** 为该基址与客户端请求原始 URI（含 query）的拼接。

未设置、为空或非法 URL 时行为与此前一致：未匹配路由返回统一 JSON **`NOT_FOUND`**。非法值在配置加载阶段被忽略，不阻塞启动。

## 实现要点

- **`internal/config`**：从 **`NEXUSROUTER_FRONTEND_BASE_URL`** 读入并校验、规范化（去空白、去尾 **`/`**）。
- **`internal/provider`**：在 **`NoRoute`** 中若基址有效则 **`Redirect`**，否则沿用 **`WriteGatewayError`**。

## 兼容性

未配置该变量时与旧版完全一致。
