# 流式 Chat 合并 stream_options.include_usage

## 行为

- 对 `POST /v1/chat/completions`，当请求 JSON 顶层 `stream` 为 `true` 且 `stream_options` 缺失或为对象时，网关可在转发前合并 `stream_options`，将 `include_usage` 设为 `true`，并保留 `stream_options` 中其它字段。
- 环境变量 `NEXUSROUTER_CHAT_STREAM_INCLUDE_USAGE`：未设置时按启用处理；为 `false` 时不做上述合并。
- 若 `stream_options` 已存在且无法解析为 JSON 对象，则不修改请求体，避免破坏非标准载荷。

## 验收要点

- 流式请求在启用合并时，上游收到的 body 含 `stream_options.include_usage === true`。
- 关闭合并时，上游 body 与客户端一致（不因本特性改写 `stream_options`）。
