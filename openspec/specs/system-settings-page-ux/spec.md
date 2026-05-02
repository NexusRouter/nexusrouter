# system-settings-page-ux Specification

## Purpose
TBD - created by archiving change system-settings-page-ux-refactor. Update Purpose after archive.
## Requirements
### Requirement: 系统设置页场景化信息架构

`web/dashboard` 的系统设置路由页面 MUST 将内容组织为三个用户可理解的模块：**系统运行状态**（只读）、**代理访问日志**（可编辑）、**高级配置**（可选折叠的技术细节）。模块顺序 MUST 为运行状态在上、日志配置在中、高级配置在下（或等价视觉层级）。用户 MUST 无需依赖后端键名即可识别各模块用途。

#### Scenario: 管理员打开系统设置

- **WHEN** 已登录管理员打开系统设置页
- **THEN** 页面同时呈现「系统运行状态」与「代理访问日志」两个主区块，且「高级配置」以折叠或明确入口展示并可展开

#### Scenario: 操作员打开系统设置

- **WHEN** 已登录操作员打开系统设置页
- **THEN** 其可见与管理员相同的只读与表单布局，但保存类控件处于禁用态且存在只读提示（与 `admin-rbac` 一致）

### Requirement: 术语与国际化一致性

界面主标签、模块标题、生效方式说明 MUST 使用业务向中文/英文文案；后端字段名（如 `proxy_access_log_enabled`）MUST NOT 作为主标题或唯一识别方式，MAY 出现在高级配置或次要位置。中文与英文资源 MUST 描述同一含义，不得出现同一键在中英文下指向不同操作路径的表述。

#### Scenario: 主文案不暴露裸键名

- **WHEN** 用户查看「系统运行状态」中任一项
- **THEN** 首要可见文本为用户术语（如监听地址、超时、日志是否启用），而非未翻译的 snake_case 键名

#### Scenario: 语言切换

- **WHEN** 用户在中文与英文之间切换
- **THEN** 系统设置页全部用户可见文案随语言切换，且模块结构与交互可用性保持不变

### Requirement: 运行状态与生效方式透明

对 `GET /api/admin/v1/system/settings` 返回的每一项，页面 MUST 展示：**当前值**、**生效方式**（与 `mutability` 一致：热更新 / 需重启进程 / 只读展示）、以及**操作或风险提示**（合并服务端 `hint` 与产品化短说明）。对 `restart_required` 与 `read_only` MUST 明确告知用户本页不能直接在线修改生效值（若适用）。

#### Scenario: 需重启项展示

- **WHEN** 某字段的 `mutability` 为 `restart_required`
- **THEN** 界面说明须通过环境变量等方式修改并重启进程，且不与「保存即可热生效」混淆

#### Scenario: 热更新项展示

- **WHEN** 某字段的 `mutability` 为 `hot_reload`
- **THEN** 界面说明与「代理访问日志」保存路径一致（可写回并 Reload），且保存成功后只读区展示与之一致

### Requirement: 代理访问日志表单与状态联动

代理访问日志表单 MUST 使用现有 `PUT /api/admin/v1/system/settings` 请求体（`proxy_access_log`、`persist`），不得引入新的请求字段。保存成功 MUST 触发对系统设置只读数据的重新获取，使「系统运行状态」中与日志相关的展示与表单提交结果一致。

#### Scenario: 保存后刷新

- **WHEN** 管理员提交日志表单且 API 返回成功
- **THEN** 客户端重新请求系统设置并更新运行状态区中与日志相关的显示（启用状态、路径、级别等）

### Requirement: 风险提示与写回说明

页面 MUST 对「写回 gateway.yaml」或等价的持久化选项提供简短说明，说明其与运行时生效的关系及失败时的用户可见反馈（沿用现有错误提示机制）。对可能导致服务异常的配置修改（如无效路径、不可写目录）MUST 在文案或校验提示中引导用户检查权限与路径（在不新增后端的前提下优先使用前端校验与服务端错误映射）。

#### Scenario: 持久化开关可见

- **WHEN** 用户查看代理访问日志表单
- **THEN** 存在明确的「写回配置文件」或等价开关及一句辅助说明

### Requirement: 兼容性与非回归

本页重构 MUST NOT 改变对现有管理 API 的调用路径、HTTP 方法与成功/失败语义。MUST 继续兼容当前 `gateway.yaml` 与运行时热更新行为；不得要求用户升级网关后端即可使用新界面。

#### Scenario: API 未变更

- **WHEN** 开发者比对变更前后的网络请求
- **THEN** `GET`/`PUT` 的 URL 与 JSON 形状与重构前一致（除用户输入值变化外）

