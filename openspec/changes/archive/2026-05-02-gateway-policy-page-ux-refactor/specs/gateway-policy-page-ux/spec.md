# gateway-policy-page-ux Specification

## Purpose

定义 NexusRouter 管理控制台中**网关策略**相关页面的信息架构、业务化文案、交互一致性与分层展示要求。运行时行为仍以 `admin-ip-access-control`、`admin-rate-limit-policy`、`admin-cors-management` 及网关后端规格为准；本规格仅约束**前端呈现与操作体验**。

## ADDED Requirements

### Requirement: 场景化信息架构

管理控制台 SHALL 在网关策略相关界面以**三大场景**组织内容：**IP 访问控制**、**限流规则**、**跨域配置**（中英文标题与副文案 SHALL 体现用户任务而非后端模块内部名称）。三场景 SHALL 在同一逻辑页面内连续呈现（单页滚动或等价体验），并 MAY 通过 URL 片段或查询参数支持定位到某一场景以便深链。

#### Scenario: 用户按任务找到区块

- **WHEN** 已登录管理员打开网关策略页
- **THEN**  SHALL 在同一视口内依次（或经锚点）可发现上述三个场景区块，且 SHALL NOT 仅依赖「限流 / CORS / IP 名单」等纯技术标签作为唯一导航语义

#### Scenario: 旧书签仍可达

- **WHEN** 用户访问此前保存的 `/gateway/rate-limits`、`/gateway/cors` 或 `/gateway/ip-access` 路径
- **THEN** 应用 SHALL 重定向或导航至新网关策略页并定位到对应场景（或保持兼容子路由但布局与单页一致）

### Requirement: 术语与说明文案

所有用户可见选项 SHALL 使用**业务语言**作为主标签；若需暴露 API/YAML 字段名或枚举值，SHALL 以次要说明、Tooltip 或折叠帮助呈现。`allowlist` / `denylist`、`api_key_fp`、HTTP 方法名（如 PATCH）等 SHALL NOT 单独作为唯一标签而无解释。

#### Scenario: 限流维度可读

- **WHEN** 用户配置限流规则并查看维度选项
- **THEN** 每个选项 SHALL 展示业务描述（例如按 IP、按已识别 API 访问身份），并 MAY 在辅助位置标明与 `ip` / `api_key_fp` 的对应关系

#### Scenario: IP 模式可读

- **WHEN** 用户选择 IP 访问模式
- **THEN** SHALL 展示该模式对请求的实际效果（允许名单内 / 拦截名单内 / 不启用名单）

### Requirement: 交互一致性

IP 访问、限流规则、跨域配置三个场景在**保存、新增/删除列表项、加载中、只读禁用、成功/失败反馈**上 SHALL 遵循同一套交互模式：主保存操作位置与视觉层级一致；失败时 SHALL 明确提示且 SHALL 不暗示已成功持久化；与 `gateway.yaml` 写回相关的选项 SHALL 在三处均有等价的可见控件与说明（除非该 API 不支持 persist，则 SHALL 在界面注明并隐藏无效开关）。

#### Scenario: Operator 只读一致

- **WHEN** 当前角色为 Operator（或等价只读角色）
- **THEN** 三个场景的编辑控件 SHALL 均被禁用，且行为一致

#### Scenario: 保存失败不误导

- **WHEN** 任一区块的保存 API 返回错误
- **THEN** 用户 SHALL 收到错误反馈，且界面状态 SHALL 反映未成功提交（例如保留编辑值，不显示成功提示）

### Requirement: CORS 分层展示

跨域配置 SHALL 区分**基础配置**与**高级配置**：基础配置 SHALL 包含启用开关、允许来源（含批量导入）及与大多数场景相关的预检缓存等；高级配置 SHALL 包含 HTTP 方法列表、允许请求头及其他低频字段。高级配置 SHALL 默认折叠或隐藏在「高级」展开区域内，展开后 SHALL 保持可编辑与校验。

#### Scenario: 默认简化视图

- **WHEN** 用户首次展开跨域配置区块且未主动打开高级区域
- **THEN** SHALL NOT 要求用户在同一屏内填写方法与请求头即可理解跨域是否开启及允许哪些浏览器来源

### Requirement: 校验与重复规则提示

前端 SHALL 在用户提交前对明显非法输入进行校验或前置提示，包括不限于：IP/CIDR 格式、路径前缀格式、正数 RPS/Burst、合理的 `max_age_seconds` 范围。对**可能重复或冲突的规则**（例如相同路径前缀与维度下优先级冲突）SHALL 给出警告或阻止提交（具体策略以实现为准，但 SHALL 与后端校验结果不矛盾）。

#### Scenario: 非法 CIDR 不可静默成功

- **WHEN** 用户在 IP 名单输入非法条目并尝试保存
- **THEN** SHALL 阻止或标记错误，且 SHALL NOT 在无提示情况下静默丢弃

### Requirement: 风险与持久化说明

对「写回 `gateway.yaml`」或等价持久化开关，界面 SHALL 提供简短说明：说明写入的是配置文件、与仅内存热更新的区别，以及失败时保留旧配置的行为（与后端实际一致）。在切换 IP 名单模式等高风险操作处，SHALL 提供额外确认或醒目提示。

#### Scenario: 写回开关有据可依

- **WHEN** 用户将「写回配置文件」从关闭切换为开启
- **THEN** SHALL 可见辅助文案说明该操作的影响范围

### Requirement: 用户角色与使用场景说明

规格文档或产品附录 SHALL 描述至少两类典型场景：**日常运维**（调整限流与跨域、临时加白名单）与**安全加固**（启用白名单、收紧 CORS）。界面文案 MAY 引用此类场景作为侧栏帮助或文档链接，但 SHALL NOT 替代角色权限控制（RBAC 仍以既有 `admin-rbac` 为准）。

#### Scenario: 文档化角色期望

- **WHEN** 审查者阅读本变更的 specs 与设计说明
- **THEN** SHALL 能找到对典型使用场景的叙述，用于验收测试用例编写
