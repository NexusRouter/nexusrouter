# saas-first-boot-initialization Specification

## Purpose
TBD - created by archiving change saas-first-boot-initialization. Update Purpose after archive.
## Requirements
### Requirement: 全局初始化状态权威源

系统 MUST 在数据库中维护**唯一**一条全局初始化状态记录（或等价单例聚合），字段 MUST 至少包含：**`initialized`**（布尔）、**`updated_at`**。该状态 MUST NOT 与租户或普通用户主键绑定；服务重启或水平扩容后 MUST 从数据库重新读取或经短 TTL 缓存刷新后保持一致语义。

#### Scenario: 重启后状态保持

- **WHEN** 系统已完成初始化且进程全部重启
- **THEN** 任意实例对状态查询接口返回 **`initialized=true`**

### Requirement: 初始化状态查询接口

系统 MUST 提供**无需登录**即可调用的初始化状态查询能力（HTTP），响应 MUST 包含：**`initialized`**（布尔）及可选 **`phase`**（如 `ready` | `initializing` | `completed`）。该接口 MUST 列入未初始化阶段的网关白名单。

#### Scenario: 未初始化时匿名可查询

- **WHEN** 客户端在未登录且系统未初始化时请求状态接口
- **THEN** 响应 **200** 且 body 中 **`initialized`** 为 **false**

#### Scenario: 已初始化后状态为真

- **WHEN** 客户端请求状态接口且数据库已标记完成初始化
- **THEN** 响应 **200** 且 **`initialized`** 为 **true**

### Requirement: 完成初始化事务与一次性锁定

完成初始化提交接口 MUST 在未初始化阶段**匿名可调用**（无管理令牌）。服务端 MUST 在**单个数据库事务**内完成：创建超级管理员（或等价账户）、持久化系统基础配置、将全局 **`initialized`** 置为 **true**；任一步失败 MUST **整事务回滚**且 **`initialized`** 保持 **false**。事务成功提交后，同一接口再次成功修改初始化数据 MUST **禁止**；后续调用 MUST 返回 **409**（或 **423**）及稳定机器可读 **`code`**（如 **`bootstrap_already_completed`**），且 MUST NOT 再次创建第二个超管或重复翻转标志。

#### Scenario: 首次成功初始化

- **WHEN** 首个合法完成初始化请求在空库默认状态下提交有效载荷
- **THEN** 事务提交成功，状态查询返回 **`initialized=true`**，且超管账户存在

#### Scenario: 重复成功初始化被拒绝

- **WHEN** 系统在 **`initialized=true`** 后再次调用完成初始化接口
- **THEN** 响应为客户端错误且 **`code`** 为稳定枚举 **`bootstrap_already_completed`**（或文档化等价），且数据库中超管数量与标志不被篡改

### Requirement: 初始化并发互斥

当多个客户端并发调用完成初始化时，MUST 保证**至多一个**事务可成功提交为已完成状态；其余并发请求 MUST 收到表示**冲突或进行中**的响应（如 **409** 与 **`code`** **`bootstrap_in_progress`** 或 **`bootstrap_conflict`**），且 MUST NOT 留下 **`initialized=true`** 与半创建超管并存的不可恢复状态。

#### Scenario: 双并发提交仅一成功

- **WHEN** 两个会话在未初始化状态下几乎同时提交合法且完整的初始化载荷
- **THEN** 恰好一个响应表示成功完成初始化，另一个响应表示冲突或进行中，且数据库最终 **`initialized=true`** 仅出现一次

### Requirement: 超级管理员重置为未初始化

系统 MUST 提供将全局状态恢复为**未初始化**的 HTTP 能力，且该能力 MUST 要求调用方具备**超级管理员**身份（有效凭证 + 角色校验）。成功执行后，状态查询 MUST 返回 **`initialized=false`**，且完成初始化接口 MUST 重新允许（在实现定义的清理策略完成后）匿名成功路径至多一次。重置操作 MUST 写入审计日志（结构化日志可接受为 MVP）。

#### Scenario: 非超管无法重置

- **WHEN** 已登录但非超级管理员调用重置接口
- **THEN** 响应 **403** 且 body 符合网关统一错误约定

#### Scenario: 超管重置成功

- **WHEN** 超级管理员调用重置接口且系统当前已初始化
- **THEN** 响应 **200** 或 **204**，且后续状态查询 **`initialized=false`**

### Requirement: 敏感数据存储

超级管理员密码及任何需保密的配置密钥 MUST NOT 以明文写入数据库。密码 MUST 使用单向密码散列（Argon2id 或 bcrypt）存储；若存在需解密使用的密钥材料，MUST 使用应用配置的对称密钥加密后存储。

#### Scenario: 数据库泄漏无明文密码

- **WHEN** 审查者仅读取初始化相关表行
- **THEN** 不存在与提交密码相同的可逆明文字段

