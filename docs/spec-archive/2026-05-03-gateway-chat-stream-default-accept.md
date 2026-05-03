# Chat 流式缺省 Accept

对 **`POST /v1/chat/completions`**：若 JSON 顶层 **`stream`** 为 **`true`** 且客户端未提供非空 **`Accept`**，网关在转发至上游前为出站请求设置 **`Accept: text/event-stream`**，以便上游按服务端推送事件流协商响应。若客户端已带非空 **`Accept`**，或 **`stream`** 不为 **`true`**，则不修改 **`Accept`**。
