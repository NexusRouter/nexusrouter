# 网关：Accept-Language 请求语言标签

## 功能说明

引擎在 **`RequestID`** 之后根据请求头 **`Accept-Language`** 归约出稳定语言标签 **`zh-CN`** 或 **`en`**，并写入 Gin 上下文与 **`c.Request.Context()`**，便于后续处理器、访问日志或错误文案按客户端语言偏好扩展，而不解析完整 BCP 47 列表。

## 实现要点

- 中间件 **`AcceptLanguage`** 注册于 **`RequestID`** 之后、**`ZapHTTPAccessLog`** 之前（见 `ProvideEngine`）。
- 归约规则：空头为 **`en`**；整段头值（不区分大小写）以 **`zh`** 为前缀则为 **`zh-CN`**，否则为 **`en`**。
- 包 **`internal/locale`** 提供 **`FromContext`** / **`WithLocale`** 与 Gin 键常量，子上下文继承同一标签。

## 兼容性

- 不修改响应头；不改变未使用该 API 的现有处理器行为。
