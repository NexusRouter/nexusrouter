# 健康检查：进程启动时间与运行时长

## 功能说明

`GET /health` 在原有 `status`、`version`、`server_time` 基础上增加：

- **`start_time`**：进程启动时刻（UTC，RFC3339Nano），与 `buildinfo.ProcessStart` 一致。
- **`uptime_seconds`**：自启动至响应该请求时已运行秒数（浮点）。

便于外部探活与监控关联进程生命周期，无需额外接口。

## 实现要点

- 在 `buildinfo` 包 `init` 中记录 `ProcessStart`（UTC）。
- `Health` 处理器计算 `uptime_seconds = now - ProcessStart`。

## 兼容性

- 仅扩展 JSON 字段；现有字段语义不变。
