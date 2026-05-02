## ADDED Requirements

### Requirement: 进阶管理页面与导航

`web/dashboard` MUST 在已登录管理布局中提供入口与路由页面：**日志查询**（筛选表单、表格、导出按钮）、**限流规则**（表格与编辑抽屉/弹窗）、**跨域配置**（含批量域名输入）、**IP 黑白名单**（类型切换、批量粘贴、列表展示）。文案与注释 MUST 使用**中文**（与现有仪表盘规范一致）。

#### Scenario: 各页在无后端时优雅降级

- **WHEN** 管理 API 返回错误或网络失败
- **THEN** 页面 MUST 展示 Ant Design `Result`/`Alert` 或等价错误提示，且 MUST NOT 白屏崩溃

### Requirement: 导出与批量交互

日志页 MUST 提供 **CSV 导出**触发控件；CORS 与 IP 名单页 MUST 提供**批量输入**控件（多行文本）。操作成功或失败后 MUST 有明确 `message` 反馈。

#### Scenario: 批量域名为空提交被拦截

- **WHEN** 用户未输入任何有效域名即提交批量添加
- **THEN** 前端 MUST 提示校验错误且不发送空污染请求（或后端拒绝二者择一，须在实现中一致）
