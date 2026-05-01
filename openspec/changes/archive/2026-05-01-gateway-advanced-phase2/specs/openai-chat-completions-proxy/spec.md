## MODIFIED Requirements

### Requirement: 上游目标可配置

网关 MUST 通过配置支持 **一个或多个** 上游 **基址 URL**；对 **POST `/v1/chat/completions`** 的转发目标 MUST 由 **`upstream-target-management`** 规范定义的选择逻辑（含 **`id`**、**`weight`**、**`default_upstream_id`**、**`active_upstream_id`** 等）从配置中选出 **一条** 上游条目，再将其 **`base_url`** 与 OpenAI 标准路径 **`/v1/chat/completions`** 按 **RFC 3986** 合并为绝对 URL。轮询（round-robin）MAY 作为兼容策略保留，但 MUST 与加权/默认/手动切换策略在 **`design.md`** 中说明优先级。环境变量形式的遗留列表 MUST 仍可映射为无 **`id`** 的等价条目集直至废弃窗口结束。

#### Scenario: 多上游加权分布

- **WHEN** 配置至少两条带正 **`weight`** 的上游且策略为加权随机
- **THEN** 长期抽样命中各 **`id`** 的比例 MUST 与权重一致（允许统计误差，测试可用固定随机种子）

#### Scenario: 手动切换后固定上游

- **WHEN** **`active_upstream_id`** 被设置为 **`b`** 且 **`b`** 合法
- **THEN** 后续成功鉴权的 **`POST /v1/chat/completions`** MUST 转发至 **`b`** 对应 **`base_url`**，直至配置再次变更

#### Scenario: 配置缺失时拒绝启动或拒绝转发

- **WHEN** 进程启动时 **所有** 上游基址均未设置或均非法
- **THEN** MUST 启动失败 **或** 对该路径返回 **503** 统一 JSON（二者择一并在实现中一致）；禁止向空 host 发起转发

#### Scenario: 合法配置下转发

- **WHEN** 所选上游基址为 `https://api.example.com` 且客户端请求 **POST `/v1/chat/completions`**
- **THEN** 上游收到的请求 URL MUST 指向 `https://api.example.com/v1/chat/completions`（若基址带 path 前缀，MUST 与 RFC 3986 路径合并规则一致）

## ADDED Requirements

### Requirement: 进阶中间件与代理链顺序

对 **`POST /v1/chat/completions`** 的处理链 MUST 遵循 **`design.md`** 固定顺序，且 MUST 在文档中体现 **CORS**、**RequestID**、**per-IP 限流**（若启用）、**`api-key-management` 鉴权**、**per-Key 限流**（若启用）、**`ChatProxy`** 的相对次序；**`OPTIONS`** 预检行为 MUST 与 **`http-cors`**、**`http-rate-limiting`** 规格一致。

#### Scenario: 限流先于上游调用

- **WHEN** 请求在任一限流阶段被拒绝
- **THEN** MUST **不**发起上游 HTTP 请求
