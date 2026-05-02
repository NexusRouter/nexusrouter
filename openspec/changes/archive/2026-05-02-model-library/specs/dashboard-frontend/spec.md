# dashboard-frontend Specification（变更 `model-library` — 增量）

## ADDED Requirements

### Requirement: 模型库页面与导航

`web/dashboard` SHALL 提供 **`/model-library`**（或实现选定路径）路由页面，并在桌面侧栏导航中增加一项（图标 + 文案 i18n）；页面 MUST 使用 **Ant Design** 与 **TanStack Query** 与现有管理页一致。

#### Scenario: 导航可达

- **WHEN** 已登录管理员打开侧栏并点击模型库
- **THEN** 路由切换至模型库页面且无控制台报错

### Requirement: 国际化

模型库页面所有用户可见文案 MUST 具备 **中文与英文** 键（`pages.modelLibrary.*` 或等价前缀），并符合 **`dashboard-frontend`** 注释语言约定。

#### Scenario: 切换语言

- **WHEN** 用户在顶栏切换语言
- **THEN** 模型库标题与表格列头随之切换

### Requirement: 可访问性

主要操作按钮（查询、保存、同步）MUST 具备可聚焦与 **`aria-label` 或可见文本**；表格 MUST 提供合理 **`rowKey`**。

#### Scenario: 键盘可达

- **WHEN** 使用 Tab 聚焦主要按钮
- **THEN** 可触发且焦点环可见（浏览器默认或主题）
