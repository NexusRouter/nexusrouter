# admin-i18n Specification

## Purpose
TBD - created by archiving change admin-auxiliary-i18n-rbac-alerts. Update Purpose after archive.
## Requirements
### Requirement: 仪表盘语言资源与默认语言

系统 MUST 为管理仪表盘提供 **中文（zh）** 与 **英文（en）** 两套用户可见文案资源；**默认语言 MUST 为中文**。所有管理页标题、菜单、按钮、表单标签、表格列头、空态与 `message` / `notification` 提示 MUST 通过国际化键解析，MUST NOT 在 JSX 中保留仅单一自然语言的终态用户文案（开发期占位除外）。

#### Scenario: 首次访问为中文

- **WHEN** 用户首次打开仪表盘根路径且浏览器未持久化语言偏好
- **THEN** 界面显示中文文案

#### Scenario: 切换为英文

- **WHEN** 用户在设置或顶栏选择 English
- **THEN** 上述文案切换为英文且 Ant Design 组件语言与日期类展示与所选语言一致

### Requirement: 语言偏好持久化

系统 MUST 将用户所选语言持久化到 **浏览器本地存储**（或等价机制），以便刷新后保持；MUST NOT 将语言偏好写入需服务端信任的安全敏感 Cookie 除非同时满足 HTTPS 与团队安全基线（首版推荐 localStorage）。

#### Scenario: 刷新后保持语言

- **WHEN** 用户已选择英文并刷新页面
- **THEN** 界面仍为英文

