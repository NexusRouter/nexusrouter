# gateway-backend Specification

## Purpose
TBD - created by archiving change project-initialization. Update Purpose after archive.
## Requirements
### Requirement: Go 与 Web 框架版本

`services/gateway` MUST 使用 **Go 1.26.x** 工具链（`go` directive），并依赖 **github.com/gin-gonic/gin v1.10.0** 作为 HTTP 框架。

#### Scenario: 模块解析

- **WHEN** 在模块根执行 `go list -m github.com/gin-gonic/gin`
- **THEN** 输出主版本为 **v1.10.0**（允许补丁后缀以模块代理为准）

### Requirement: 后端依赖矩阵

模块 MUST 声明并可通过构建解析以下依赖（主模块或间接，以 `go.mod` 为准，版本约束与规范一致）：**gorm.io/gorm v1.25.x**、**gorm.io/driver/postgres v1.5.x**、**gorm.io/driver/sqlite v1.x**（或与 GORM 兼容的当前主版本）、**github.com/redis/go-redis/v9 v9.x**、**github.com/google/wire v0.6.x**、**github.com/spf13/viper v1.19.x**、**go.uber.org/zap v1.27.x**、**github.com/golang-jwt/jwt/v5 v5.x**、**github.com/go-playground/validator/v10 v10.x**、**github.com/golang-migrate/migrate/v4 v4.x**、**github.com/stretchr/testify v1.10.x**。

#### Scenario: 编译通过

- **WHEN** 在 `services/gateway` 执行 `go build ./...`
- **THEN** 以零退出码完成

### Requirement: 应用入口与监听端口

HTTP 服务 MUST 自 **`cmd/api`**（或等价的 `cmd/api/main.go`）启动，并监听 **8080** 端口（可配置但默认值 MUST 为 8080）。

#### Scenario: 本地启动

- **WHEN** 开发者运行 API 入口程序（如 `go run ./cmd/api`）
- **THEN** 进程在 **8080** 上接受连接（与现有健康检查路径可并存或迁移）

### Requirement: 分层目录结构

`services/gateway` MUST 具备以下目录：**`cmd/api/`**、**`internal/config/`**、**`internal/router/`**、**`internal/handler/`**、**`internal/service/`**、**`internal/repository/`**。

#### Scenario: 目录存在性检查

- **WHEN** 审查者列出上述路径
- **THEN** 各目录均存在且包含至少占位文件（如 `.gitkeep` 或包注释）以避免空包被工具链忽略

### Requirement: Zap 日志集成

服务启动路径 MUST 初始化 **zap** 日志（生产或开发合理配置），后续请求处理中的错误 MUST 具备统一记录到 Zap 的路径（骨架级即可）。

#### Scenario: 启动日志

- **WHEN** API 进程启动
- **THEN** Zap 向标准输出或配置目标写入至少一条表明服务已启动的日志

### Requirement: Wire 依赖注入

项目 MUST 采用 **google/wire** 组织依赖注入，并提供可生成的 `wire.go` 与生成物（如 `wire_gen.go`）或等价文档化生成步骤。

#### Scenario: 生成注入代码

- **WHEN** 开发者执行 `wire ./...`（在指定包路径）
- **THEN** 生成成功且无待提交的手改冲突（或 CI 文档化步骤通过）

### Requirement: 本地热重载工具

文档或脚本 MUST 说明使用 **github.com/air-verse/air** 进行本地热重载的安装方式（如 `go install github.com/air-verse/air@latest`）及配置文件位置。

#### Scenario: 文档可发现

- **WHEN** 新开发者阅读 `services/gateway` 下 README 或 `openspec/project.md` 链接
- **THEN** 能找到 air 的启动命令说明

### Requirement: 错误响应与日志约定

网关在处理 **自身生成的错误响应**（不包含上游透传的原始错误体）时 MUST 使用 **JSON**，且 body MUST 为对象并 MUST 包含：**`code`**（字符串，机器可读、稳定枚举）、**`message`**（字符串，人类可读）、**`request_id`**（字符串，与响应头 **`X-Request-ID`** 一致；若客户端请求已携带 **`X-Request-ID`**，MUST 使用该值）。错误路径 MUST 记录到 **Zap**，且日志 MUST 包含 **`request_id`** 与 **`code`**（或等价结构化字段），且 MUST NOT 记录完整 API Key 或 **`Authorization`** 头原文。

#### Scenario: 未知路由携带请求 ID

- **WHEN** 客户端请求未注册路由且请求头包含 **`X-Request-ID: probe-1`**
- **THEN** 响应为 JSON，其 **`request_id`** 为 **`probe-1`**，响应头 **`X-Request-ID`** 为 **`probe-1`**，且 Zap 记录该 **404**（或等价）事件并带相同 **`request_id`**

#### Scenario: 未带请求 ID 时服务端生成

- **WHEN** 客户端请求触发网关错误且未携带 **`X-Request-ID`**
- **THEN** 响应头 **`X-Request-ID`** 非空，JSON body 中 **`request_id`** 与该头一致，且 Zap 含相同 **`request_id`**

#### Scenario: Panic 恢复响应含请求 ID

- **WHEN** 处理链触发 panic 并由恢复中间件转换为 **500** JSON
- **THEN** 该 JSON MUST 包含 **`request_id`** 且与 **`X-Request-ID`** 响应头一致

### Requirement: 请求 ID 中间件全局启用

HTTP 引擎 MUST 在业务路由之前注册 **`RequestID`**（或等价命名）中间件：若请求已带 **`X-Request-ID`**，MUST 透传该值；否则 MUST 生成并写入响应头 **`X-Request-ID`**；且 MUST 将当前请求 ID 存入 Gin 上下文供后续处理器与错误封装读取。

#### Scenario: 下游处理器可读 request_id

- **WHEN** 任意已注册业务处理器执行
- **THEN** 其可通过上下文读取与 **`X-Request-ID`** 一致的请求 ID 字符串

### Requirement: 文档与提交语言

后端源码中的 **注释** MUST 使用 **中文**；导出标识符与包名 MUST 使用 **英文**；Git commit message MUST 使用 **中文**。

#### Scenario: 代码审查抽样

- **WHEN** 审查者打开 `cmd/api` 与 `internal` 下主要文件
- **THEN** 注释为中文且 API 命名为英文

### Requirement: Chat 路径逻辑模型解析与模型实例选择

在 **POST `/v1/chat/completions`** 处理链中，当启用模型库聚合时，网关 SHALL 从 JSON body 读取 **`model`**，解析 **`model_base.model_code`**，在 **`model_instance`** 上按 **`model-library-aggregation`** 与 **`model-library`** 所载选择策略（各表 **`status=1`**、**`priority`、`weight`、`is_official`**；**不**含上游 HTTP 探针临时剔除，见 **`model-library`**）选中 **一条实例**，使用其关联 **`model_upstream`** 的 **`base_url`** 作为转发根、**`api_key`** 作为上游鉴权、**`timeout`/`max_concurrent`** 参与客户端/队列策略（以实现为准）。**本路径 MUST NOT 再使用 `gateway.yaml` 中声明的 Upstream 列表参与同一请求的寻址或回退**（与静态配置**不并存**，单一事实来源为四表）。若无可用实例，SHALL 返回与 **`gateway-backend`** 错误约定一致的响应，且 MUST NOT 泄露内部表名或 **`api_key`**。

#### Scenario: 选择结果可追踪

- **WHEN** 请求成功转发
- **THEN** 结构化日志 MAY 记录 **`model_instance.id`**，且 MUST NOT 记录 **`api_key`**

#### Scenario: 无可用实例

- **WHEN** 逻辑模型存在但无 **`status=1`** 的可用链
- **THEN** 响应为 **4xx**（实现选定）且 **`code`** 稳定可枚举

### Requirement: 模型库管理路由注册

`services/gateway` SHALL 在管理控制台已启用且鉴权链完整时，于 **`/api/admin/v1/model-library/**`**（或文档化之等价前缀）下注册足以支撑 **`model_vendor`、`model_base`、`model_upstream`、`model_instance`** 的管理端点（含列表/创建/更新/删除及上游 **`/v1/models`** 同步触发等，以实现为准）；路由 MUST 复用现有 **`adminJWTMiddleware`** 与写保护策略，且错误体符合既有 JSON 约定。

#### Scenario: 路由可发现

- **WHEN** 审查者检索路由注册表或 OpenAPI
- **THEN** 可见厂商/逻辑模型/上游/实例相关路径与 HTTP 方法映射

### Requirement: 模型库数据库迁移

模型库相关表 MUST 通过项目约定的迁移机制（GORM AutoMigrate 或 SQL 迁移）创建；迁移 MUST 在 **SQLite 与 Postgres** 目标下均可应用（与现有 `repository` 策略一致）。

#### Scenario: 全新数据库启动

- **WHEN** 进程首次连接空库并执行迁移
- **THEN** 模型库表存在且 `go test` 或冒烟脚本通过

### Requirement: 错误响应一致性

模型库处理器在失败时 MUST 返回与 **`gateway-backend`** 既有要求一致的 JSON 错误体（**`code`/`message`/`request_id`**），且 MUST 记录 Zap（含 **`request_id`**）。

#### Scenario: 客户端错误

- **WHEN** 提交非法 JSON 或违反校验
- **THEN** 响应 **400** 且 body 含稳定 **`code`**

### Requirement: 首次启动门闸中间件

`services/gateway` MUST 在未初始化全局状态下，对除**明确白名单**外的所有 HTTP 请求返回 **403** 或 **503**（实现选定其一并在 **`design.md`** 或 **README** 中固定），且 body MUST 符合既有 **`gateway-backend`** JSON 错误约定（含 **`code`**、**`message`**、**`request_id`**）。白名单 MUST 至少包含：**初始化状态查询**、**完成初始化提交**、**健康检查**（若项目已存在标准健康路径）、**静态资源**（若由网关直出）。**已初始化**状态下，白名单逻辑 MUST 不再拦截正常业务与管理 API。

#### Scenario: 未初始化访问非白名单 API

- **WHEN** 系统 **`initialized=false`** 且客户端请求任意非白名单已注册路由（如管理业务 API）
- **THEN** 响应为 JSON 错误且 **`code`** 为稳定枚举（如 **`bootstrap_required`**），且 **`request_id`** 与头一致

#### Scenario: 已初始化不触发门闸拒绝

- **WHEN** 系统 **`initialized=true`** 且客户端请求普通管理或代理 API
- **THEN** MUST NOT 因本门闸单独返回 **`bootstrap_required`**

### Requirement: 初始化与重置路由注册

初始化状态查询、完成初始化提交、超级管理员重置 MUST 注册为独立 handler，路径前缀 MUST 在 **`design.md`** 或 **README** 中固定（建议 **`/api/bootstrap`** 或等价）。完成初始化与状态查询在未初始化时 MUST 不要求 Bearer 令牌；重置 MUST 要求认证与超管角色。

#### Scenario: 重置未携带令牌

- **WHEN** 客户端不带认证调用重置接口
- **THEN** 响应 **401** 且 JSON 错误格式符合规范

### Requirement: 管理端 HTTP API 与鉴权

`services/gateway` MUST 暴露一组**管理用途**的 HTTP API（路径前缀在实现中固定，如 **`/admin` 或 `/api/admin`**），用于支撑控制台的登录、指标查询、上游配置与 API Key 管理。除登录/重置密码等公开端点外，其余管理端点 MUST 要求**有效管理员权限令牌**（或等价会话），未携带或无效时 MUST 返回 **401** 且 body 符合 **`gateway-backend`** 既有 JSON 错误约定（含 **`code`**、**`message`**、**`request_id`**）。

#### Scenario: 无令牌访问受保护管理 API

- **WHEN** 客户端调用需认证的管理 API 且不携带有效凭证
- **THEN** 响应为 **401**，且 body 为统一 JSON 错误格式

### Requirement: 网关运行指标对外查询

网关 MUST 提供管理端可消费的**指标查询接口**（或单一聚合端点），使 **`admin-dashboard-metrics`** 所需之在线状态、请求量、成功率、平均耗时、今日/昨日对比与错误 **`code`** 聚合可被前端拉取或订阅。指标覆盖范围 MUST 在 `design.md` 声明；若某指标暂不可用 MUST 明确返回占位或省略策略而非伪造数据。

#### Scenario: 指标端点需管理员权限

- **WHEN** 未认证客户端请求指标端点
- **THEN** MUST **不**返回可识别业务的详细统计（**401**）

### Requirement: 上游与密钥管理 API 与运行时一致

管理 API 对 **`gateway.yaml`** 与 **API Key JSON** 的写入 MUST 与 `design.md` 一致：写后 MUST 使 **`KeyStore`** / **`runtime.Store`** 进入与磁盘一致的有效状态；失败时 MUST 保留先前快照并返回错误。

#### Scenario: 写文件失败不损坏旧快照

- **WHEN** 持久化配置写入失败
- **THEN** 运行时继续服务先前有效配置，且响应指示失败原因

### Requirement: 进阶管理 HTTP API 面

在 **`/api/admin/v1`**（或文档化之等价前缀）下 MUST 增加以下能力对应的受 JWT 保护端点：**日志查询与 CSV 导出**、**限流规则 CRUD**、**CORS 配置读写与批量域名**、**IP 名单读写与批量操作**。未认证访问 MUST 返回 **401** 且 body 符合既有网关 JSON 错误约定。

#### Scenario: 导出接口需管理员权限

- **WHEN** 匿名客户端请求日志 CSV 导出 URL
- **THEN** MUST **不**返回文件内容且为 **401**

### Requirement: 请求链整合 IP 名单

在 **IP 限流** 与 **鉴权** 之间的精确顺序在 `design.md` 固定；IP 名单拒绝 MUST 在调用上游前发生，且 MUST 记录结构化日志（含 **`request_id`**），且 MUST NOT 记录完整 API Key。

#### Scenario: 名单拒绝不产生上游流量

- **WHEN** 请求被黑名单或白名单缺省拒绝
- **THEN** MUST **不**对上游发起 **`POST /v1/chat/completions`**

### Requirement: 管理端系统设置 API

在 **`/api/admin/v1`** 下 MUST 增加受 JWT 保护的 **系统设置** 端点（路径以实现为准，如 `GET/PUT /api/admin/v1/system/settings`），语义 MUST 满足 `admin-system-settings` 能力规范；写入 MUST 经校验且失败时保留旧配置。

#### Scenario: 未认证访问设置

- **WHEN** 匿名请求系统设置 GET
- **THEN** 返回 **401**

### Requirement: 管理端 RBAC 强制校验

所有 **非只读** 管理写接口（含密钥、网关配置、安全策略、CORS、限流、IP 名单、系统设置写等，以实现维护的清单为准）MUST 在校验 JWT 后检查 **`role`**；当 `role=operator` 时 MUST 返回 **403**。只读 GET 清单 MUST 在 `design.md` 或实现代码注释中维护并与 `admin-rbac` 一致。

#### Scenario: 操作员 PUT 网关配置被拒绝

- **WHEN** `operator` 调用 `PUT /api/admin/v1/gateway/config`
- **THEN** 返回 **403**

### Requirement: 管理端告警状态 API

MUST 提供 `GET /api/admin/v1/alerts/status`（或等价路径）返回 `admin-runtime-alerts` 规范定义的状态与 `reasons`；实现 MUST 基于进程内指标与配置阈值计算，MUST NOT 在响应中泄露完整 API Key 或 `Authorization` 原文。

#### Scenario: 操作员可读告警

- **WHEN** `operator` 调用告警状态 GET
- **THEN** 返回 **200** 且 body 符合规范

