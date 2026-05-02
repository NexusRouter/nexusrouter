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

模块 MUST 声明并可通过构建解析以下依赖（主模块或间接，以 `go.mod` 为准，版本约束与规范一致）：**gorm.io/gorm v1.25.x**、**gorm.io/driver/postgres v1.5.x**、**gorm.io/driver/sqlite v1.x**（或与 GORM 兼容的当前主版本）、**github.com/redis/go-redis/v9 v9.x**、**github.com/google/wire v0.6.x**、**github.com/spf13/viper v1.19.x**、**go.uber.org/zap v1.27.x**、**github.com/golang-jwt/jwt/v5 v5.x**、**github.com/go-playground/validator/v10 v10.x**、**github.com/golang-migrate/migrate/v4 v4.x**、**github.com/stretchr/testify v1.10.x**、**github.com/gin-contrib/gzip v1.x**。

#### Scenario: 编译通过

- **WHEN** 在 `services/gateway` 执行 `go build ./...`
- **THEN** 以零退出码完成

### Requirement: 应用入口与监听端口

HTTP 服务 MUST 自 **`cmd/api`**（或等价的 `cmd/api/main.go`）启动，并监听 **8080** 端口（可配置但默认值 MUST 为 8080）。

#### Scenario: 本地启动

- **WHEN** 开发者运行 API 入口程序（如 `go run ./cmd/api`）
- **THEN** 进程在 **8080** 上接受连接（与现有健康检查路径可并存或迁移）

### Requirement: 通用环境变量 PORT 推导监听地址

当 **`NEXUSROUTER_HTTP_LISTEN_ADDR`** 未设置或仅空白时，若进程环境 **`PORT`** 为非空白字符串，**`config.Load`** MUST 将其纳入 **`HTTPListenAddr`** 解析：若 **`PORT`** 已含冒号（**`host:port`** 等形式），MUST 将该字符串用作监听地址；否则 MUST 将 **`PORT`** 视为十进制端口并在前加 **`:`**（如 **`3000`** → **`:3000`**）。当 **`NEXUSROUTER_HTTP_LISTEN_ADDR`** 为非空白时，MUST NOT 使用 **`PORT`** 覆盖之。二者皆缺省时默认仍为 **`:8080`**（见上一节）。

#### Scenario: 仅设置 PORT

- **WHEN** **`NEXUSROUTER_HTTP_LISTEN_ADDR`** 未设置或为空，且 **`PORT`** 为 **`3000`**
- **THEN** **`HTTPListenAddr`** 为 **`:3000`**

#### Scenario: 显式监听地址优先

- **WHEN** **`NEXUSROUTER_HTTP_LISTEN_ADDR`** 为 **`:9090`**，且 **`PORT`** 为 **`3000`**
- **THEN** **`HTTPListenAddr`** 为 **`:9090`**

### Requirement: 入口命令行 `-version`

`cmd/api` 进程在加载 **`.env`**、初始化数据库或监听端口之前，MUST 解析标准命令行参数 **`-version`**（布尔标志）。当该标志为真时，进程 MUST 向标准输出写入一行 **`internal/buildinfo`** 包中 **`Version`** 的当前取值（与 **`GET /health`** 的 **`version`** 字段语义一致），然后以退出码 **0** 终止，且 MUST NOT 启动 HTTP 服务。

#### Scenario: 仅查询版本

- **WHEN** 以 **`cmd/api`** 可执行文件传入 **`-version`** 启动
- **THEN** 标准输出恰好一行版本字符串，进程退出码为 **0**，且无监听套接字被打开

### Requirement: 入口命令行 `-help` / `-h`

`cmd/api` 进程在加载 **`.env`**、初始化数据库或监听端口之前，MUST 解析标准命令行参数 **`-help`** 与 **`-h`**（布尔标志，语义等价）。当任一标志为真且未同时传入为真的 **`-version`** 时，进程 MUST 向标准输出写入简短用法说明（含可执行文件名、**`-version`** 与帮助类标志说明），然后以退出码 **0** 终止，且 MUST NOT 启动 HTTP 服务。若 **`-version`** 与其它标志同时出现且 **`-version`** 为真，MUST 按 **`-version`** 行为处理（见上一节）。

#### Scenario: 仅查询用法

- **WHEN** 以 **`cmd/api`** 可执行文件传入 **`-help`** 或 **`-h`** 启动
- **THEN** 标准输出包含用法说明，进程退出码为 **0**，且无监听套接字被打开

### Requirement: 可选工作目录 `.env` 加载

API 进程在读取 **`internal/config`** 等启动配置之前，MUST 尝试自当前工作目录加载名为 **`.env`** 的 dotenv 文件，将其中定义的键合并进进程环境；若文件不存在或读取失败，MUST 静默忽略且不改变其它启动语义。对已存在于操作系统环境中的键，**`.env`** 中的同名项 MUST NOT 覆盖之（进程外已注入的变量优先）。

#### Scenario: `.env` 中的 NEXUSROUTER 键生效

- **WHEN** 工作目录存在 **`.env`** 且包含 **`NEXUSROUTER_*`** 键，且对应键在启动前未在 OS 环境中设置
- **THEN** 后续 **`config.Load`** 等逻辑读取到的值与 **`.env`** 中声明一致

### Requirement: 入口命令行 `-port`（在 `.env` 合并之后参与推导）

`cmd/api` MUST 在启动参数中接受 **`-port <整数>`**（十进制，范围 **`1`**–**`65535`**；未传或值为 **`0`** 表示未使用该标志）。**`-version`** 与 **`-help`** / **`-h`** 的解析与退出行为 MUST 仍发生在加载工作目录 **`.env`** 之前（见前述要求）。在 **`.env`** 已按「可选工作目录 `.env` 加载」规则合并进环境之后，若此时 **`NEXUSROUTER_HTTP_LISTEN_ADDR`** 与 **`PORT`** 经首尾空白去除后仍均为空，进程 MUST 将环境变量 **`PORT`** 设为该整数的十进制字符串，以便后续 **`config.Load`** 与「通用环境变量 PORT 推导监听地址」要求一致；若二者任一已非空，MUST 忽略 **`-port`**。当 **`-port`** 超出 **`1`**–**`65535`** 时，进程 MUST 向标准错误输出说明并以非零退出码终止，且 MUST NOT 启动 HTTP 服务。

#### Scenario: 无监听环境变量且 `.env` 未提供 PORT 时 `-port` 生效

- **WHEN** 合并 **`.env`** 后 **`NEXUSROUTER_HTTP_LISTEN_ADDR`** 与 **`PORT`** 仍均为空，且命令行含 **`-port`** **`3000`**
- **THEN** **`config.Load`** 得到的 **`HTTPListenAddr`** 为 **`:3000`**

#### Scenario: 显式监听地址仍优先于 `-port`

- **WHEN** 环境 **`NEXUSROUTER_HTTP_LISTEN_ADDR`** 为 **`:9090`**，命令行含 **`-port`** **`3000`**
- **THEN** **`HTTPListenAddr`** 为 **`:9090`**

#### Scenario: `.env` 中的 PORT 优先于 `-port`

- **WHEN** **`.env`** 已将 **`PORT`** 设为 **`8080`**（且 **`NEXUSROUTER_HTTP_LISTEN_ADDR`** 为空），命令行含 **`-port`** **`3000`**
- **THEN** **`HTTPListenAddr`** 为 **`:8080`**

### Requirement: 持久化日志目录（环境变量与 `-log-dir`）

当 **`NEXUSROUTER_LOG_DIR`** 为非空白字符串时，**`config.Load`** MUST 将其首尾空白去除后写入 **`LogDir`**。**`config.Load`** MUST 将布尔环境变量 **`NEXUSROUTER_LOG_DAILY_FILE`**（未设置时视为 **`false`**）解析为 **`LogDailyFile`**。进程初始化 Zap 时 MUST 在该路径下创建目录（若不存在）、并向其中持久化 JSON 日志文件以 **`Info`** 及以上级别追加写入，同时 MUST 继续向标准错误输出开发风格日志（与未配置 **`LogDir`** 时一致为可读的彩色控制台输出）。当 **`LogDailyFile`** 为 **`false`** 时，持久化文件名 MUST 为 **`gateway.log`**；当为 **`true`** 时，文件名 MUST 为 **`gateway-YYYYMMDD.log`**（**`YYYYMMDD`** 为进程初始化 Zap 时的本地日历日）。**`cmd/api`** MUST 接受 **`-log-dir <路径>`**：在加载工作目录 **`.env`** 之后、**`config.Load`** 之前，若此时 **`NEXUSROUTER_LOG_DIR`** 在环境中仍为空，进程 MUST 将 **`NEXUSROUTER_LOG_DIR`** 设为该路径的绝对路径；若 **`NEXUSROUTER_LOG_DIR`** 已由环境或 **`.env`** 设置，MUST 忽略 **`-log-dir`**。

#### Scenario: 仅环境变量指定目录

- **WHEN** **`NEXUSROUTER_LOG_DIR`** 为已存在的可写目录的绝对路径，且未将 **`NEXUSROUTER_LOG_DAILY_FILE`** 设为真
- **THEN** 进程运行期间 **`gateway.log`** 可被创建或追加，且包含至少一条结构化日志记录

#### Scenario: 按日文件名

- **WHEN** **`NEXUSROUTER_LOG_DIR`** 为非空白且 **`NEXUSROUTER_LOG_DAILY_FILE`** 为真（如 **`true`** / **`1`**）
- **THEN** 持久化 JSON 日志写入 **`gateway-YYYYMMDD.log`**（与初始化当日一致），且 **`config.Load`** 得到的 **`LogDailyFile`** 为 **`true`**

#### Scenario: 环境变量优先于 `-log-dir`

- **WHEN** **`.env`** 或操作系统环境已将 **`NEXUSROUTER_LOG_DIR`** 设为 **`/a`**，且命令行含 **`-log-dir`** **`/b`**
- **THEN** **`config.Load`** 得到的 **`LogDir`** 为 **`/a`**

### Requirement: Gin 运行模式环境变量

构造 HTTP 引擎时 MUST 根据进程环境变量 **`GIN_MODE`** 调用 **`gin.SetMode`**：**`GIN_MODE`** 经首尾空白去除后等于 **`debug`**（与 **`github.com/gin-gonic/gin`** 的 **`gin.DebugMode`** 一致）时 MUST 为 **`gin.DebugMode`**；否则（含未设置、空串或其它取值）MUST 为 **`gin.ReleaseMode`**。

#### Scenario: 未设置 GIN_MODE 时为发布模式

- **WHEN** 进程未设置 **`GIN_MODE`** 且构造网关引擎
- **THEN** **`gin.Mode()`** 为 **`gin.ReleaseMode`**

#### Scenario: GIN_MODE 为 debug 时为调试模式

- **WHEN** 进程设置 **`GIN_MODE=debug`**（允许首尾空白）且构造网关引擎
- **THEN** **`gin.Mode()`** 为 **`gin.DebugMode`**

### Requirement: 健康检查 GET 与 HEAD

网关 MUST 注册 **`GET /health`** 与 **`HEAD /health`**，二者在成功时 HTTP 状态码 MUST 均为 **200**，且 JSON 字段语义与 **`GET`** 一致（**`status`**、**`version`**、**`start_time`**、**`uptime_seconds`**、**`server_time`**）。**`HEAD`** 响应 MUST NOT 携带消息体；MUST 设置 **`Content-Type: application/json; charset=utf-8`**，且 SHOULD 设置 **`Content-Length`** 为与同负载 **`GET`** 响应体等价的 UTF-8 JSON 字节长度。

#### Scenario: GET 返回 JSON 体

- **WHEN** 客户端 **`GET /health`**
- **THEN** 状态码为 **200**，body 为 JSON 对象且 **`status`** 为 **`ok`**，**`version`** 非空

#### Scenario: HEAD 无体且可探活

- **WHEN** 客户端 **`HEAD /health`**
- **THEN** 状态码为 **200**，响应体长度为 **0**，**`Content-Type`** 为 **`application/json; charset=utf-8`**

### Requirement: 公开 GET /api/status

网关 MUST 注册 **`GET /api/status`**，MUST NOT 要求 **`GatewayAuth`** 或其它登录态。成功时 HTTP 状态码 MUST 为 **200**，body MUST 为 JSON 对象，且 MUST 包含布尔 **`success`**（值为 **`true`**）、字符串 **`message`**（成功时为空字符串）、对象 **`data`**。**`data`** MUST 包含非空字符串 **`version`**（与 **`internal/buildinfo`** 的 **`Version`** 一致）与非空字符串 **`start_time`**（当前进程启动时刻，**RFC3339Nano**）。在未完成首次系统初始化、**`BootstrapGate`** 拦截其它路径时，**`GET /api/status`** MUST 仍可达（与 **`GET /health`** 同为引导期可读路径）。

#### Scenario: GET /api/status 返回封装 JSON

- **WHEN** 客户端 **`GET /api/status`**
- **THEN** 状态码为 **200**，**`success`** 为 **`true`**，**`data.version`** 非空，**`data.start_time`** 可被解析为 **RFC3339Nano** 时间戳

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

网关在处理 **自身生成的错误响应**（不包含上游透传的原始错误体）时 MUST 使用 **JSON**，且 body MUST 为对象并 MUST 包含：**`code`**（字符串，机器可读、稳定枚举）、**`message`**（字符串，人类可读）、**`request_id`**（字符串，与响应头 **`X-Request-ID`** 一致；若客户端请求已携带 **`X-Request-ID`**，MUST 使用该值；若未携带 **`X-Request-ID`** 但存在「请求 ID 中间件全局启用」所定义的备用入站头，MUST 按该中间件与主头相同的优先规则确定 **`request_id`**）。错误路径 MUST 记录到 **Zap**，且日志 MUST 包含 **`request_id`** 与 **`code`**（或等价结构化字段），且 MUST NOT 记录完整 API Key 或 **`Authorization`** 头原文。

#### Scenario: 未知路由携带请求 ID

- **WHEN** 客户端请求未注册路由且请求头包含 **`X-Request-ID: probe-1`**
- **THEN** 响应为 JSON，其 **`request_id`** 为 **`probe-1`**，响应头 **`X-Request-ID`** 为 **`probe-1`**，且 Zap 记录该 **404**（或等价）事件并带相同 **`request_id`**

#### Scenario: 未带请求 ID 时服务端生成

- **WHEN** 客户端请求触发网关错误且未携带 **`X-Request-ID`**，亦未携带「请求 ID 中间件」所识别的备用入站头
- **THEN** 响应头 **`X-Request-ID`** 非空，JSON body 中 **`request_id`** 与该头一致，且 Zap 含相同 **`request_id`**

#### Scenario: Panic 恢复响应含请求 ID

- **WHEN** 处理链触发 panic 并由恢复中间件转换为 **500** JSON
- **THEN** 该 JSON MUST 包含 **`request_id`** 且与 **`X-Request-ID`** 响应头一致

#### Scenario: Panic 恢复时 Zap 含 HTTP 方法与路径

- **WHEN** 处理链触发 panic 且 **`ZapRecovery`** 写入 **`Error`** 级别日志
- **THEN** 该条记录 MUST 包含 **`method`**（请求 HTTP 方法）与 **`path`**（优先 Gin **`FullPath()`** 路由模板，若为空则 **`URL.Path`**），且 MUST NOT 包含请求体原文

### Requirement: 显式列出的 OpenAI 兼容 v1 未实现路径

对实现维护的一组 **OpenAI 兼容** 子路径（如旧版 **`POST /v1/completions`**（非 Chat）、Files、Fine-tuning、Assistants、Threads、按 id 删除模型等常见平台路径），网关 MUST 在通过 **`GatewayAuth`** 与 **`KeyRateLimit`**（与 **`POST /v1/chat/completions`** 相同的鉴权与按 key 限流链）后，对表中声明的 HTTP 方法返回 **501 Not Implemented**，且响应 body MUST 符合本节上文 **错误响应与日志约定**（**`code`** 为 **`NOT_IMPLEMENTED`**）。本要求不适用于已实现的路径（**`POST /v1/chat/completions`**、**`POST /v1/embeddings`**、**`POST /v1/engines/:model/embeddings`**、**`POST /v1/moderations`**、**`POST /v1/images/generations`**、**`POST /v1/audio/speech`**、**`GET /v1/models`**、**`GET /v1/models/:model`**）。对 **`DELETE /v1/models/:model`**，MUST 使用本规则的 **501** 语义，而 MUST NOT 再使用「仅支持 GET」的 **405** 占位。

#### Scenario: 已列出的未实现子路径返回 501

- **WHEN** 客户端使用有效网关凭证请求已注册的未实现子路径（如 **`POST /v1/completions`** 或 **`POST /v1/images/edits`**）
- **THEN** HTTP 状态码为 **501**，JSON **`code`** 为 **`NOT_IMPLEMENTED`**

#### Scenario: 未带凭证时不返回 501

- **WHEN** 客户端请求上述子路径但不带有效 API 密钥
- **THEN** MUST 返回 **401** 及统一未授权 **`code`**（**MUST NOT** 返回 **501**）

### Requirement: OpenAI 兼容 Embeddings 反向代理

网关 MUST 注册 **`POST /v1/embeddings`** 与 **`POST /v1/engines/:model/embeddings`**；对 **`GET/PUT/PATCH/DELETE/HEAD`** 等非 **POST** 方法（**`OPTIONS`** 除外）MUST 返回 **405**，body 为网关统一 JSON（**`code`** 为 **`METHOD_NOT_ALLOWED`**）。**`OPTIONS`** 对上述两路径 MUST 返回 **204 No Content**（与 **`/v1/chat/completions`** 预检占位语义一致）。在通过 **`GatewayAuth`** 与 **`KeyRateLimit`** 后，网关 MUST 将请求反向代理至与 Chat 代理一致的上游选择结果（模型库启用时按逻辑 **`model`** 选实例并改写 **`model`**；否则按运行时快照 **`Picker`** 与绑定改写）；出站 MUST 剔除与 Chat 代理相同的 hop-by-hop 头及 **`Accept-Encoding`**，且 **`Host`** 语义与 Chat 一致。请求体 MUST 为 JSON 对象且含非空 **`input`**（键须存在且值不可为 **`null`**）；若 JSON 未提供非空 **`model`** 且路径为 **`/v1/engines/:model/embeddings`**，MUST 将路径中的 **`model`** 写入 body 后再校验与转发。

#### Scenario: POST /v1/embeddings 鉴权后转发

- **WHEN** 客户端 **`POST /v1/embeddings`** 带有效网关凭证、合法 JSON（含 **`input`** 与 **`model`** 或等价可解析字段）、且上游已配置
- **THEN** 上游收到与合并规则一致后的请求，且 HTTP 状态码为上游结果或网关 **`BAD_GATEWAY`** 等已文档化错误

#### Scenario: 引擎路径补全 model

- **WHEN** 客户端 **`POST /v1/engines/my-engine/embeddings`** 的 JSON 含 **`input`** 但省略 **`model`**
- **THEN** 转发前 body MUST 含 **`model`** 为路径段 **`my-engine`**

### Requirement: OpenAI 兼容 Moderations 反向代理

网关 MUST 注册 **`POST /v1/moderations`**；对 **`GET/PUT/PATCH/DELETE/HEAD`** 等非 **POST** 方法（**`OPTIONS`** 除外）MUST 返回 **405**，body 为网关统一 JSON（**`code`** 为 **`METHOD_NOT_ALLOWED`**）。**`OPTIONS`** 对该路径 MUST 返回 **204 No Content**（与 **`/v1/embeddings`** 预检占位语义一致）。在通过 **`GatewayAuth`** 与 **`KeyRateLimit`** 后，网关 MUST 将请求反向代理至与 Embeddings 代理一致的上游选择结果（模型库启用时按逻辑 **`model`** 选实例并改写 **`model`**；否则按运行时快照 **`Picker`** 与绑定改写）；出站 MUST 剔除与 Chat 代理相同的 hop-by-hop 头及 **`Accept-Encoding`**，且 **`Host`** 语义与 Chat 一致。请求体 MUST 为 JSON 对象且含合法 **`input`**（键须存在、值不可为 **`null`**；若为字符串则去首尾空白后非空；若为 JSON 数组则须至少一项；若为 JSON 对象则须至少一个键）。若 JSON 未提供非空 **`model`**（缺键、**`null`** 或空字符串），MUST 在转发前合并默认 **`model`** 为 **`text-moderation-latest`**。

#### Scenario: POST /v1/moderations 鉴权后转发

- **WHEN** 客户端 **`POST /v1/moderations`** 带有效网关凭证、合法 JSON（含 **`input`**，**`model`** 可省略）、且上游已配置
- **THEN** 上游收到与合并规则一致后的请求，且 HTTP 状态码为上游结果或网关 **`BAD_GATEWAY`** 等已文档化错误

#### Scenario: 省略 model 时合并默认 moderation 模型名

- **WHEN** 客户端 **`POST /v1/moderations`** 的 JSON 含非空 **`input`** 且省略 **`model`** 或 **`model`** 为空字符串
- **THEN** 转发前 body MUST 含非空 **`model`**，且为 **`text-moderation-latest`**

### Requirement: OpenAI 兼容图像生成反向代理

网关 MUST 注册 **`POST /v1/images/generations`**；对 **`GET/PUT/PATCH/DELETE/HEAD`** 等非 **POST** 方法（**`OPTIONS`** 除外）MUST 返回 **405**，body 为网关统一 JSON（**`code`** 为 **`METHOD_NOT_ALLOWED`**）。**`OPTIONS`** 对该路径 MUST 返回 **204 No Content**（与 **`/v1/moderations`** 预检占位语义一致）。在通过 **`GatewayAuth`** 与 **`KeyRateLimit`** 后，网关 MUST 将请求反向代理至与 Moderations 代理一致的上游选择结果（模型库启用时按逻辑 **`model`** 选实例并改写 **`model`**；否则按运行时快照 **`Picker`** 与绑定改写）；出站 MUST 剔除与 Chat 代理相同的 hop-by-hop 头及 **`Accept-Encoding`**，且 **`Host`** 语义与 Chat 一致。请求体 MUST 为 JSON 对象且含非空字符串 **`prompt`**（键须存在、值不可为 **`null`**、去首尾空白后非空）。若 JSON 未提供非空 **`model`**（缺键、**`null`** 或空字符串），MUST 在转发前合并默认 **`model`** 为 **`dall-e-2`**。

#### Scenario: POST /v1/images/generations 鉴权后转发

- **WHEN** 客户端 **`POST /v1/images/generations`** 带有效网关凭证、合法 JSON（含 **`prompt`**，**`model`** 可省略）、且上游已配置
- **THEN** 上游收到与合并规则一致后的请求，且 HTTP 状态码为上游结果或网关 **`BAD_GATEWAY`** 等已文档化错误

#### Scenario: 省略 model 时合并默认图像模型名

- **WHEN** 客户端 **`POST /v1/images/generations`** 的 JSON 含非空 **`prompt`** 且省略 **`model`** 或 **`model`** 为空字符串
- **THEN** 转发前 body MUST 含非空 **`model`**，且为 **`dall-e-2`**

### Requirement: OpenAI 兼容语音合成（JSON）反向代理

网关 MUST 注册 **`POST /v1/audio/speech`**；对 **`GET/PUT/PATCH/DELETE/HEAD`** 等非 **POST** 方法（**`OPTIONS`** 除外）MUST 返回 **405**，body 为网关统一 JSON（**`code`** 为 **`METHOD_NOT_ALLOWED`**）。**`OPTIONS`** 对该路径 MUST 返回 **204 No Content**（与 **`/v1/images/generations`** 预检占位语义一致）。在通过 **`GatewayAuth`** 与 **`KeyRateLimit`** 后，网关 MUST 将请求反向代理至与图像生成代理一致的上游选择结果（模型库启用时按逻辑 **`model`** 选实例并改写 **`model`**；否则按运行时快照 **`Picker`** 与绑定改写）；出站 MUST 剔除与 Chat 代理相同的 hop-by-hop 头及 **`Accept-Encoding`**，且 **`Host`** 语义与 Chat 一致。请求体 MUST 为 JSON 对象且含非空字符串 **`input`** 与非空字符串 **`voice`**（两键均须存在、值不可为 **`null`**、去首尾空白后非空）。若 JSON 未提供非空 **`model`**（缺键、**`null`** 或空字符串），MUST 在转发前合并默认 **`model`** 为 **`tts-1`**。

#### Scenario: POST /v1/audio/speech 鉴权后转发

- **WHEN** 客户端 **`POST /v1/audio/speech`** 带有效网关凭证、合法 JSON（含 **`input`** 与 **`voice`**，**`model`** 可省略）、且上游已配置
- **THEN** 上游收到与合并规则一致后的请求，且 HTTP 状态码为上游结果或网关 **`BAD_GATEWAY`** 等已文档化错误

#### Scenario: 省略 model 时合并默认语音合成模型名

- **WHEN** 客户端 **`POST /v1/audio/speech`** 的 JSON 含非空 **`input`**、**`voice`** 且省略 **`model`** 或 **`model`** 为空字符串
- **THEN** 转发前 body MUST 含非空 **`model`**，且为 **`tts-1`**

### Requirement: 未注册 OpenAI 兼容 v1 路径的 NoRoute 404 形状

当请求未匹配任何已注册路由且 **`NoRoute`** 处理生效时，若 **`URL.Path`** 为 **`/v1`** 或以 **`/v1/`** 为前缀，网关 MUST 返回 **HTTP 404**，且响应 body MUST 为 JSON 对象，其顶层含 **`error`** 字段；**`error`** MUST 为对象，且至少包含 **`message`**（字符串，含方法与路径说明）、**`type`**（字符串，值为 **`invalid_request_error`**）、**`param`** 与 **`code`**（字符串，允许为空）。MUST 设置响应头 **`X-Request-ID`**：若 **`RequestID`** 中间件已写入 Gin 上下文或请求头，MUST 与该值一致；否则 MUST 生成并写入。本规则 MUST NOT 改变已注册 **`/v1/...`** 路由（含 **`POST /v1/chat/completions`**、**`POST /v1/embeddings`**、**`POST /v1/engines/:model/embeddings`**、**`POST /v1/moderations`**、**`POST /v1/images/generations`**、**`POST /v1/audio/speech`**、**`GET /v1/models`**、显式 **501** 表等）的既有语义。其它未匹配路径仍 MUST 使用网关统一 **`code`/`message`/`request_id`** 的 **404** 约定。

#### Scenario: 未注册 v1 子路径返回 invalid_request_error

- **WHEN** 客户端请求 **`GET /v1/no-such-endpoint`**（或任意未注册且路径以 **`/v1/`** 开头的 URI）
- **THEN** HTTP 状态码为 **404**，JSON 顶层存在 **`error.type`** 为 **`invalid_request_error`**，且响应带非空 **`X-Request-ID`**

#### Scenario: 非 v1 未知路径仍为网关统一 404

- **WHEN** 客户端请求 **`GET /unknown-route-xyz`**
- **THEN** HTTP 状态码为 **404**，JSON **`code`** 为 **`NOT_FOUND`**（与既有网关错误体一致）

### Requirement: 请求 ID 中间件全局启用

HTTP 引擎 MUST 在业务路由之前注册 **`RequestID`**（或等价命名）中间件。确定当前请求 ID 时：对 **`X-Request-ID`** 与备用入站头 **`X-Oneapi-Request-Id`** 的值 MUST 先去除首尾空白；若 **`X-Request-ID`** 非空，MUST 使用该值；否则若 **`X-Oneapi-Request-Id`** 非空，MUST 使用该值；否则 MUST 生成新 ID。MUST 将最终请求 ID 写入响应头 **`X-Request-ID`**，并 MUST 将**同一字符串**写入响应头 **`X-Oneapi-Request-Id`**；且 MUST 将当前请求 ID 存入 Gin 上下文（及 **`c.Request.Context()`**）供后续处理器与错误封装读取。

#### Scenario: 下游处理器可读 request_id

- **WHEN** 任意已注册业务处理器执行
- **THEN** 其可通过上下文读取与 **`X-Request-ID`** 一致的请求 ID 字符串

#### Scenario: 标准库 Context 携带请求 ID

- **WHEN** **`RequestID`** 中间件已执行且当前请求 ID 已确定（透传或生成）
- **THEN** **`c.Request.Context()`** 经网关文档化 API 解析得到的请求 ID 字符串 MUST 与 **`X-Request-ID`** 响应头及 Gin 上下文中存储的值一致，且以该 **`Context`** 为父节点的 **`context.WithTimeout` / `WithCancel` 等子上下文** MUST 仍能通过同一 API 读到相同请求 ID

#### Scenario: 双响应头同值

- **WHEN** 任意请求完成且 **`RequestID`** 中间件已确定最终请求 ID
- **THEN** 响应头 **`X-Request-ID`** 与 **`X-Oneapi-Request-Id`** MUST 均为该最终值（二者相等）

#### Scenario: 仅带备用入站头时透传

- **WHEN** 请求未带非空 **`X-Request-ID`**，但带非空 **`X-Oneapi-Request-Id: alt-1`**（经首尾空白去除后非空）
- **THEN** 最终请求 ID 为 **`alt-1`**，响应 **`X-Request-ID`** 与 **`X-Oneapi-Request-Id`** 均为 **`alt-1`**，且 Gin 上下文与 **`c.Request.Context()`** 解析结果均为 **`alt-1`**

#### Scenario: 主入站头优先于备用入站头

- **WHEN** 请求同时带 **`X-Request-ID: main-1`** 与 **`X-Oneapi-Request-Id: alt-2`**（二者经去除首尾空白后均非空）
- **THEN** 最终请求 ID 为 **`main-1`**，响应双头均为 **`main-1`**

### Requirement: Accept-Language 与请求语言标签

HTTP 引擎 MUST 在 **`RequestID`** 之后、**`ZapHTTPAccessLog`**（或等价命名）之前注册 **`AcceptLanguage`**（或等价命名）中间件：读取请求头 **`Accept-Language`**，将语言偏好归约为稳定标签 **`zh-CN`** 或 **`en`**（判定规则：头值为空时视为 **`en`**；否则对整段头值做不区分大小写判断，若以 **`zh`** 为前缀则标签为 **`zh-CN`**，否则为 **`en`**）。MUST 将归约结果存入 Gin 上下文（键 **`locale`**）并写入 **`c.Request.Context()`**，使处理器与 **`WithTimeout` 等子上下文** 能通过 **`locale.FromContext`** 读取与 Gin 中一致的语言标签。

#### Scenario: 中文首选

- **WHEN** 请求携带 **`Accept-Language`** 且其值不区分大小写以 **`zh`** 开头（如 **`zh-CN`** 或 **`zh-TW,en;q=0.9`**）
- **THEN** 归约标签为 **`zh-CN`**，且 Gin 与 **`c.Request.Context()`** 解析结果一致

#### Scenario: 非中文或未携带头

- **WHEN** 请求未带 **`Accept-Language`** 或头值不以 **`zh`** 为前缀
- **THEN** 归约标签为 **`en`**

### Requirement: 引擎级 HTTP 访问日志中间件

HTTP 引擎 MUST 在 **`AcceptLanguage`** 之后、**`GzipRequestDecode`** 之前注册 **`ZapHTTPAccessLog`**（或等价命名）中间件：在 **`c.Next()`** 返回后 MUST 向主 Zap 日志写入**一条** **`Info`** 级别记录，消息键为 **`http_access`**（或文档化之稳定枚举），且 MUST 包含结构化字段：**`request_id`**（与 **`X-Request-ID`** / Gin 上下文一致）、**`status`**（最终 HTTP 状态码；若尚未写入则视为 **200**）、**`method`**、**`path`**（优先使用路由模板路径如 Gin **`FullPath()`**，若为空则使用 **`URL.Path`**）、**`client_ip`**、**`latency_ms`**（非负整数，自进入该中间件至 **`Next` 返回**的耗时）。MUST NOT 在该条日志中记录 query 字符串、请求体或 **`Authorization`** 原文。

#### Scenario: 成功请求产生访问日志

- **WHEN** 客户端 **`GET /health`**（或任意已注册成功响应路由）且 **`RequestID`** 已执行
- **THEN** Zap 至少有一条 **`http_access`** 记录，其 **`request_id`** 与响应头 **`X-Request-ID`** 一致，**`status`** 与响应状态码一致，**`latency_ms`** 存在

### Requirement: Gzip 请求体透明解码

HTTP 引擎 MUST 在 **`ZapHTTPAccessLog`** 之后、**`Recovery`** 之前注册 **`GzipRequestDecode`**（或等价命名）中间件：当请求头 **`Content-Encoding`**（大小写不敏感、忽略首尾空白）为 **`gzip`** 时，MUST 将 **`c.Request.Body`** 解压为**未压缩字节流**并替换 **`Content-Length`** 为解压后长度，且 MUST 从请求头中移除 **`Content-Encoding`**，使后续处理器与上游转发与客户端直接发送未压缩 body 的语义一致。解压失败（含非法 gzip 流）时 MUST 返回 **400** 且 body 符合既有网关统一 JSON 错误约定（含 **`request_id`**）。

#### Scenario: 未声明 gzip 时原样通过

- **WHEN** 请求未带 **`Content-Encoding: gzip`**
- **THEN** 中间件不改变 body 与相关头（除后续处理器正常消费外）

#### Scenario: 声明 gzip 且流合法

- **WHEN** 请求带 **`Content-Encoding: gzip`** 且 body 为合法 gzip 压缩内容
- **THEN** 后续处理器 **`io.ReadAll(c.Request.Body)`** 得到解压后字节序列；请求上不再携带 **`Content-Encoding: gzip`**；**`Content-Length`** 与解压后长度一致

### Requirement: 公开上传路径成功响应的 Cache-Control

HTTP 引擎 MUST 在 **`ErrorJSON`**（或等价命名）之后注册 **`UploadsStaticCache`**（或等价命名）中间件：在 **`c.Next()`** 返回后，若请求方法为 **`GET`** 或 **`HEAD`**、路径以 **`/uploads/`** 为前缀、响应 **`Cache-Control`** 尚未为非空白字符串、且 HTTP 状态码为 **2xx**（含 Gin 在尚未 **`WriteHeader`** 时视为 **200** 的约定）或 **`304 Not Modified`**，则 MUST 设置 **`Cache-Control: public, max-age=604800`**。若状态码为 **4xx**、**5xx** 或其它不满足前述条件者，MUST NOT 因本规则写入该头。MUST NOT 覆盖已由处理器或静态文件处理链写入的非空 **`Cache-Control`**。

#### Scenario: 成功返回上传目录下文件时带长期缓存提示

- **WHEN** 客户端 **`GET /uploads/…`**（路径至少包含 **`/uploads/`** 前缀加一级以上路径段）且处理器返回 **200** 且未设置 **`Cache-Control`**
- **THEN** 响应 MUST 包含 **`Cache-Control: public, max-age=604800`**

#### Scenario: 未命中文件时不强缓存

- **WHEN** 客户端 **`GET /uploads/…`** 且处理器返回 **404**
- **THEN** 响应 MUST NOT 因本中间件出现 **`public, max-age=604800`**

### Requirement: 根路径严格 URI 的 Cache-Control

HTTP 引擎 MUST 在 **`ErrorJSON`**（或等价命名）之后、**`UploadsStaticCache`**（或等价命名）之前注册 **`RootStrictNoCache`**（或等价命名）中间件：当 **`net/http.Request.RequestURI`**（或 Gin 上下文中与之等价的原始请求目标字符串）**严格等于** **`/`**（不含 **`?`** 查询串、不含额外路径段）时，在调用 **`c.Next()`** 之前 MUST 向响应头写入 **`Cache-Control: no-cache`**（或与之等价的「禁止将响应当作可长期复用缓存」语义）。当 **`RequestURI`** 不为 **`/`**（例如 **`/?q=1`** 或 **`/foo`**）时，MUST NOT 仅因本规则写入该头。若后续处理器或 **`NoRoute`** 另行设置非空 **`Cache-Control`**，最终响应以写入顺序后者为准（实现 MAY 允许后续覆盖）。

#### Scenario: 仅根路径无查询时声明不可缓存

- **WHEN** 客户端请求 **`GET /`** 且原始 **`RequestURI`** 为 **`/`**（无 **`?`** 查询）
- **THEN** 响应 MUST 包含 **`Cache-Control: no-cache`**

#### Scenario: 根路径带查询时不套用本规则

- **WHEN** 客户端请求 **`GET /?x=1`**
- **THEN** 响应 MUST NOT 仅因本规则出现 **`Cache-Control: no-cache`**（除非其它独立规则写入）

### Requirement: 管理端 HTTP 响应可选 Gzip

在管理控制台 API 前缀（如 **`/api/admin/v1`**）下，MUST 注册 **`github.com/gin-contrib/gzip`**（或等价实现）于该前缀对应的路由组：**仅**作用于管理端路径，MUST NOT 作为引擎级中间件挂到 **`POST /v1/chat/completions`** 或其它流式代理路由。当客户端请求头 **`Accept-Encoding`** 表明接受 **`gzip`**（大小写不敏感、含 **`gzip`** 子串即可）且中间件判定可压缩时，响应 MAY 使用 **`Content-Encoding: gzip`** 压缩 body；当客户端未表明接受 **`gzip`** 时，MUST NOT 仅因本规则增设 **`Content-Encoding: gzip`**。

#### Scenario: 接受 gzip 时可能返回压缩体

- **WHEN** 客户端请求某管理端 JSON 端点且 **`Accept-Encoding`** 包含 **`gzip`**
- **THEN** 响应 MAY 带 **`Content-Encoding: gzip`**，且解压后 body 与未压缩语义一致

#### Scenario: 未接受 gzip 时不强制压缩

- **WHEN** 客户端未在 **`Accept-Encoding`** 中表明接受 **`gzip`**
- **THEN** 响应 MUST NOT 带 **`Content-Encoding: gzip`**（除非其它独立规则设置）

### Requirement: 请求体可复用读取辅助

`services/gateway` MUST 在 **`internal/bodyreuse`**（或文档化之等价包）提供：**`GetRequestBody`**（首次从 **`c.Request.Body`** 读取并缓存在 Gin 上下文，再次调用返回同一份字节）、**`ResetRequestBody`**（将字节写回 **`c.Request.Body`** 并同步缓存与 **`Content-Length`**）、**`UnmarshalBodyReusable`**（当 **`Content-Type`** 为 **`application/json`** 前缀时对目标结构体做 JSON 反序列化，否则走 Gin **`ShouldBind`**，成功后还原 body 供后续读取）。**`POST /v1/chat/completions`** 转发路径 MUST 使用上述辅助之一保证改写 body 后仍与缓存一致（以实现为准）。

#### Scenario: 多次获取同一 body

- **WHEN** 同一请求处理链内两次调用 **`GetRequestBody`**
- **THEN** 两次返回的字节序列相同，且不因首次读取导致第二次 **`io.ReadAll`** 失败（若实现依赖缓存）

### Requirement: SQLite 文件模式连接忙碌等待

当 **`NEXUSROUTER_DATABASE_URL`** 为空、网关以 **SQLite 文件**（含默认 **`services/gateway/gateway.db`** 或 **`NEXUSROUTER_SQLITE_PATH`**）打开数据库时，**`internal/repository.OpenDB`**（或等价实现）MUST 在传给驱动的连接串上附带 URI 参数 **`_busy_timeout`**，其值为**毫秒**。**`config.Load`** MUST 读取整数环境变量 **`NEXUSROUTER_SQLITE_BUSY_TIMEOUT_MS`** 并写入配置字段（未设置或解析失败时视为 **`0`**）。当该字段 **`<= 0`** 时，**`_busy_timeout`** MUST 取 **`3000`**；当大于 **`600000`** 时 MUST 钳制为 **`600000`**。若路径段已含 **`?`**（如内存 DSN），MUST 使用 **`&_busy_timeout=…`** 追加。使用 **Postgres**（**`NEXUSROUTER_DATABASE_URL`** 非空）时 MUST NOT 依赖本变量改变连接串。

#### Scenario: 未设置环境变量时的默认忙碌等待

- **WHEN** **`NEXUSROUTER_DATABASE_URL`** 为空且未设置 **`NEXUSROUTER_SQLITE_BUSY_TIMEOUT_MS`**（或值为非正整数）
- **THEN** 实际打开 SQLite 使用的 DSN MUST 含 **`_busy_timeout=3000`**

#### Scenario: 显式配置忙碌等待

- **WHEN** **`NEXUSROUTER_SQLITE_BUSY_TIMEOUT_MS`** 为 **`8000`** 且走 SQLite 文件
- **THEN** DSN MUST 含 **`_busy_timeout=8000`**

### Requirement: GORM 预编译语句（PrepareStmt）

**`internal/repository.OpenDB`** 在通过 **`gorm.Open`** 打开 **Postgres**（**`NEXUSROUTER_DATABASE_URL`** 非空）或 **SQLite** 时，MUST 使用将 **`PrepareStmt`** 设为 **`true`** 的 **`gorm.Config`**（与静默日志等既有字段一并构造），以在进程内缓存预编译 SQL、减少重复解析开销。本要求仅约束网关生产 **`OpenDB`** 路径；测试辅助代码使用内存 DSN 或独立 **`gorm.Config`** 时 MAY 不显式开启。

#### Scenario: 网关 GORM 配置含预编译

- **WHEN** 审查 **`OpenDB`** 传入 **`gorm.Open`** 的 **`*gorm.Config`** 构造方式
- **THEN** **`PrepareStmt`** MUST 为 **`true`**

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

### Requirement: 可选外置前端基址与未匹配路由重定向

当环境变量 **`NEXUSROUTER_FRONTEND_BASE_URL`** 设置为合法的 **`http`** 或 **`https`** 绝对 URL（含主机名，可含路径前缀；首尾空白去除，末尾 **`/`** 去除）时，网关对**未匹配任何已注册路由**的请求 MUST 响应 **`301 Moved Permanently`**，**`Location`** 为此前缀与请求**原始 URI**（路径与 query，与 **`Request-URI`** 一致）的字符串拼接。当该变量未设置、为空或无法解析为上述合法 URL 时，未匹配路由的行为 MUST 与既有约定一致（如返回 JSON **`NOT_FOUND`**），且 MUST NOT 因非法值导致进程启动失败。

#### Scenario: 配置有效时未匹配路径重定向

- **WHEN** **`NEXUSROUTER_FRONTEND_BASE_URL`** 为 **`https://ui.example`** 且客户端请求 **`GET /foo/bar?q=1`**，且该路径无对应路由
- **THEN** 响应状态码为 **301**，**`Location`** 为 **`https://ui.example/foo/bar?q=1`**

#### Scenario: 非法基址不启用重定向

- **WHEN** **`NEXUSROUTER_FRONTEND_BASE_URL`** 为非 **`http`/`https`** 方案或缺少主机
- **THEN** 未匹配路由时仍返回既有 JSON 错误，且进程正常完成配置加载

