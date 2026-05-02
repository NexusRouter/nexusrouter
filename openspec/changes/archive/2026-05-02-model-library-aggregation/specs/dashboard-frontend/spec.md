## MODIFIED Requirements

### Requirement: 模型库页面与导航

`web/dashboard` SHALL 提供 **`/model-library`**（或实现选定路径）路由页面，并在桌面侧栏导航中增加一项（图标 + 文案 i18n）；页面 MUST 使用 **Ant Design** 与 **TanStack Query** 与现有管理页一致。页面 MUST 支持在 **厂商**、**逻辑模型（model_code）**、**实例** 维度浏览与筛选，并提供 **实例启用/禁用**、**创建/编辑** 入口（与后端聚合 API 对齐）；MAY 使用子路由或标签页拆分 **厂商 / 逻辑模型 / 实例** 视图。

#### Scenario: 导航可达

- **WHEN** 已登录管理员打开侧栏并点击模型库
- **THEN** 路由切换至模型库页面且无控制台报错
