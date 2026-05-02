## MODIFIED Requirements

### Requirement: 文档与提交语言

仪表盘源码中的 **注释**与 **Git commit message** MUST 使用 **中文**；变量与函数名 MUST 使用 **英文**。**用户可见界面文案** MUST 通过国际化资源提供 **中文（默认）** 与 **英文** 两种语言，且 MUST 提供语言切换入口；键名与资源文件组织 MUST 符合 `admin-i18n` 能力规范。

#### Scenario: 代码审查抽样

- **WHEN** 审查者打开主要入口与配置文件的注释
- **THEN** 注释为中文且标识符为英文

#### Scenario: 默认语言为中文

- **WHEN** 用户首次打开仪表盘且无持久化语言偏好
- **THEN** 用户可见文案为中文

#### Scenario: 切换为英文

- **WHEN** 用户选择 English
- **THEN** 用户可见文案切换为英文

## ADDED Requirements

### Requirement: 系统设置页

仪表盘 MUST 提供 **系统设置** 路由与表单（或分组只读卡片），展示 `admin-system-settings` 规范中的字段及 **mutability** 提示；管理员保存后 MUST 根据响应展示「已热更新」或「需重启进程」说明。

#### Scenario: 操作员不可见保存控件

- **WHEN** 当前用户角色为 `operator`
- **THEN** 系统设置页不展示写提交或等价保存入口

### Requirement: 运行告警展示

仪表盘 MUST 在管理布局内消费告警状态 API，并按 `admin-runtime-alerts` 规范展示告警条。

#### Scenario: 告警与 i18n

- **WHEN** 告警条展示摘要编码
- **THEN** 文案经 i18n 解析为中英之一，与当前语言一致
