## ADDED Requirements

### Requirement: 上游条目与标识

网关 MUST 为每个上游配置条目分配稳定 **`id`**（字符串，在配置文件中唯一），并包含 **`base_url`**（RFC3986 合法绝对基址）及非负 **`weight`**（**0** 表示不参与随机选择但仍可被 **pinned** 指向）。MUST 支持声明 **`default_upstream_id`**：当权重选择未命中或未配置权重时，回退到该默认条目。

#### Scenario: 缺省默认且全权重为正

- **WHEN** 配置存在多条上游且 **`default_upstream_id`** 指向其中一条
- **THEN** 在 **`weighted-random`**（或 `design.md` 命名的等价策略）下，长期抽样分布 MUST 与权重比例一致（允许统计误差，由测试固定种子或 mock 随机源验证）

#### Scenario: 非法 id 引用

- **WHEN** **`default_upstream_id`** 或 **`active_upstream_id`** 引用不存在的 **`id`**
- **THEN** MUST 启动失败 **或** 拒绝加载新配置并保留上一快照（二者在实现与 README 中一致）

### Requirement: 运行时切换当前上游

网关 MUST 支持将 **当前生效上游** 固定为指定 **`id`**（**pinned** 模式），以适配手动切流；切换可通过 **热更新配置文件** 中的 **`active_upstream_id`** 字段和/或 **`PUT /internal/upstream/active`**（若实现）完成。解除 pin 后 MUST 恢复 **`design.md`** 声明的策略（如加权随机）。

#### Scenario: 热加载切换生效

- **WHEN** 运维将 **`active_upstream_id`** 从 **A** 改为 **B** 并触发配置热加载
- **THEN** 后续 **`POST /v1/chat/completions`** 在未命中其他例外时 MUST 转发至 **B** 对应基址

#### Scenario: 管理 API 切换需鉴权

- **WHEN** 客户端调用 **`PUT /internal/upstream/active`** 且未携带有效 **管理 Bearer 令牌**
- **THEN** 响应为 **401**，且 MUST **不**修改当前上游选择状态

### Requirement: 与 Chat 代理路径的衔接

**`POST /v1/chat/completions`** 解析上游 URL 时 MUST 使用 **`upstream-target-management`** 选出的条目之 **`base_url`** 与 **`/v1/chat/completions`** 按 **RFC 3986** 合并；该选择逻辑 MUST 在鉴权通过之后、发起上游 HTTP 请求之前执行。

#### Scenario: 鉴权失败不改变上游计数语义

- **WHEN** 请求未通过鉴权
- **THEN** MUST **不**因该请求推进任何上游轮询/加权随机内部状态（若适用）
