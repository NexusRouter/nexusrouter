# 网关对显式列出的 OpenAI 兼容 v1 子路径返回 501

## 行为

- 对一组在路由层显式注册的 OpenAI 兼容子路径（如 Files、Fine-tuning、Assistants、Threads、按模型 id 删除等），在通过与其他受保护 v1 接口相同的 **`GatewayAuth`** 与 **`KeyRateLimit`** 链后，返回 **HTTP 501**，JSON 错误 **`code`** 为 **`NOT_IMPLEMENTED`**，**`message`** 说明该接口尚未实现，**`request_id`** 与 **`X-Request-ID`** 一致。
- 未携带有效网关凭证时仍先返回 **401**，不返回 501。
- **`DELETE /v1/models/:model`** 采用上述 501 语义，不再对 **`DELETE`** 使用「仅支持 GET」的 **405** 占位。

## 动机

区分「路由存在但能力未交付」与「完全未注册」，便于客户端与运维将 **501** 与 **404** 区分处理。
