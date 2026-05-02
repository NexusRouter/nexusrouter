# dashboard-frontend Specification

## Purpose

描述仪表盘前端在已登录管理布局中的导航与体验实现，使其满足全局控制台 UX 规格中的侧栏分组、路由可达性与危险操作/空错载基线。

## Requirements

### Requirement: 管理控制台侧栏域分组

`web/dashboard` MUST 在已登录管理布局侧栏实现**域分组**导航，且 MUST 满足 `admin-console-global-ux` 中「侧栏信息架构与域分组」「导航高亮与路径匹配」的全部场景。各既有路由与「管理控制台路由与导航」所列页面入口 MUST 保持可达。

#### Scenario: 全入口可达

- **WHEN** 已登录用户依次从侧栏访问仪表盘、上游、模型库、API Key、访问日志、网关策略、系统设置
- **THEN** 每一入口均可到达对应页面且无 404（在网关正常运行前提下）

### Requirement: 全局危险操作与空错载体验基线

`web/dashboard` 各管理页 MUST 符合 `admin-console-global-ux` 中「危险操作分级与统一确认」「空态、加载与错误基线」的要求；新增页面 MUST 继承同一基线。

#### Scenario: 删除操作有确认

- **WHEN** 用户在任一管理页触发删除类不可逆操作
- **THEN** 满足全局规格中删除确认场景（见 `admin-console-global-ux`）
