# 网关：Chat 代理路径 panic 的 Zap 调用栈

## 功能说明

在 **`POST /v1/chat/completions`** 反向代理处理中，若转发链发生 **panic** 并由本路径 **`recover`** 捕获，除错误值与 **`request_id`** 外，向 **Zap** 写入 **`stack`** 字段（捕获时刻的 **Go** 调用栈文本），便于与引擎级恢复中间件一致地从日志定位问题。回写客户端的 **JSON** 错误体仍不含栈或请求体原文。

## 实现要点

- 在 **`internal/handler/chatproxy.go`** 中 **`recover`** 分支内，于 **`Error`** 级别日志增加 **`zap.String("stack", string(debug.Stack()))`**（或等价）。

## 兼容性

- 未发生 **panic** 时行为不变；客户端可见的 **500** 与错误码语义不变。
