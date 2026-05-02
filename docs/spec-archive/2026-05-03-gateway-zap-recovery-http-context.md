# ZapRecovery 记录 HTTP 方法与路径

## 功能说明

当请求处理链发生 panic 并由 **`ZapRecovery`** 捕获时，除 **`error`**、**`request_id`** 与 **`stack`** 外，结构化日志中补充 **`method`**（HTTP 方法）与 **`path`**（与访问日志一致：优先路由模板路径，否则为请求 URL 路径），便于在日志系统中按接口定位问题，且不写入请求体。

## 实现要点

- 路径选择与 **`ZapHTTPAccessLog`** 一致：**`FullPath()`** 非空则用模板路径，否则用 **`Request.URL.Path`**。

## 兼容性

- 客户端可见的 **500** JSON 与响应头行为不变；仅服务端 Zap 字段扩展。
