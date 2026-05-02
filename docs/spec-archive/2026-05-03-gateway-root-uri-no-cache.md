# 网关：根路径严格 URI 的 Cache-Control

## 功能说明

当 HTTP 请求的 **`RequestURI`** 严格等于 **`/`**（无查询串）时，在业务处理链执行之前向响应写入 **`Cache-Control: no-cache`**，避免中间层或浏览器将根路径响应当作可长期复用的缓存资源。带查询的根请求（如 **`/?x=1`**）不套用本规则。若后续处理器或 **`NoRoute`** 另行设置 **`Cache-Control`**，允许按常规头合并语义由后者覆盖。

## 实现要点

- 在 **`internal/router/middleware.go`** 中实现 **`RootStrictNoCache`**，判定 **`c.Request.RequestURI == "/"`** 后 **`c.Header("Cache-Control", "no-cache")`**，再 **`c.Next()`**。
- 在 **`internal/provider/router.go`** 中于 **`ErrorJSON`** 之后、**`UploadsStaticCache`** 之前注册该中间件。

## 兼容性

- 非根路径、或根路径带查询的请求不因本规则单独出现上述 **`Cache-Control`**；**`/health`**、**`/uploads/`** 等既有行为不变。
