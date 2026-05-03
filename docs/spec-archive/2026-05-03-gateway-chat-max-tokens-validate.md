# Chat Completions 请求体 max_tokens 校验

## 功能说明

对 **`POST /v1/chat/completions`**：当请求体为可解析的 JSON 对象且顶层包含 **`max_tokens`**（且不为 **`null`**）时，网关在鉴权通过后、转发上游前校验该字段须为 JSON 数字、非负整数，且不大于 **`MaxInt32/2`**。不满足时返回 **HTTP 400**，错误码 **`INVALID_REQUEST`**，且不向上游转发。

## 行为要点

- 省略 **`max_tokens`** 或值为 **`null`** 时不因此规则拒绝。
- 非对象 JSON（无法解为对象）不因本规则单独拦截。
