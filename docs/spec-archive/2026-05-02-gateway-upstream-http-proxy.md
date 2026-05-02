# 网关：Chat 上游出站 HTTP 代理

## 功能说明

网关为 **POST `/v1/chat/completions`** 反向代理创建的上游 **`http.Transport`** 支持通过环境变量 **`NEXUSROUTER_UPSTREAM_HTTP_PROXY`** 指定固定代理 URL（如 **`http://host:port`**）。设置后，对该路径发往模型上游的 TCP 连接按 **`net/http.ProxyURL`** 语义经该代理建立。

未设置或值为空时，不写入 **`Transport.Proxy`**（保持 **`nil`**），从而与仅依赖 **`HTTP_PROXY`/`HTTPS_PROXY`/`NO_PROXY`** 等标准环境变量的部署方式兼容。无法解析为含 scheme 与 host 的 URL 时忽略该变量，不阻断进程启动。

## 实现要点

- 配置字段由 **`internal/config`** 从环境变量加载；**`internal/handler/chatproxy`** 在构造 **`Transport`** 时按需设置 **`Proxy`**。
- 模型库实例路径使用实例级 **`timeout`** 构造独立 **`Transport`** 时，同步应用上述代理配置。

## 兼容性

- 与未配置该变量时的行为一致；与仅使用系统代理环境变量的场景并存。
