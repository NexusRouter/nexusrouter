# 公开上传路径响应 Cache-Control

## 行为

在引擎级 **`ErrorJSON`** 之后注册中间件，于 **`c.Next()`** 之后对 **`GET`、`HEAD`** 且路径以 **`/uploads/`** 开头的请求，在响应尚未包含非空 **`Cache-Control`**、且状态为成功（**2xx**）或 **`304`** 时，补充 **`Cache-Control: public, max-age=604800`**。**404** 等错误响应不写入该头；处理器已设置 **`Cache-Control`** 时不覆盖。
