# Chat Completions 请求体 messages 必填与非空

## 功能说明

对 **`POST /v1/chat/completions`**：当请求体为可解析的 JSON 对象时，网关在鉴权通过后、转发上游前校验顶层 **`messages`** 须存在、须为 JSON 数组且至少含一项；**`null`**、缺省、空数组或非数组类型均返回 **HTTP 400**，错误码 **`INVALID_REQUEST`**，且不向上游转发。请求体无法解为对象时不因此规则单独拦截。

## 行为要点

- 与 **`max_tokens`** 等其它请求体校验并存；**`messages`** 为非空数组时，不因本规则单独拒绝。
